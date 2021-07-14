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

package proc

import (
	"encoding/binary"
	"sync"

	cutils "github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	consutils "github.com/saveio/themis/consensus/utils"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/core/utils"
	"github.com/saveio/themis/errors"
	msgpack "github.com/saveio/themis/p2pserver/message/msg_pack"
	gov "github.com/saveio/themis/smartcontract/service/native/governance"
	tc "github.com/saveio/themis/txnpool/common"
)

// pendingParam contains the SubmitNonceParam, the verified result
type pendingParam struct {
	param *gov.SubmitNonceParam // That is unverified or on the verifying process
	ret   bool                  // if the puzzle result is valid
}

// txPoolWorker handles the tasks scheduled by server
type pocPoolWorker struct {
	mu      sync.RWMutex
	workId  uint8             // Worker ID
	rcvTXCh chan *CheckPoCReq // The channel of receive transaction

	server *TXPoolServer // The txn pool server pointer

	stopCh           chan bool                        // stop routine
	pendingParamList map[common.Uint256]*pendingParam // The transaction on the verifying process
}

// init initializes the worker with the configured settings
func (worker *pocPoolWorker) init(workID uint8, s *TXPoolServer) {
	worker.rcvTXCh = make(chan *CheckPoCReq, tc.MAX_PENDING_TXN)
	worker.pendingParamList = make(map[common.Uint256]*pendingParam)
	worker.stopCh = make(chan bool)
	worker.workId = workID
	worker.server = s
}

// putParamPool adds a valid poc puzzle to the poc pool and removes it from
// the pending list.
func (worker *pocPoolWorker) putPoCPool(pp *pendingParam) bool {
	needBroadcast := false

	pocEntry := &tc.PoCEntry{
		Param: pp.param,
		Ret:   pp.ret,
	}

	if pocEntry.Ret {
		needBroadcast = worker.server.addPoCList(pocEntry)
	}
	worker.server.removePendingParam(pp.param.Hash(), errors.ErrNoError)

	//non consensus node always forward
	if !config.DefConfig.Consensus.EnableConsensus {
		needBroadcast = true
	}

	// notify p2p to gossip valid poc to peers!
	if worker.server.Net != nil && needBroadcast {
		go worker.server.Net.Broadcast(msgpack.NewSubmitNonce(pocEntry.Param))
		log.Debugf("putPoCPool: try to send poc to p2p for gossip, hash %v",
			pocEntry.Param.Hash())
	}

	return true
}

// verifyTx prepares a check request and sends it to the validators.
func (worker *pocPoolWorker) verifyParam(pocReq *CheckPoCReq) {
	var returnSlot bool

	defer func() {
		if returnSlot {
			worker.server.pocSlots <- struct{}{}
		}
	}()

	param := pocReq.param
	if p := worker.server.getParam(param.Hash()); p != nil {
		log.Debugf("verifyParam: param %x already in the poc pool",
			p.Hash())
		worker.server.removePendingParam(p.Hash(), errors.ErrDuplicateInput)
		return
	}

	if _, ok := worker.pendingParamList[param.Hash()]; ok {
		log.Debugf("verifyParam: param %x already in the verifying process",
			param.Hash())
		return
	}

	invalid := false
	//don't forward the param if miner use other miner's plot file!
	accountId := uint64(cutils.WalletAddressToId([]byte(param.Address.ToBase58())))
	if accountId != uint64(param.Id) {
		log.Debugf("verifyParam: param id %d doesn't match address %s",
			param.Id, param.Address.ToBase58())
		invalid = true
	}

	//check if the plot used is registered in FS
	plotFile := param.PlotName
	reg, _ := consutils.GetPlotRegInfo(param.Address, plotFile)
	log.Debugf("verifyParam: param check plot %s for address %s", plotFile, param.Address.ToBase58())
	if !reg {
		log.Infof("verifyParam: plot file %s not registered in FS!", plotFile)
		invalid = true
	}

	if invalid {
		worker.mu.Lock()
		worker.server.removePendingParam(param.Hash(), errors.ErrNoError)
		delete(worker.pendingParamList, param.Hash())
		worker.mu.Unlock()
		return
	}

	//skip deadline verify, forward directly
	skipVerify := false
	//non-consensus node, just record and forward all the poc without verify!
	if !config.DefConfig.Consensus.EnableConsensus {
		skipVerify = true
	} else if existPoC := worker.server.getPoCParam(param.View); existPoC != nil {
		//consensus node skip verify deadline which is not better than current one
		if existPoC.Deadline < param.Deadline {
			skipVerify = true
			log.Debugf("[pocPoolWorker] already has better dealine %d for view %d", existPoC.Deadline, param.View)
			log.Debugf("[pocPoolWorker] skip verify incomming param %v", param)
		}
	}

	if skipVerify {
		//enforce ret true!
		p := &pendingParam{
			param: param,
			ret:   true,
		}
		worker.mu.Lock()
		worker.putPoCPool(p)
		delete(worker.pendingParamList, param.Hash())
		worker.mu.Unlock()
		return
	}

	//begin time consuming operation
	<-worker.server.pocSlots
	returnSlot = true

	p := &pendingParam{
		param: param,
		ret:   false,
	}

	// Add it to the pending poc list
	worker.mu.Lock()
	worker.pendingParamList[param.Hash()] = p
	worker.mu.Unlock()

	//construct plot file
	plot := types.NewMiningPlot(param.Id, param.Nonce)
	miningInfo := pocReq.info
	gensig := miningInfo.GenerationSignature.ToArray()
	scoop := utils.CalculateScoop(uint64(miningInfo.View), gensig)

	scoopData := plot.GetScoopData(int(scoop))

	data := append([]byte{}, gensig[:]...) // gensig 32 bytes
	data = append(data, scoopData[:]...)   // scoop 64 bytes

	md := common.NewShabal256()
	md.Update(data, 0, int64(len(data)))
	hash := md.Digest()

	//same with burst calculateHit
	deadline := binary.LittleEndian.Uint64(hash)

	log.Debugf("verifyParam dump param %v", param)
	log.Debugf("verifyParam for view: %d, from id: %d, nonce: %d, deadline: %d\n", param.View, param.Id, param.Nonce, param.Deadline)
	log.Debugf("verifyParam scoop:%d, deadline calculated: %d\n", scoop, deadline)

	if param.Deadline == deadline {
		log.Debugf("verifyParam accept the nonce!")
		p.ret = true
	} else {
		log.Debugf("verfiyParam failed deadline not match, expect %v but got %v, p.ret %v",
			deadline, param.Deadline, p.ret)
	}

	worker.mu.Lock()
	worker.putPoCPool(p)
	delete(worker.pendingParamList, param.Hash())
	worker.mu.Unlock()

	return
}

// Start is the main event loop.
func (worker *pocPoolWorker) start() {

	for {
		select {
		case <-worker.stopCh:
			worker.server.wg.Done()
			return
		case rcvParam, ok := <-worker.rcvTXCh:
			if ok {
				// Verify poc Param
				worker.verifyParam(rcvParam)
			}

		}
	}
}

// stop closes/releases channels and stops timer
func (worker *pocPoolWorker) stop() {

	if worker.rcvTXCh != nil {
		close(worker.rcvTXCh)
	}

	if worker.stopCh != nil {
		worker.stopCh <- true
		close(worker.stopCh)
	}
}
