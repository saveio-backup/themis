/*
 * Copyright (C) 2019 The themis Authors
 * This file is part of The themis library.
 *
 * The themis is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The themis is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package proc provides functions for handle messages from
// consensus/ledger/net/http/validators
package proc

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	consutils "github.com/saveio/themis/consensus/utils"
	"github.com/saveio/themis/core/ledger"
	tx "github.com/saveio/themis/core/types"
	"github.com/saveio/themis/errors"
	httpcom "github.com/saveio/themis/http/base/common"
	msgpack "github.com/saveio/themis/p2pserver/message/msg_pack"
	p2p "github.com/saveio/themis/p2pserver/net/protocol"
	params "github.com/saveio/themis/smartcontract/service/native/global_params"
	gov "github.com/saveio/themis/smartcontract/service/native/governance"
	nutils "github.com/saveio/themis/smartcontract/service/native/utils"
	tc "github.com/saveio/themis/txnpool/common"
	"github.com/saveio/themis/validator/types"
)

type txStats struct {
	sync.RWMutex
	count []uint64
}

type serverPendingTx struct {
	tx     *tx.Transaction   // Pending tx
	sender tc.SenderType     // Indicate which sender tx is from
	ch     chan *tc.TxResult // channel to send tx result
}

type pendingBlock struct {
	mu             sync.RWMutex
	sender         *actor.PID                            // Consensus PID
	height         uint32                                // The block height
	processedTxs   map[common.Uint256]*tc.VerifyTxResult // Transaction which has been processed
	unProcessedTxs map[common.Uint256]*tx.Transaction    // Transaction which is not processed
}

type roundRobinState struct {
	state map[types.VerifyType]int // Keep the round robin index for each verify type
}

type registerValidators struct {
	sync.RWMutex
	entries map[types.VerifyType][]*types.RegisterValidator // Registered validator container
	state   roundRobinState                                 // For loadbance
}

type serverPendingParam struct {
	param  *gov.SubmitNonceParam // Pending poc puzzle
	sender tc.SenderType         // Indicate which sender poc param is from
}

type CheckPoCReq struct {
	param *gov.SubmitNonceParam // Pending poc puzzle
	info  *gov.MiningInfo       // mining info
}

// TXPoolServer contains all api to external modules
type TXPoolServer struct {
	mu                    sync.RWMutex                        // Sync mutex
	wg                    sync.WaitGroup                      // Worker sync
	workers               []txPoolWorker                      // Worker pool
	txPool                *tc.TXPool                          // The tx pool that holds the valid transaction
	allPendingTxs         map[common.Uint256]*serverPendingTx // The txs that server is processing
	pendingBlock          *pendingBlock                       // The block that server is processing
	actors                map[tc.ActorType]*actor.PID         // The actors running in the server
	Net                   p2p.P2P
	validators            *registerValidators // The registered validators
	stats                 txStats             // The transaction statstics
	slots                 chan struct{}       // The limited slots for the new transaction
	height                uint32              // The current block height
	gasPrice              uint64              // Gas price to enforce for acceptance into the pool
	disablePreExec        bool                // Disbale PreExecute a transaction
	disableBroadcastNetTx bool                // Disable broadcast tx from network
	pocWg                 sync.WaitGroup      // poc Worker sync
	pocWorkers            []pocPoolWorker     // poc Worker pool
	pocPool               *tc.PoCPool
	allPendingParam       map[common.Uint256]*serverPendingParam // The poc that server is processing
	pocSlots              chan struct{}
	miningInfo            *gov.MiningInfo
}

// NewTxPoolServer creates a new tx pool server to schedule workers to
// handle and filter inbound transactions from the network, http, and consensus.
func NewTxPoolServer(num uint8, disablePreExec, disableBroadcastNetTx bool) *TXPoolServer {
	s := &TXPoolServer{}
	s.init(num, disablePreExec, disableBroadcastNetTx)
	return s
}

// getGlobalGasPrice returns a global gas price
func getGlobalGasPrice() (uint64, error) {
	mutable, err := httpcom.NewNativeInvokeTransaction(0, 0, nutils.ParamContractAddress, 0, "getGlobalParam", []interface{}{[]interface{}{"gasPrice"}})
	if err != nil {
		return 0, fmt.Errorf("NewNativeInvokeTransaction error:%s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return 0, err
	}
	result, err := ledger.DefLedger.PreExecuteContract(tx)
	if err != nil {
		return 0, fmt.Errorf("PreExecuteContract failed %v", err)
	}

	queriedParams := new(params.Params)
	data, err := hex.DecodeString(result.Result.(string))
	if err != nil {
		return 0, fmt.Errorf("decode result error %v", err)
	}

	err = queriedParams.Deserialization(common.NewZeroCopySource([]byte(data)))
	if err != nil {
		return 0, fmt.Errorf("deserialize result error %v", err)
	}
	_, param := queriedParams.GetParam("gasPrice")
	if param.Value == "" {
		return 0, fmt.Errorf("failed to get param for gasPrice")
	}

	gasPrice, err := strconv.ParseUint(param.Value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint %v", err)
	}

	return gasPrice, nil
}

// getGasPriceConfig returns the bigger one between global and cmd configured
func getGasPriceConfig() uint64 {
	globalGasPrice, err := getGlobalGasPrice()
	if err != nil {
		log.Info(err)
		return 0
	}

	if globalGasPrice < config.DefConfig.Common.GasPrice {
		return config.DefConfig.Common.GasPrice
	}
	return globalGasPrice
}

// init initializes the server with the configured settings
func (s *TXPoolServer) init(num uint8, disablePreExec, disableBroadcastNetTx bool) {
	// Initial txnPool
	s.txPool = &tc.TXPool{}
	s.txPool.Init()
	s.allPendingTxs = make(map[common.Uint256]*serverPendingTx)
	s.actors = make(map[tc.ActorType]*actor.PID)

	s.validators = &registerValidators{
		entries: make(map[types.VerifyType][]*types.RegisterValidator),
		state: roundRobinState{
			state: make(map[types.VerifyType]int),
		},
	}

	s.pendingBlock = &pendingBlock{
		processedTxs:   make(map[common.Uint256]*tc.VerifyTxResult, 0),
		unProcessedTxs: make(map[common.Uint256]*tx.Transaction, 0),
	}

	s.stats = txStats{count: make([]uint64, tc.MaxStats-1)}

	s.slots = make(chan struct{}, tc.MAX_LIMITATION)
	for i := 0; i < tc.MAX_LIMITATION; i++ {
		s.slots <- struct{}{}
	}

	s.gasPrice = getGasPriceConfig()
	log.Infof("tx pool: the current local gas price is %d", s.gasPrice)

	s.disablePreExec = disablePreExec
	s.disableBroadcastNetTx = disableBroadcastNetTx
	// Create the given concurrent workers
	s.workers = make([]txPoolWorker, num)
	// Initial and start the workers
	var i uint8
	for i = 0; i < num; i++ {
		s.wg.Add(1)
		s.workers[i].init(i, s)
		go s.workers[i].start()
	}

	// init poc stuff
	miningInfo, err := consutils.GetMiningInfo()
	if err != nil {
		log.Infof("tx pool: fail to get mining info: %s", err)
	}
	s.miningInfo = miningInfo

	// Initial PoCPool
	s.pocPool = &tc.PoCPool{}
	s.pocPool.Init(miningInfo.View)
	s.allPendingParam = make(map[common.Uint256]*serverPendingParam)

	s.pocSlots = make(chan struct{}, num)
	for i = 0; i < num; i++ {
		s.pocSlots <- struct{}{}
	}

	// Create the given concurrent workers for PoC
	s.pocWorkers = make([]pocPoolWorker, num)
	// Initial and start the poc workers
	for i = 0; i < num; i++ {
		s.pocWg.Add(1)
		s.pocWorkers[i].init(i, s)
		go s.pocWorkers[i].start()
	}

	// update miningInfo and notify PoCPool
	go s.monitor()
}

func (s *TXPoolServer) monitor() {
	interval := config.DefConfig.PoC.QueryMiningInfoInterval
	timer := time.NewTimer(time.Second * time.Duration(interval))

	log.Infof("TXPoolServer monitor start with interval:%d", interval)

	for {
		select {
		case <-timer.C:
			timer.Reset(time.Second * time.Duration(interval))
			log.Infof("TXPoolServer monitor get mining info")

			miningInfo, err := consutils.GetMiningInfo()
			if err != nil {
				log.Infof("TXPoolServer monitor: fail to get mining info: %s", err)
			}

			if miningInfo.View > s.miningInfo.View {
				s.mu.Lock()
				s.miningInfo = miningInfo
				s.mu.Unlock()

				list := s.pocPool.RemoveFuturePoC(miningInfo.View)
				go func(pocList []*tc.PoCEntry) {
					for _, param := range pocList {
						s.assignParamToWorker(param.Param, tc.PoCSender)
					}
				}(list)

				go s.pocPool.UpdateParam(miningInfo.View)
			}

		}
	}
}

// checkPendingBlockOk checks whether a block from consensus is verified.
// If some transaction is invalid, return the result directly at once, no
// need to wait for verifying the complete block.
func (s *TXPoolServer) checkPendingBlockOk(hash common.Uint256,
	err errors.ErrCode) {

	// Check if the tx is in pending block, if yes, move it to
	// the verified tx list
	s.pendingBlock.mu.Lock()
	defer s.pendingBlock.mu.Unlock()

	tx, ok := s.pendingBlock.unProcessedTxs[hash]
	if !ok {
		return
	}

	entry := &tc.VerifyTxResult{
		Height:  s.pendingBlock.height,
		Tx:      tx,
		ErrCode: err,
	}

	s.pendingBlock.processedTxs[hash] = entry
	delete(s.pendingBlock.unProcessedTxs, hash)

	// if the tx is invalid, send the response at once
	if err != errors.ErrNoError ||
		len(s.pendingBlock.unProcessedTxs) == 0 {
		s.sendBlkResult2Consensus()
	}
}

// getPendingListSize return the length of the pending tx list.
func (s *TXPoolServer) getPendingListSize() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.allPendingTxs)
}

// getHeight return current block height
func (s *TXPoolServer) getHeight() uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.height
}

// setHeight set current block height
func (s *TXPoolServer) setHeight(height uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.height = height
}

// getGasPrice returns the current gas price enforced by the transaction pool
func (s *TXPoolServer) getGasPrice() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gasPrice
}

// removePendingTx removes a transaction from the pending list
// when it is handled. And if the submitter of the valid transaction
// is from http, broadcast it to the network. Meanwhile, check if it
// is in the block from consensus.
func (s *TXPoolServer) removePendingTx(hash common.Uint256,
	err errors.ErrCode) {

	s.mu.Lock()

	pt, ok := s.allPendingTxs[hash]
	if !ok {
		s.mu.Unlock()
		return
	}

	if err == errors.ErrNoError && ((pt.sender == tc.HttpSender) ||
		(pt.sender == tc.NetSender && !s.disableBroadcastNetTx)) {
		if s.Net != nil {
			msg := msgpack.NewTxn(pt.tx)
			go s.Net.Broadcast(msg)
		}
	}

	if pt.sender == tc.HttpSender && pt.ch != nil {
		replyTxResult(pt.ch, hash, err, err.Error())
	}

	delete(s.allPendingTxs, hash)

	if len(s.allPendingTxs) < tc.MAX_LIMITATION {
		select {
		case s.slots <- struct{}{}:
		default:
			log.Debug("removePendingTx: slots is full")
		}
	}

	s.mu.Unlock()

	// Check if the tx is in the pending block and
	// the pending block is verified
	s.checkPendingBlockOk(hash, err)
}

// setPendingTx adds a transaction to the pending list, if the
// transaction is already in the pending list, just return false.
func (s *TXPoolServer) setPendingTx(tx *tx.Transaction,
	sender tc.SenderType, txResultCh chan *tc.TxResult) bool {

	s.mu.Lock()
	defer s.mu.Unlock()
	if ok := s.allPendingTxs[tx.Hash()]; ok != nil {
		log.Debugf("setPendingTx: transaction %x already in the verifying process",
			tx.Hash())
		return false
	}

	pt := &serverPendingTx{
		tx:     tx,
		sender: sender,
		ch:     txResultCh,
	}

	s.allPendingTxs[tx.Hash()] = pt
	return true
}

// assignTxToWorker assigns a new transaction to a worker by LB
func (s *TXPoolServer) assignTxToWorker(tx *tx.Transaction,
	sender tc.SenderType, txResultCh chan *tc.TxResult) bool {

	if tx == nil {
		return false
	}

	if ok := s.setPendingTx(tx, sender, txResultCh); !ok {
		s.increaseStats(tc.DuplicateStats)
		if sender == tc.HttpSender && txResultCh != nil {
			replyTxResult(txResultCh, tx.Hash(), errors.ErrDuplicateInput,
				"duplicated transaction input detected")
		}
		return false
	}
	// Add the rcvTxn to the worker
	lb := make(tc.LBSlice, len(s.workers))
	for i := 0; i < len(s.workers); i++ {
		pending := atomic.LoadInt64(&s.workers[i].pendingTxLen)
		entry := tc.LB{
			Size:     len(s.workers[i].rcvTXCh) + int(pending),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}
	sort.Sort(lb)
	s.workers[lb[0].WorkerID].rcvTXCh <- tx
	return true
}

// assignRspToWorker assigns a check response from the validator to
// the correct worker.
func (s *TXPoolServer) assignRspToWorker(rsp *types.CheckResponse) bool {

	if rsp == nil {
		return false
	}

	if rsp.WorkerId >= 0 && rsp.WorkerId < uint8(len(s.workers)) {
		s.workers[rsp.WorkerId].rspCh <- rsp
	}

	if rsp.ErrCode == errors.ErrNoError {
		s.increaseStats(tc.SuccessStats)
	} else {
		s.increaseStats(tc.FailureStats)
		if rsp.Type == types.Stateless {
			s.increaseStats(tc.SigErrStats)
		} else {
			s.increaseStats(tc.StateErrStats)
		}
	}
	return true
}

// GetPID returns an actor pid with the actor type, If the type
// doesn't exist, return nil.
func (s *TXPoolServer) GetPID(actor tc.ActorType) *actor.PID {
	if actor < tc.TxActor || actor >= tc.MaxActor {
		return nil
	}

	return s.actors[actor]
}

// RegisterActor registers an actor with the actor type and pid.
func (s *TXPoolServer) RegisterActor(actor tc.ActorType, pid *actor.PID) {
	s.actors[actor] = pid
}

// UnRegisterActor cancels the actor with the actor type.
func (s *TXPoolServer) UnRegisterActor(actor tc.ActorType) {
	delete(s.actors, actor)
}

// registerValidator registers a validator to verify a transaction.
func (s *TXPoolServer) registerValidator(v *types.RegisterValidator) {
	s.validators.Lock()
	defer s.validators.Unlock()

	_, ok := s.validators.entries[v.Type]

	if !ok {
		s.validators.entries[v.Type] = make([]*types.RegisterValidator, 0, 1)
	}
	s.validators.entries[v.Type] = append(s.validators.entries[v.Type], v)
}

// unRegisterValidator cancels a validator with the verify type and id.
func (s *TXPoolServer) unRegisterValidator(checkType types.VerifyType,
	id string) {

	s.validators.Lock()
	defer s.validators.Unlock()

	tmpSlice, ok := s.validators.entries[checkType]
	if !ok {
		log.Errorf("unRegisterValidator: validator not found with type:%d, id:%s",
			checkType, id)
		return
	}

	for i, v := range tmpSlice {
		if v.Id == id {
			s.validators.entries[checkType] =
				append(tmpSlice[0:i], tmpSlice[i+1:]...)
			if v.Sender != nil {
				v.Sender.Tell(&types.UnRegisterAck{Id: id, Type: checkType})
			}
			if len(s.validators.entries[checkType]) == 0 {
				delete(s.validators.entries, checkType)
			}
		}
	}
}

// getNextValidatorPIDs returns the next pids to verify the transaction using
// roundRobin LB.
func (s *TXPoolServer) getNextValidatorPIDs() []*actor.PID {
	s.validators.Lock()
	defer s.validators.Unlock()

	if len(s.validators.entries) == 0 {
		return nil
	}

	ret := make([]*actor.PID, 0, len(s.validators.entries))
	for k, v := range s.validators.entries {
		lastIdx := s.validators.state.state[k]
		next := (lastIdx + 1) % len(v)
		s.validators.state.state[k] = next
		ret = append(ret, v[next].Sender)
	}
	return ret
}

// getNextValidatorPID returns the next pid with the verify type using roundRobin LB
func (s *TXPoolServer) getNextValidatorPID(key types.VerifyType) *actor.PID {
	s.validators.Lock()
	defer s.validators.Unlock()

	length := len(s.validators.entries[key])
	if length == 0 {
		return nil
	}

	entries := s.validators.entries[key]
	lastIdx := s.validators.state.state[key]
	next := (lastIdx + 1) % length
	s.validators.state.state[key] = next
	return entries[next].Sender
}

// Stop stops server and workers.
func (s *TXPoolServer) Stop() {
	for _, v := range s.actors {
		v.Stop()
	}
	//Stop worker
	for i := 0; i < len(s.workers); i++ {
		s.workers[i].stop()
	}
	s.wg.Wait()

	if s.slots != nil {
		close(s.slots)
	}

	if s.pocSlots != nil {
		close(s.pocSlots)
	}
}

// getTransaction returns a transaction with the transaction hash.
func (s *TXPoolServer) getTransaction(hash common.Uint256) *tx.Transaction {
	return s.txPool.GetTransaction(hash)
}

// getTxPool returns a tx list for consensus.
func (s *TXPoolServer) getTxPool(byCount bool, height uint32) []*tc.TXEntry {
	s.setHeight(height)

	avlTxList, oldTxList := s.txPool.GetTxPool(byCount, height)

	for _, t := range oldTxList {
		s.delTransaction(t)
		s.reVerifyStateful(t, tc.NilSender)
	}

	return avlTxList
}

// getTxCount returns current tx count, including pending and verified
func (s *TXPoolServer) getTxCount() []uint32 {
	ret := make([]uint32, 0)
	ret = append(ret, uint32(s.txPool.GetTransactionCount()))
	ret = append(ret, uint32(s.getPendingListSize()))
	return ret
}

// getPendingTxs returns a currently pending tx list
func (s *TXPoolServer) getPendingTxs(byCount bool) []*tx.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ret := make([]*tx.Transaction, 0, len(s.allPendingTxs))
	for _, v := range s.allPendingTxs {
		ret = append(ret, v.tx)
	}
	return ret
}

// getTxHashList returns a currently pending tx hash list
func (s *TXPoolServer) getTxHashList() []common.Uint256 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	txHashPool := s.txPool.GetTransactionHashList()
	ret := make([]common.Uint256, 0, len(s.allPendingTxs)+len(txHashPool))
	existedTxHash := make(map[common.Uint256]bool)
	for _, hash := range txHashPool {
		ret = append(ret, hash)
		existedTxHash[hash] = true
	}
	for _, v := range s.allPendingTxs {
		hash := v.tx.Hash()
		if !existedTxHash[hash] {
			ret = append(ret, hash)
			existedTxHash[hash] = true
		}
	}
	return ret
}

// cleanTransactionList cleans the txs in the block from the ledger
func (s *TXPoolServer) cleanTransactionList(txs []*tx.Transaction, height uint32) {
	s.txPool.CleanTransactionList(txs)

	// Check whether to update the gas price and remove txs below the
	// threshold
	if height%tc.UPDATE_FREQUENCY == 0 {
		gasPrice := getGasPriceConfig()
		s.mu.Lock()
		oldGasPrice := s.gasPrice
		s.gasPrice = gasPrice
		s.mu.Unlock()
		if oldGasPrice != gasPrice {
			log.Infof("Transaction pool price threshold updated from %d to %d",
				oldGasPrice, gasPrice)
		}

		if oldGasPrice < gasPrice {
			s.txPool.RemoveTxsBelowGasPrice(gasPrice)
		}
	}
	// Cleanup tx pool
	if !s.disablePreExec {
		remain := s.txPool.Remain()
		for _, t := range remain {
			if ok, _ := preExecCheck(t); !ok {
				log.Debugf("cleanTransactionList: preExecCheck tx %x failed", t.Hash())
				continue
			}
			s.reVerifyStateful(t, tc.NilSender)
		}
	}
}

// delTransaction deletes a transaction in the tx pool.
func (s *TXPoolServer) delTransaction(t *tx.Transaction) {
	s.txPool.DelTxList(t)
}

// addTxList adds a valid transaction to the tx pool.
func (s *TXPoolServer) addTxList(txEntry *tc.TXEntry) bool {
	ret := s.txPool.AddTxList(txEntry)
	if !ret {
		s.increaseStats(tc.DuplicateStats)
	}
	return ret
}

// increaseStats increases the count with the stats type
func (s *TXPoolServer) increaseStats(v tc.TxnStatsType) {
	s.stats.Lock()
	defer s.stats.Unlock()
	s.stats.count[v-1]++
}

// getStats returns the transaction statistics
func (s *TXPoolServer) getStats() []uint64 {
	s.stats.RLock()
	defer s.stats.RUnlock()
	ret := make([]uint64, 0, len(s.stats.count))
	ret = append(ret, s.stats.count...)

	return ret
}

// checkTx checks whether a transaction is in the pending list or
// the transacton pool
func (s *TXPoolServer) checkTx(hash common.Uint256) bool {
	// Check if the tx is in pending list
	s.mu.RLock()
	if ok := s.allPendingTxs[hash]; ok != nil {
		s.mu.RUnlock()
		return true
	}
	s.mu.RUnlock()

	// Check if the tx is in txn pool
	if res := s.txPool.GetTransaction(hash); res != nil {
		return true
	}

	return false
}

// getTxStatusReq returns a transaction's status with the transaction hash.
func (s *TXPoolServer) getTxStatusReq(hash common.Uint256) *tc.TxStatus {
	for i := 0; i < len(s.workers); i++ {
		ret := s.workers[i].GetTxStatus(hash)
		if ret != nil {
			return ret
		}
	}

	return s.txPool.GetTxStatus(hash)
}

// getTransactionCount returns the tx size of the transaction pool.
func (s *TXPoolServer) getTransactionCount() int {
	return s.txPool.GetTransactionCount()
}

// reVerifyStateful re-verify a transaction's stateful data.
func (s *TXPoolServer) reVerifyStateful(tx *tx.Transaction, sender tc.SenderType) {
	if ok := s.setPendingTx(tx, sender, nil); !ok {
		s.increaseStats(tc.DuplicateStats)
		return
	}

	// Add the rcvTxn to the worker
	lb := make(tc.LBSlice, len(s.workers))
	for i := 0; i < len(s.workers); i++ {
		pending := atomic.LoadInt64(&s.workers[i].pendingTxLen)
		entry := tc.LB{
			Size:     len(s.workers[i].stfTxCh) + int(pending),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}

	sort.Sort(lb)
	s.workers[lb[0].WorkerID].stfTxCh <- tx
}

// sendBlkResult2Consensus sends the result of verifying block to  consensus
func (s *TXPoolServer) sendBlkResult2Consensus() {
	rsp := &tc.VerifyBlockRsp{
		TxnPool: make([]*tc.VerifyTxResult,
			0, len(s.pendingBlock.processedTxs)),
	}
	for _, v := range s.pendingBlock.processedTxs {
		rsp.TxnPool = append(rsp.TxnPool, v)
	}

	if s.pendingBlock.sender != nil {
		s.pendingBlock.sender.Tell(rsp)
	}

	// Clear the processedTxs for the next block verify req
	for k := range s.pendingBlock.processedTxs {
		delete(s.pendingBlock.processedTxs, k)
	}
}

// verifyBlock verifies the block from consensus.
// There are three cases to handle.
// 1, for those unverified txs, assign them to the available worker;
// 2, for those verified txs whose height >= block's height, nothing to do;
// 3, for those verified txs whose height < block's height, re-verify their
// stateful data.
func (s *TXPoolServer) verifyBlock(req *tc.VerifyBlockReq, sender *actor.PID) {
	if req == nil || len(req.Txs) == 0 {
		return
	}

	s.setHeight(req.Height)
	s.pendingBlock.mu.Lock()
	defer s.pendingBlock.mu.Unlock()

	s.pendingBlock.sender = sender
	s.pendingBlock.height = req.Height
	s.pendingBlock.processedTxs = make(map[common.Uint256]*tc.VerifyTxResult, len(req.Txs))
	s.pendingBlock.unProcessedTxs = make(map[common.Uint256]*tx.Transaction, 0)

	txs := make(map[common.Uint256]bool, len(req.Txs))

	// Check whether a tx's gas price is lower than the required, if yes,
	// just return error
	for _, t := range req.Txs {
		if t.GasPrice < s.gasPrice {
			entry := &tc.VerifyTxResult{
				Height:  s.pendingBlock.height,
				Tx:      t,
				ErrCode: errors.ErrGasPrice,
			}
			s.pendingBlock.processedTxs[t.Hash()] = entry
			s.sendBlkResult2Consensus()
			return
		}
		// Check whether double spent
		if _, ok := txs[t.Hash()]; ok {
			entry := &tc.VerifyTxResult{
				Height:  s.pendingBlock.height,
				Tx:      t,
				ErrCode: errors.ErrDoubleSpend,
			}
			s.pendingBlock.processedTxs[t.Hash()] = entry
			s.sendBlkResult2Consensus()
			return
		}
		txs[t.Hash()] = true
	}

	checkBlkResult := s.txPool.GetUnverifiedTxs(req.Txs, req.Height)

	for _, t := range checkBlkResult.UnverifiedTxs {
		s.assignTxToWorker(t, tc.NilSender, nil)
		s.pendingBlock.unProcessedTxs[t.Hash()] = t
	}

	for _, t := range checkBlkResult.OldTxs {
		s.reVerifyStateful(t, tc.NilSender)
		s.pendingBlock.unProcessedTxs[t.Hash()] = t
	}

	for _, t := range checkBlkResult.VerifiedTxs {
		s.pendingBlock.processedTxs[t.Tx.Hash()] = t
	}

	/* If all the txs in the blocks are verified, send response
	 * to the consensus directly
	 */
	if len(s.pendingBlock.unProcessedTxs) == 0 {
		s.sendBlkResult2Consensus()
	}
}

// setPendingTx adds a transaction to the pending list, if the
// transaction is already in the pending list, just return false.
func (s *TXPoolServer) setPendingParam(param *gov.SubmitNonceParam,
	sender tc.SenderType) bool {

	s.mu.Lock()
	defer s.mu.Unlock()

	if ok := s.allPendingParam[param.Hash()]; ok != nil {
		log.Debugf("setPendingParam: transaction %x already in the verifying process",
			param.Hash())
		return false
	}

	pt := &serverPendingParam{
		param:  param,
		sender: sender,
	}

	s.allPendingParam[param.Hash()] = pt
	return true
}

// getTransaction returns a poc param with the hash.
func (s *TXPoolServer) getParam(hash common.Uint256) *gov.SubmitNonceParam {
	return s.pocPool.GetParam(hash)
}

// getPoC returns winner PoC param  for consensus.
func (s *TXPoolServer) getPoCParam(view uint32) *gov.SubmitNonceParam {
	param := s.pocPool.GetPoCParam(view)
	return param
}

func (s *TXPoolServer) removePendingParam(hash common.Uint256, err errors.ErrCode) {

	s.mu.Lock()

	_, ok := s.allPendingParam[hash]
	if !ok {
		s.mu.Unlock()
		return
	}

	delete(s.allPendingParam, hash)
	s.mu.Unlock()

}

// assignTxToWorker assigns a new transaction to a worker by LB
func (s *TXPoolServer) assignParamToWorker(param *gov.SubmitNonceParam,
	sender tc.SenderType) bool {

	if param == nil {
		return false
	}

	s.mu.Lock()
	miningInfo := s.miningInfo
	s.mu.Unlock()

	if param.View < miningInfo.View {
		log.Debugf("assignParamToWorker: submitted view %d less than %d",
			param.View, miningInfo.View)
		return false
	} else if param.View > miningInfo.View {
		if config.DefConfig.Consensus.EnableConsensus {
			//Put future poc in furture list of pocPool!
			pocEntry := &tc.PoCEntry{
				Param: param,
			}
			ok := s.pocPool.AddFuturePoC(pocEntry)

			if ok {
				log.Debugf("assignParamToWorker: add param %v for view %d to future list", param, param.View)
			}
			return ok
		}
	}

	if ok := s.setPendingParam(param, sender); !ok {
		return false
	}

	// Assign the poc param to the worker
	lb := make(tc.LBSlice, len(s.pocWorkers))
	for i := 0; i < len(s.pocWorkers); i++ {
		entry := tc.LB{Size: len(s.pocWorkers[i].rcvTXCh) +
			len(s.pocWorkers[i].pendingParamList),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}
	sort.Sort(lb)

	s.pocWorkers[lb[0].WorkerID].rcvTXCh <- &CheckPoCReq{param: param, info: miningInfo}
	return true
}

// addPoCList adds a valid poc puzzle result to the poc pool.
func (s *TXPoolServer) addPoCList(entry *tc.PoCEntry) bool {
	ret := s.pocPool.AddPoC(entry)

	return ret
}
