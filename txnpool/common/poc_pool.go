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

// Package common provides constants, common types for other packages
package common

import (
	"sync"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	gov "github.com/saveio/themis/smartcontract/service/native/governance"
)

type PoCEntry struct {
	Param *gov.SubmitNonceParam // transaction which has been verified or will be verified in the furture
	Ret   bool                  // the validate result
}

// PoCPool contains all currently valid poc puzzle.
// enter the pool when they are valid from the network,
// or PoC consensus. They exit the pool when PoCPool found
// current PoC period bigger than PoC Puzzle
type PoCPool struct {
	sync.RWMutex
	pocList        map[common.Uint256]*PoCEntry
	pocByView      map[uint32]*PoCEntry
	pocByViewMiner map[uint32]map[common.Address]*PoCEntry
	furtureList    map[common.Uint256]*PoCEntry
	view           uint32
}

// Init creates a new poc pool to gather.
func (pp *PoCPool) Init(view uint32) {
	pp.Lock()
	defer pp.Unlock()
	pp.pocList = make(map[common.Uint256]*PoCEntry)
	pp.pocByView = make(map[uint32]*PoCEntry)
	pp.pocByViewMiner = make(map[uint32]map[common.Address]*PoCEntry)
	pp.furtureList = make(map[common.Uint256]*PoCEntry)
	pp.view = view
}

func (pp *PoCPool) AddPoC(entry *PoCEntry) bool {
	pp.Lock()
	defer pp.Unlock()

	hash := entry.Param.Hash()
	if _, ok := pp.pocList[hash]; ok {
		log.Infof("AddPoC: poc params %x is already in the pool", hash)
		return false
	}

	isBest := false
	pp.pocList[hash] = entry
	if other, ok := pp.pocByView[entry.Param.View]; ok {
		log.Debugf("AddPoC: view %d already have deadline %d, new deadline %d",
			entry.Param.View, other.Param.Deadline, entry.Param.Deadline)

		if entry.Param.Deadline < other.Param.Deadline {
			log.Debugf("AddPoC: view %d replace deadline %d with deadline %d",
				entry.Param.View, other.Param.Deadline, entry.Param.Deadline)

			pp.pocByView[entry.Param.View] = entry
			isBest = true
		}
	} else {
		pp.pocByView[entry.Param.View] = entry
		isBest = true
	}

	walletAddr := entry.Param.Address
	if _, ok := pp.pocByViewMiner[entry.Param.View]; !ok {
		pp.pocByViewMiner[entry.Param.View] = make(map[common.Address]*PoCEntry)

	}
	log.Debugf("AddPoC: view %d add param from miner %s", entry.Param.View, walletAddr)
	pp.pocByViewMiner[entry.Param.View][walletAddr] = entry

	return isBest
}

// return PoC puzzle for view.
func (pp *PoCPool) GetPoCParam(view uint32) *gov.SubmitNonceParam {
	pp.RLock()
	defer pp.RUnlock()

	if len(pp.pocList) == 0 {
		log.Debugf("pocList is empty for view %v", view)
		return nil
	}

	if len(pp.pocByView) == 0 {
		log.Debugf("pocByView is empty for view %v", view)
		return nil
	}

	if entry, ok := pp.pocByView[view]; ok {
		return entry.Param
	}
	log.Debugf("poc param not found for view %v", view)

	return nil
}

// GetParam returns a poc param if it is contained in the pool
// and nil otherwise.
func (pp *PoCPool) GetParam(hash common.Uint256) *gov.SubmitNonceParam {
	pp.RLock()
	defer pp.RUnlock()
	if tx := pp.pocList[hash]; tx == nil {
		return nil
	}
	return pp.pocList[hash].Param
}

func (pp *PoCPool) UpdateParam(newView uint32) {
	pp.Lock()
	defer pp.Unlock()

	oldView := make(map[uint32]int)
	for view, _ := range pp.pocByView {
		if view+1 < newView {
			oldView[view] = 1
		}
	}

	for view, _ := range oldView {
		delete(pp.pocByView, view)
		delete(pp.pocByViewMiner, view)
	}

	oldEntry := make(map[common.Uint256]int)
	for hash, entry := range pp.pocList {
		if entry.Param.View+1 < newView {
			oldEntry[hash] = 1
		}
	}

	for hash, entry := range pp.furtureList {
		if entry.Param.View <= newView {
			oldEntry[hash] = 1
		}
	}

	for hash, _ := range oldEntry {
		delete(pp.pocList, hash)
		delete(pp.furtureList, hash)
	}

	pp.view = newView
}

func (pp *PoCPool) AddFuturePoC(entry *PoCEntry) bool {
	pp.Lock()
	defer pp.Unlock()

	hash := entry.Param.Hash()
	if _, ok := pp.furtureList[hash]; ok {
		log.Infof("AddFuturePoC: poc params %x is already in future pool", hash)
		return false
	}
	pp.furtureList[hash] = entry

	return true
}

func (pp *PoCPool) RemoveFuturePoC(view uint32) []*PoCEntry {
	pp.Lock()
	defer pp.Unlock()

	list := []*PoCEntry{}
	entryToRemove := make(map[common.Uint256]int)

	for hash, entry := range pp.furtureList {
		if entry.Param.View == view {
			list = append(list, entry)
			entryToRemove[hash] = 1
		}
	}

	for hash, _ := range entryToRemove {
		delete(pp.furtureList, hash)
	}

	return list
}
