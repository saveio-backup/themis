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

package ledger

import (
	"fmt"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/payload"
	"github.com/saveio/themis/core/states"
	"github.com/saveio/themis/core/store"
	"github.com/saveio/themis/core/store/ledgerstore"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/smartcontract/event"
	cstate "github.com/saveio/themis/smartcontract/states"
)

var DefLedger *Ledger

type Ledger struct {
	ldgStore store.LedgerStore
}

func NewLedger(dataDir string, stateHashHeight uint32) (*Ledger, error) {
	ldgStore, err := ledgerstore.NewLedgerStore(dataDir, stateHashHeight)
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore error %s", err)
	}
	return &Ledger{
		ldgStore: ldgStore,
	}, nil
}

func (self *Ledger) GetStore() store.LedgerStore {
	return self.ldgStore
}

func (self *Ledger) Init(defaultBookkeeper []keypair.PublicKey, genesisBlock *types.Block) error {
	err := self.ldgStore.InitLedgerStoreWithGenesisBlock(genesisBlock, defaultBookkeeper)
	if err != nil {
		return fmt.Errorf("InitLedgerStoreWithGenesisBlock error %s", err)
	}
	return nil
}

func (self *Ledger) AddHeaders(headers []*types.Header) error {
	return self.ldgStore.AddHeaders(headers)
}

func (self *Ledger) AddBlock(block *types.Block, ccMsg *types.CrossChainMsg, stateMerkleRoot common.Uint256) error {
	err := self.ldgStore.AddBlock(block, ccMsg, stateMerkleRoot)
	if err != nil {
		log.Errorf("Ledger AddBlock BlockHeight:%d BlockHash:%x error:%s", block.Header.Height, block.Hash(), err)
	}
	return err
}

func (self *Ledger) ExecuteBlock(b *types.Block) (store.ExecuteResult, error) {
	return self.ldgStore.ExecuteBlock(b)
}

func (self *Ledger) SubmitBlock(b *types.Block, crossChainMsg *types.CrossChainMsg, exec store.ExecuteResult) error {
	return self.ldgStore.SubmitBlock(b, crossChainMsg, exec)
}

func (self *Ledger) GetStateMerkleRoot(height uint32) (result common.Uint256, err error) {
	return self.ldgStore.GetStateMerkleRoot(height)
}

func (self *Ledger) GetCrossStatesRoot(height uint32) (common.Uint256, error) {
	return self.ldgStore.GetCrossStatesRoot(height)
}

func (self *Ledger) GetBlockRootWithNewTxRoots(startHeight uint32, txRoots []common.Uint256) common.Uint256 {
	return self.ldgStore.GetBlockRootWithNewTxRoots(startHeight, txRoots)
}

func (self *Ledger) GetBlockByHeight(height uint32) (*types.Block, error) {
	return self.ldgStore.GetBlockByHeight(height)
}

func (self *Ledger) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	return self.ldgStore.GetBlockByHash(blockHash)
}

func (self *Ledger) GetHeaderByHeight(height uint32) (*types.Header, error) {
	return self.ldgStore.GetHeaderByHeight(height)
}

func (self *Ledger) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	return self.ldgStore.GetHeaderByHash(blockHash)
}
func (self *Ledger) GetRawHeaderByHash(blockHash common.Uint256) (*types.RawHeader, error) {
	return self.ldgStore.GetRawHeaderByHash(blockHash)
}

func (self *Ledger) GetBlockHash(height uint32) common.Uint256 {
	return self.ldgStore.GetBlockHash(height)
}

func (self *Ledger) GetTransaction(txHash common.Uint256) (*types.Transaction, error) {
	tx, _, err := self.ldgStore.GetTransaction(txHash)
	return tx, err
}

func (self *Ledger) GetTransactionWithHeight(txHash common.Uint256) (*types.Transaction, uint32, error) {
	return self.ldgStore.GetTransaction(txHash)
}

func (self *Ledger) GetCurrentBlockHeight() uint32 {
	return self.ldgStore.GetCurrentBlockHeight()
}

func (self *Ledger) GetCurrentBlockHash() common.Uint256 {
	return self.ldgStore.GetCurrentBlockHash()
}

func (self *Ledger) GetCurrentHeaderHeight() uint32 {
	return self.ldgStore.GetCurrentHeaderHeight()
}

func (self *Ledger) GetCurrentHeaderHash() common.Uint256 {
	return self.ldgStore.GetCurrentHeaderHash()
}

func (self *Ledger) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return self.ldgStore.IsContainTransaction(txHash)
}

func (self *Ledger) IsContainBlock(blockHash common.Uint256) (bool, error) {
	return self.ldgStore.IsContainBlock(blockHash)
}

func (self *Ledger) GetBookkeeperState() (*states.BookkeeperState, error) {
	return self.ldgStore.GetBookkeeperState()
}

func (self *Ledger) GetStorageItem(codeHash common.Address, key []byte) ([]byte, error) {
	storageKey := &states.StorageKey{
		ContractAddress: codeHash,
		Key:             key,
	}
	storageItem, err := self.ldgStore.GetStorageItem(storageKey)
	if err != nil {
		return nil, err
	}
	if storageItem == nil {
		return nil, nil
	}
	return storageItem.Value, nil
}

func (self *Ledger) GetContractState(contractHash common.Address) (*payload.DeployCode, error) {
	return self.ldgStore.GetContractState(contractHash)
}

func (self *Ledger) GetMerkleProof(proofHeight, rootHeight uint32) ([]common.Uint256, error) {
	return self.ldgStore.GetMerkleProof(proofHeight, rootHeight)
}

func (self *Ledger) PreExecuteContract(tx *types.Transaction) (*cstate.PreExecResult, error) {
	return self.ldgStore.PreExecuteContract(tx)
}

func (self *Ledger) PreExecuteContractBatch(txes []*types.Transaction, atomic bool) ([]*cstate.PreExecResult, uint32, error) {
	return self.ldgStore.PreExecuteContractBatch(txes, atomic)
}

func (self *Ledger) GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error) {
	return self.ldgStore.GetEventNotifyByTx(tx)
}

func (self *Ledger) GetEventNotifyByBlock(height uint32) ([]*event.ExecuteNotify, error) {
	return self.ldgStore.GetEventNotifyByBlock(height)
}

func (self *Ledger) GetCrossChainMsg(height uint32) (*types.CrossChainMsg, error) {
	return self.ldgStore.GetCrossChainMsg(height)
}

func (self *Ledger) GetCrossStatesProof(height uint32, key []byte) ([]byte, error) {
	return self.ldgStore.GetCrossStatesProof(height, key)
}

func (self *Ledger) Close() error {
	return self.ldgStore.Close()
}

func (self *Ledger) EnableBlockPrune(numBeforeCurr uint32) {
	self.ldgStore.EnableBlockPrune(numBeforeCurr)
}

func (self *Ledger) GetEventNotifyByEventId(contractAddress common.Address, address common.Address, eventId uint32) (
	[]*event.ExecuteNotify, error) {
	return self.ldgStore.GetEventNotifyByEventId(contractAddress, address, eventId)
}
func (self *Ledger) GetEventNotifyByEventIdAndHeight(contractAddress common.Address, address []byte,
	eventId, startHeight, endHeight uint32) ([]*event.ExecuteNotify, error) {
	return self.ldgStore.GetEventNotifyByEventIdAndHeights(contractAddress, address, eventId, startHeight, endHeight)
}
