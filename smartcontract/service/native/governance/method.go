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

package governance

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/big"
	"sort"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/constants"
	"github.com/saveio/themis/common/log"
	cstates "github.com/saveio/themis/core/states"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func registerCandidate(native *native.NativeService, flag string) error {
	params := new(RegisterCandidateParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check auth of OntID
	err := appCallVerifyToken(native, contract, params.Caller, REGISTER_CANDIDATE, uint64(params.KeyNo))
	if err != nil {
		return fmt.Errorf("appCallVerifyToken, verifyToken failed: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return fmt.Errorf("getView, get view error: %v", err)
	}

	//check peerPubkey
	if err := validatePeerPubKeyFormat(params.PeerPubkey); err != nil {
		return fmt.Errorf("invalid peer pubkey")
	}

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	//get black list
	blackList, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return fmt.Errorf("native.CacheDB.Get, get BlackList error: %v", err)
	}
	if blackList != nil {
		return fmt.Errorf("registerCandidate, this Peer is in BlackList")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	//check if exist in PeerPool
	_, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if ok {
		return fmt.Errorf("registerCandidate, peerPubkey is already in peerPoolMap")
	}

	peerPoolItem := &PeerPoolItem{
		PeerPubkey: params.PeerPubkey,
		Address:    params.Address,
		InitPos:    uint64(params.InitPos),
		Status:     RegisterCandidateStatus,
		Goverance:  params.Goverance,
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	switch flag {
	case "transfer":
		//ont transfer
		err = appCallTransferOnt(native, params.Address, utils.GovernanceContractAddress, uint64(params.InitPos))
		if err != nil {
			return fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
		}

	case "transferFrom":
		//ont transfer from
		err = appCallTransferFromOnt(native, utils.GovernanceContractAddress, params.Address, utils.GovernanceContractAddress, uint64(params.InitPos))
		if err != nil {
			return fmt.Errorf("appCallTransferFromOnt, ont transfer error: %v", err)
		}

		// //ong transfer from
		// err = appCallTransferFromOng(native, utils.GovernanceContractAddress, params.Address, utils.GovernanceContractAddress, globalParam.CandidateFee)
		// if err != nil {
		// 	return fmt.Errorf("appCallTransferFromOng, ong transfer error: %v", err)
		// }
	}

	//update total stake
	err = depositTotalStake(native, contract, params.Address, uint64(params.InitPos))
	if err != nil {
		return fmt.Errorf("depositTotalStake, depositTotalStake error: %v", err)
	}
	return nil
}

func authorizeForPeer(native *native.NativeService, flag string) error {
	params := &AuthorizeForPeerParam{
		PeerPubkeyList: make([]string, 0),
		PosList:        make([]uint32, 0),
	}
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return fmt.Errorf("getView, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	//get globalParam2
	globalParam2, err := getGlobalParam2(native, contract)
	if err != nil {
		return fmt.Errorf("getGlobalParam2, getGlobalParam2 error: %v", err)
	}

	var total uint64
	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := params.PosList[i]

		//check pos
		if pos < 1 {
			return fmt.Errorf("authorizeForPeer, pos must >= 1")
		}
		if pos < globalParam2.MinAuthorizePos || pos%globalParam2.MinAuthorizePos != 0 {
			return fmt.Errorf("authorizeForPeer, pos must be times of %d", globalParam2.MinAuthorizePos)
		}

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return fmt.Errorf("authorizeForPeer, peerPubkey is not in peerPoolMap")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus {
			return fmt.Errorf("authorizeForPeer, peerPubkey is not candidate and can not be authorized")
		}

		//check if peer can receive authorization
		peerAttributes, err := getPeerAttributes(native, contract, peerPubkey)
		if err != nil {
			return fmt.Errorf("getPeerAttributes error: %v", err)
		}

		//check if address is peer owner address
		if params.Address == peerPoolItem.Address {
			return fmt.Errorf("authorizeForPeer error, address can not be node owner")
		}

		authorizeInfo, err := getAuthorizeInfo(native, contract, peerPubkey, params.Address)
		if err != nil {
			return fmt.Errorf("getAuthorizeInfo, get authorizeInfo error: %v", err)
		}
		authorizeInfo.NewPos = authorizeInfo.NewPos + uint64(pos)
		total = total + uint64(pos)
		peerPoolItem.TotalPos = peerPoolItem.TotalPos + uint64(pos)
		if peerPoolItem.TotalPos > uint64(globalParam.PosLimit)*peerPoolItem.InitPos {
			return fmt.Errorf("authorizeForPeer, pos of this peer is full")
		}
		if peerPoolItem.TotalPos > peerAttributes.MaxAuthorize {
			return fmt.Errorf("authorizeForPeer, pos of this peer is more than peerAttributes.MaxAuthorize")
		}

		peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
		}
	}
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	switch flag {
	case "transfer":
		//ont transfer
		err = appCallTransferOnt(native, params.Address, utils.GovernanceContractAddress, total)
		if err != nil {
			return fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
		}
	case "transferFrom":
		//ont transfer from
		err = appCallTransferFromOnt(native, utils.GovernanceContractAddress, params.Address, utils.GovernanceContractAddress, total)
		if err != nil {
			return fmt.Errorf("appCallTransferFromOnt, ont transfer error: %v", err)
		}
	}

	//update total stake
	err = depositTotalStake(native, contract, params.Address, total)
	if err != nil {
		return fmt.Errorf("depositTotalStake, depositTotalStake error: %v", err)
	}

	return nil
}

func normalQuit(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	flag := false

	//draw back authorize pos
	iter := native.CacheDB.NewIterator(utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	defer iter.Release()
	for has := iter.First(); has; has = iter.Next() {
		authorizeInfoStore, err := cstates.GetValueFromRawStorageItem(iter.Value())
		if err != nil {
			return fmt.Errorf("authorizeInfoStore is not available!:%v", err)
		}
		var authorizeInfo AuthorizeInfo
		if err := authorizeInfo.Deserialization(common.NewZeroCopySource(authorizeInfoStore)); err != nil {
			return fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.ConsensusPos + authorizeInfo.CandidatePos + authorizeInfo.NewPos +
			authorizeInfo.WithdrawConsensusPos + authorizeInfo.WithdrawCandidatePos + authorizeInfo.WithdrawUnfreezePos
		authorizeInfo.ConsensusPos = 0
		authorizeInfo.CandidatePos = 0
		authorizeInfo.NewPos = 0
		authorizeInfo.WithdrawConsensusPos = 0
		authorizeInfo.WithdrawCandidatePos = 0
		if authorizeInfo.Address == peerPoolItem.Address {
			flag = true
			authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + peerPoolItem.InitPos
		}
		err = putAuthorizeInfo(native, contract, &authorizeInfo)
		if err != nil {
			return fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
		}
	}
	if err := iter.Error(); err != nil {
		return err
	}
	if !flag {
		authorizeInfo := &AuthorizeInfo{
			PeerPubkey:          peerPoolItem.PeerPubkey,
			Address:             peerPoolItem.Address,
			WithdrawUnfreezePos: peerPoolItem.InitPos,
		}
		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
		}
	}
	return nil
}

func blackQuit(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	// ont transfer to trigger unboundong
	err := appCallTransferOnt(native, utils.GovernanceContractAddress, utils.GovernanceContractAddress, peerPoolItem.InitPos)
	if err != nil {
		return fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
	}

	//update total stake
	err = withdrawTotalStake(native, contract, peerPoolItem.Address, peerPoolItem.InitPos)
	if err != nil {
		return fmt.Errorf("withdrawTotalStake, withdrawTotalStake error: %v", err)
	}

	initPos := peerPoolItem.InitPos
	var authorizePos uint64

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	//draw back authorize pos
	iter := native.CacheDB.NewIterator(utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	defer iter.Release()
	for has := iter.First(); has; has = iter.Next() {
		authorizeInfoStore, err := cstates.GetValueFromRawStorageItem(iter.Value())
		if err != nil {
			return fmt.Errorf("authorizeInfoStore is not available!:%v", err)
		}
		var authorizeInfo AuthorizeInfo
		if err := authorizeInfo.Deserialization(common.NewZeroCopySource(authorizeInfoStore)); err != nil {
			return fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}
		total := authorizeInfo.ConsensusPos + authorizeInfo.CandidatePos + authorizeInfo.NewPos + authorizeInfo.WithdrawConsensusPos +
			authorizeInfo.WithdrawCandidatePos
		penalty := (uint64(globalParam.Penalty)*total + 99) / 100
		authorizeInfo.WithdrawUnfreezePos = total - penalty + authorizeInfo.WithdrawUnfreezePos
		authorizeInfo.ConsensusPos = 0
		authorizeInfo.CandidatePos = 0
		authorizeInfo.NewPos = 0
		authorizeInfo.WithdrawConsensusPos = 0
		authorizeInfo.WithdrawCandidatePos = 0
		address := authorizeInfo.Address
		err = putAuthorizeInfo(native, contract, &authorizeInfo)
		if err != nil {
			return fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
		}

		//update total stake
		err = withdrawTotalStake(native, contract, address, penalty)
		if err != nil {
			return fmt.Errorf("withdrawTotalStake, withdrawTotalStake error: %v", err)
		}
		authorizePos = authorizePos + penalty
	}
	if err := iter.Error(); err != nil {
		return err
	}

	//add penalty stake
	err = depositPenaltyStake(native, contract, peerPoolItem.PeerPubkey, initPos, authorizePos)
	if err != nil {
		return fmt.Errorf("depositPenaltyStake, deposit penaltyStake error: %v", err)
	}
	return nil
}

func consensusToConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}

	//update authorizeInfoPool
	iter := native.CacheDB.NewIterator(utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	defer iter.Release()
	for has := iter.First(); has; has = iter.Next() {
		authorizeInfoStore, err := cstates.GetValueFromRawStorageItem(iter.Value())
		if err != nil {
			return fmt.Errorf("authorizeInfoStore is not available!:%v", err)
		}
		var authorizeInfo AuthorizeInfo
		if err := authorizeInfo.Deserialization(common.NewZeroCopySource(authorizeInfoStore)); err != nil {
			return fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}
		if authorizeInfo.CandidatePos != 0 {
			return fmt.Errorf("commitPos, candidatePos should be 0")
		}

		//update status
		newPos := authorizeInfo.NewPos
		authorizeInfo.ConsensusPos = authorizeInfo.ConsensusPos + newPos
		authorizeInfo.NewPos = 0
		withdrawConsensusPos := authorizeInfo.WithdrawConsensusPos
		withdrawFCandidatePos := authorizeInfo.WithdrawCandidatePos
		authorizeInfo.WithdrawCandidatePos = withdrawConsensusPos
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + withdrawFCandidatePos
		authorizeInfo.WithdrawConsensusPos = 0

		err = putAuthorizeInfo(native, contract, &authorizeInfo)
		if err != nil {
			return fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
		}
	}
	if err := iter.Error(); err != nil {
		return err
	}
	return nil
}

func unConsensusToConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}

	//update authorizeInfoPool
	iter := native.CacheDB.NewIterator(utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	defer iter.Release()
	for has := iter.First(); has; has = iter.Next() {
		authorizeInfoStore, err := cstates.GetValueFromRawStorageItem(iter.Value())
		if err != nil {
			return fmt.Errorf("authorizeInfoStore is not available!:%v", err)
		}
		var authorizeInfo AuthorizeInfo
		if err := authorizeInfo.Deserialization(common.NewZeroCopySource(authorizeInfoStore)); err != nil {
			return fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}
		if authorizeInfo.ConsensusPos != 0 {
			return fmt.Errorf("commitPos, consensusPos should be 0")
		}

		//update status
		authorizeInfo.ConsensusPos = authorizeInfo.ConsensusPos + authorizeInfo.CandidatePos + authorizeInfo.NewPos
		authorizeInfo.NewPos = 0
		authorizeInfo.CandidatePos = 0
		withdrawConsensusPos := authorizeInfo.WithdrawConsensusPos
		withdrawCandidatePos := authorizeInfo.WithdrawCandidatePos
		authorizeInfo.WithdrawCandidatePos = withdrawConsensusPos
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + withdrawCandidatePos
		authorizeInfo.WithdrawConsensusPos = 0

		err = putAuthorizeInfo(native, contract, &authorizeInfo)
		if err != nil {
			return fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
		}
	}
	if err := iter.Error(); err != nil {
		return err
	}
	return nil
}

func consensusToUnConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}

	//update authorizeInfoPool
	iter := native.CacheDB.NewIterator(utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	defer iter.Release()
	for has := iter.First(); has; has = iter.Next() {
		authorizeInfoStore, err := cstates.GetValueFromRawStorageItem(iter.Value())
		if err != nil {
			return fmt.Errorf("authorizeInfoStore is not available!:%v", err)
		}
		var authorizeInfo AuthorizeInfo
		if err := authorizeInfo.Deserialization(common.NewZeroCopySource(authorizeInfoStore)); err != nil {
			return fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}
		if authorizeInfo.CandidatePos != 0 {
			return fmt.Errorf("commitPos, candidatePos should be 0")
		}

		//update status
		authorizeInfo.CandidatePos = authorizeInfo.ConsensusPos + authorizeInfo.NewPos
		authorizeInfo.NewPos = 0
		authorizeInfo.ConsensusPos = 0
		withdrawConsensusPos := authorizeInfo.WithdrawConsensusPos
		withdrawCandidatePos := authorizeInfo.WithdrawCandidatePos
		authorizeInfo.WithdrawCandidatePos = withdrawConsensusPos
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + withdrawCandidatePos
		authorizeInfo.WithdrawConsensusPos = 0

		err = putAuthorizeInfo(native, contract, &authorizeInfo)
		if err != nil {
			return fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
		}
	}
	if err := iter.Error(); err != nil {
		return err
	}
	return nil
}

func unConsensusToUnConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}

	//update authorizeInfoPool
	iter := native.CacheDB.NewIterator(utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	defer iter.Release()
	for has := iter.First(); has; has = iter.Next() {
		authorizeInfoStore, err := cstates.GetValueFromRawStorageItem(iter.Value())
		if err != nil {
			return fmt.Errorf("authorizeInfoStore is not available!:%v", err)
		}
		var authorizeInfo AuthorizeInfo
		if err := authorizeInfo.Deserialization(common.NewZeroCopySource(authorizeInfoStore)); err != nil {
			return fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}
		if authorizeInfo.ConsensusPos != 0 {
			return fmt.Errorf("commitPos, consensusPos should be 0")
		}

		//update status
		newPos := authorizeInfo.NewPos
		candidatePos := authorizeInfo.CandidatePos
		authorizeInfo.NewPos = 0
		authorizeInfo.CandidatePos = newPos + candidatePos
		withdrawConsensusPos := authorizeInfo.WithdrawConsensusPos
		withdrawCandidatePos := authorizeInfo.WithdrawCandidatePos
		authorizeInfo.WithdrawCandidatePos = withdrawConsensusPos
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + withdrawCandidatePos
		authorizeInfo.WithdrawConsensusPos = 0

		err = putAuthorizeInfo(native, contract, &authorizeInfo)
		if err != nil {
			return fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
		}
	}
	if err := iter.Error(); err != nil {
		return err
	}
	return nil
}

func depositTotalStake(native *native.NativeService, contract common.Address, address common.Address, stake uint64) error {
	totalStake, err := getTotalStake(native, contract, address)
	if err != nil {
		return fmt.Errorf("getTotalStake, get totalStake error: %v", err)
	}

	preStake := totalStake.Stake
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	totalStake.Stake = preStake + stake
	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, contract, totalStake)
	if err != nil {
		return fmt.Errorf("putTotalStake, put totalStake error: %v", err)
	}
	return nil
}

func withdrawTotalStake(native *native.NativeService, contract common.Address, address common.Address, stake uint64) error {
	totalStake, err := getTotalStake(native, contract, address)
	if err != nil {
		return fmt.Errorf("getTotalStake, get totalStake error: %v", err)
	}
	if totalStake.Stake < stake {
		return fmt.Errorf("withdraw, ont deposit is not enough")
	}

	preStake := totalStake.Stake
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	totalStake.Stake = preStake - stake
	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, contract, totalStake)
	if err != nil {
		return fmt.Errorf("putTotalStake, put totalStake error: %v", err)
	}
	return nil
}

func depositPenaltyStake(native *native.NativeService, contract common.Address, peerPubkey string, initPos uint64, authorizePos uint64) error {
	penaltyStake, err := getPenaltyStake(native, contract, peerPubkey)
	if err != nil {
		return fmt.Errorf("getPenaltyStake, get penaltyStake error: %v", err)
	}

	preInitPos := penaltyStake.InitPos
	preAuthorizePos := penaltyStake.AuthorizePos
	// preStake := preInitPos + preAuthorizePos
	// preTimeOffset := penaltyStake.TimeOffset
	preAmount := penaltyStake.Amount
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	// amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)

	// penaltyStake.Amount = preAmount + amount
	penaltyStake.Amount = preAmount
	penaltyStake.InitPos = preInitPos + initPos
	penaltyStake.AuthorizePos = preAuthorizePos + authorizePos
	penaltyStake.TimeOffset = timeOffset

	err = putPenaltyStake(native, contract, penaltyStake)
	if err != nil {
		return fmt.Errorf("putPenaltyStake, put penaltyStake error: %v", err)
	}
	return nil
}

func withdrawPenaltyStake(native *native.NativeService, contract common.Address, peerPubkey string, address common.Address) error {
	penaltyStake, err := getPenaltyStake(native, contract, peerPubkey)
	if err != nil {
		return fmt.Errorf("getPenaltyStake, get penaltyStake error: %v", err)
	}

	preStake := penaltyStake.InitPos + penaltyStake.AuthorizePos
	// preTimeOffset := penaltyStake.TimeOffset
	// preAmount := penaltyStake.Amount
	// timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	// amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)

	//ont transfer
	err = appCallTransferOnt(native, utils.GovernanceContractAddress, address, preStake)
	if err != nil {
		return fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
	}
	// //ong approve
	// err = appCallTransferFromOng(native, utils.GovernanceContractAddress, utils.UsdtContractAddress, address, amount+preAmount)
	// if err != nil {
	// 	return fmt.Errorf("appCallTransferFromOng, transfer from ong error: %v", err)
	// }

	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(contract, []byte(PENALTY_STAKE), peerPubkeyPrefix))
	return nil
}

func executeCommitDpos(native *native.NativeService, contract common.Address) error {
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return fmt.Errorf("getGovernanceView, get GovernanceView error: %v", err)
	}
	if native.Height == governanceView.Height {
		return fmt.Errorf("commitDpos, can not do commitDpos twice in one block")
	}
	//get current view
	view := governanceView.View

	if view <= NEW_VERSION_VIEW {
		//err = executeCommitDpos1(native, contract)
		err = executeCommitDposEnhance(native, contract)
		if err != nil {
			return fmt.Errorf("executeCommitDpos1 error: %v", err)
		}
	} else {
		//err = executeCommitDpos2(native, contract)
		err = executeCommitDposEnhance(native, contract)
		if err != nil {
			return fmt.Errorf("executeCommitDpos2 error: %v", err)
		}
	}

	//update view
	governanceView = &GovernanceView{
		View:   view + 1,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	err = putGovernanceView(native, contract, governanceView)
	if err != nil {
		return fmt.Errorf("putGovernanceView, put governanceView error: %v", err)
	}

	//update cons gov view
	consGovView, err := GetConsGovView(native, contract)
	if err != nil {
		return fmt.Errorf("executeCommitDpos, get cons gov view error: %v", err)
	}

	if IsLastViewInConsGovPeriod(consGovView, view) {
		runningStartView := consGovView.ReelectStartView + NUM_VIEW_PER_CONS_PLEDGE_PERIOD + NUM_VIEW_PER_CONS_ELECT_PERIOD
		runningStartView += NUM_VIEW_PER_GOV_ELECT_PERIOD
		reelectStartView := runningStartView + NUM_VIEW_PER_RUNNING_PERIOD

		newConsGovView := &ConsGovView{
			GovView:          consGovView.GovView + 1,
			RunningStartView: runningStartView,
			ReelectStartView: reelectStartView,
		}
		err = putConsGovView(native, contract, newConsGovView)
		if err != nil {
			return fmt.Errorf("putGovernanceView, put cons gov view error: %v", err)
		}

		log.Debugf("executeCommitDpos update cons gov view %v\n", newConsGovView)
	}

	return nil
}

func executeSplitEnhance(native *native.NativeService, contract common.Address, view uint32) error {
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getConfig, get config error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view-1)
	if err != nil {
		return fmt.Errorf("executeSplit, get peerPoolMap error: %v", err)
	}

	gasRevenue, err := getGasRevenue(native, contract)
	if err != nil {
		return fmt.Errorf("executeSplit, getGasRevenue error: %v", err)
	}
	balance := gasRevenue.Total
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	peersCandidate := []*CandidateSplitInfo{}

	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.ConsStatus&InConsensusGroup == 0 {
			continue
		}

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peersCandidate = append(peersCandidate, &CandidateSplitInfo{
				PeerPubkey: peerPoolItem.PeerPubkey,
				InitPos:    peerPoolItem.InitPos,
				Address:    peerPoolItem.Address,
				Status:     peerPoolItem.Status,
				Stake:      stake,
			})
		}
	}

	// sort peers by state and stake. ConsensusStatus > CandidateStatus
	sort.SliceStable(peersCandidate, func(i, j int) bool {
		if peersCandidate[i].Status > peersCandidate[j].Status {
			return true
		} else if peersCandidate[i].Stake > peersCandidate[j].Stake {
			return true
		} else if peersCandidate[i].Stake == peersCandidate[j].Stake {
			return peersCandidate[i].PeerPubkey > peersCandidate[j].PeerPubkey
		}
		return false
	})

	// cal s of each consensus node
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	// if sum = 0, means consensus peer in config
	if sum < uint64(config.K) {
		//use same share for peer in config
		sum = uint64(config.K)
	}

	/*
		//[PoC] just use deposit without Curve
		for i := 0; i < int(config.K); i++ {
			nodeAmount := balance * uint64(globalParam.A) / 100 * peersCandidate[i].Stake / sum
			address := peersCandidate[i].Address
			log.Debugf("executeSplitEnhance send %d bonus to consensus node: %s", nodeAmount, address.ToBase58())

			err = appCallTransferRevenue(native, utils.GovernanceContractAddress, address, nodeAmount)
			if err != nil {
				return fmt.Errorf("executeSplitEnhance, bonus transfer error: %v", err)
			}
		}
	*/

	//[PoC] use deposit and Curve
	avg := sum / uint64(config.K)
	var sumS uint64
	for i := 0; i < int(config.K); i++ {
		//use same share when sum is 0
		if sum == 0 {
			peersCandidate[i].S, err = splitCurve(native, contract, 1, avg, uint64(globalParam.Yita))
			if err != nil {
				return fmt.Errorf("splitCurve, calculate splitCurve error: %v", err)
			}
		} else {
			peersCandidate[i].S, err = splitCurve(native, contract, peersCandidate[i].Stake, avg, uint64(globalParam.Yita))
			if err != nil {
				return fmt.Errorf("splitCurve, calculate splitCurve error: %v", err)
			}
		}
		sumS += peersCandidate[i].S
	}
	if sumS == 0 {
		return fmt.Errorf("executeSplitEnhance, sumS is 0")
	}

	//fee split of consensus peer
	for i := 0; i < int(config.K); i++ {
		nodeAmount := balance * uint64(globalParam.A) / 100 * peersCandidate[i].S / sumS
		address := peersCandidate[i].Address

		log.Debugf("executeSplitEnhance send %d bonus to consensus node: %s", nodeAmount, address.ToBase58())

		err = appCallTransferRevenue(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return fmt.Errorf("executeSplitEnhance, bonus transfer error: %v", err)
		}
	}

	//fee split of candidate peer
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	if sum == 0 {
		return nil
	}
	for i := int(config.K); i < len(peersCandidate); i++ {
		nodeAmount := balance * uint64(globalParam.B) / 100 * peersCandidate[i].Stake / sum
		address := peersCandidate[i].Address

		log.Debugf("executeSplitEnhance send %d bonus to candidate node: %s", nodeAmount, address.ToBase58())

		err = appCallTransferRevenue(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return fmt.Errorf("executeSplitEnhance, bonus transfer error: %v", err)
		}
	}

	return nil
}

func executeSplit(native *native.NativeService, contract common.Address, view uint32) error {
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getConfig, get config error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view-1)
	if err != nil {
		return fmt.Errorf("executeSplit, get peerPoolMap error: %v", err)
	}

	//balance, err := getOngBalance(native, utils.GovernanceContractAddress)
	gasRevenue, err := getGasRevenue(native, contract)
	if err != nil {
		return fmt.Errorf("executeSplit, getGasRevenue error: %v", err)
	}
	balance := gasRevenue.Total
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	peersCandidate := []*CandidateSplitInfo{}

	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peersCandidate = append(peersCandidate, &CandidateSplitInfo{
				PeerPubkey: peerPoolItem.PeerPubkey,
				InitPos:    peerPoolItem.InitPos,
				Address:    peerPoolItem.Address,
				Stake:      stake,
			})
		}
	}

	// sort peers by stake
	sort.SliceStable(peersCandidate, func(i, j int) bool {
		if peersCandidate[i].Stake > peersCandidate[j].Stake {
			return true
		} else if peersCandidate[i].Stake == peersCandidate[j].Stake {
			return peersCandidate[i].PeerPubkey > peersCandidate[j].PeerPubkey
		}
		return false
	})

	// cal s of each consensus node
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	// if sum = 0, means consensus peer in config, do not split
	if sum < uint64(config.K) {
		//use same share for peer in config
		sum = uint64(config.K)
	}

	avg := sum / uint64(config.K)
	var sumS uint64
	for i := 0; i < int(config.K); i++ {
		//use same share when sum is 0
		if sum == 0 {
			peersCandidate[i].S, err = splitCurve(native, contract, 1, avg, uint64(globalParam.Yita))
			if err != nil {
				return fmt.Errorf("splitCurve, calculate splitCurve error: %v", err)
			}
		} else {
			peersCandidate[i].S, err = splitCurve(native, contract, peersCandidate[i].Stake, avg, uint64(globalParam.Yita))
			if err != nil {
				return fmt.Errorf("splitCurve, calculate splitCurve error: %v", err)
			}
		}
		sumS += peersCandidate[i].S
	}
	if sumS == 0 {
		return fmt.Errorf("executeSplit, sumS is 0")
	}

	//fee split of consensus peer
	for i := 0; i < int(config.K); i++ {
		nodeAmount := balance * uint64(globalParam.A) / 100 * peersCandidate[i].S / sumS
		address := peersCandidate[i].Address
		//err = appCallTransferOng(native, utils.GovernanceContractAddress, address, nodeAmount)
		err = appCallTransferRevenue(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return fmt.Errorf("executeSplit, ong transfer error: %v", err)
		}
	}

	//fee split of candidate peer
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	if sum == 0 {
		return nil
	}
	for i := int(config.K); i < len(peersCandidate); i++ {
		nodeAmount := balance * uint64(globalParam.B) / 100 * peersCandidate[i].Stake / sum
		address := peersCandidate[i].Address
		err = appCallTransferRevenue(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return fmt.Errorf("executeSplit, ong transfer error: %v", err)
		}
	}

	return nil
}

func executeSplit2(native *native.NativeService, contract common.Address, view uint32) (uint64, error) {
	var splitSum uint64 = 0
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return splitSum, fmt.Errorf("getConfig, get config error: %v", err)
	}

	//get globalParam2
	globalParam2, err := getGlobalParam2(native, contract)
	if err != nil {
		return splitSum, fmt.Errorf("getGlobalParam2, getGlobalParam2 error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view-1)
	if err != nil {
		return splitSum, fmt.Errorf("executeSplit, get peerPoolMap error: %v", err)
	}

	//get current peerPoolMap
	currentPeerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return splitSum, fmt.Errorf("executeSplit, get currentPeerPoolMap error: %v", err)
	}

	balance, err := getUsdtBalance(native, utils.GovernanceContractAddress)
	if err != nil {
		return splitSum, fmt.Errorf("executeSplit, getUsdtBalance error: %v", err)
	}
	splitFee, err := getSplitFee(native, contract)
	if err != nil {
		return splitSum, fmt.Errorf("getSplitFee, getSplitFee error: %v", err)
	}
	if balance < splitFee {
		panic("balance less than splitFee to withdraw!")
	}
	income := balance - splitFee

	//fee split to dapp address
	dappIncome := new(big.Int).Div(new(big.Int).Mul(new(big.Int).SetUint64(income),
		new(big.Int).SetUint64(uint64(globalParam2.DappFee))), new(big.Int).SetUint64(100))
	gasAddress, err := getGasAddress(native, contract)
	if err != nil {
		return splitSum, fmt.Errorf("getGasAddress, getGasAddress error: %v", err)
	}
	if gasAddress.Address == common.ADDRESS_EMPTY {
		dappIncome = new(big.Int).SetUint64(0)
	} else {
		// err := appCallTransferOng(native, utils.GovernanceContractAddress, gasAddress.Address, dappIncome.Uint64())
		// if err != nil {
		// 	return splitSum, fmt.Errorf("appCallTransferOng, appCallTransferOng error: %v", err)
		// }
	}

	//fee split to node
	if income < dappIncome.Uint64() {
		panic("income less than dappIncome!")
	}
	nodeIncome := new(big.Int).Sub(new(big.Int).SetUint64(income), dappIncome)
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return splitSum, fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	peersCandidate := []*CandidateSplitInfo{}

	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peersCandidate = append(peersCandidate, &CandidateSplitInfo{
				PeerPubkey: peerPoolItem.PeerPubkey,
				InitPos:    peerPoolItem.InitPos,
				Address:    peerPoolItem.Address,
				Stake:      stake,
			})
		}
	}

	// sort peers by stake
	sort.SliceStable(peersCandidate, func(i, j int) bool {
		if peersCandidate[i].Stake > peersCandidate[j].Stake {
			return true
		} else if peersCandidate[i].Stake == peersCandidate[j].Stake {
			return peersCandidate[i].PeerPubkey > peersCandidate[j].PeerPubkey
		}
		return false
	})

	// cal s of each consensus node
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	// if sum = 0, means consensus peer in config, do not split
	if sum < uint64(config.K) {
		return splitSum, nil
	}
	avg := sum / uint64(config.K)
	var sumS uint64
	for i := 0; i < int(config.K); i++ {
		peersCandidate[i].S, err = splitCurve(native, contract, peersCandidate[i].Stake, avg, uint64(globalParam.Yita))
		if err != nil {
			return splitSum, fmt.Errorf("splitCurve, calculate splitCurve error: %v", err)
		}
		sumS += peersCandidate[i].S
	}
	if sumS == 0 {
		return splitSum, fmt.Errorf("executeSplit, sumS is 0")
	}

	//fee split of consensus peer
	for i := 0; i < int(config.K); i++ {
		//nodeAmount := nodeIncome * uint64(globalParam.A) / 100 * peersCandidate[i].S / sumS
		//consensusWeight := nodeIncome * uint64(globalParam.A)
		//consensusAmount := consensusWeight / 100
		//nodeWeight := consensusAmount * peersCandidate[i].S
		//nodeAmount := nodeWeight / sumS
		consensusWeight := new(big.Int).Mul(nodeIncome, new(big.Int).SetUint64(uint64(globalParam.A)))
		consensusAmount := new(big.Int).Div(consensusWeight, new(big.Int).SetUint64(100))
		nodeWeight := new(big.Int).Mul(consensusAmount, new(big.Int).SetUint64(peersCandidate[i].S))
		nodeAmount := new(big.Int).Div(nodeWeight, new(big.Int).SetUint64(sumS))

		err = splitNodeFee(native, contract, peersCandidate[i].PeerPubkey, peersCandidate[i].Address,
			peerPoolMap.PeerPoolMap[peersCandidate[i].PeerPubkey].Status == ConsensusStatus,
			currentPeerPoolMap.PeerPoolMap[peersCandidate[i].PeerPubkey].Status == ConsensusStatus,
			peerPoolMap.PeerPoolMap[peersCandidate[i].PeerPubkey].InitPos, peerPoolMap.PeerPoolMap[peersCandidate[i].PeerPubkey].TotalPos, nodeAmount.Uint64())
		if err != nil {
			return splitSum, fmt.Errorf("executeSplit2, splitNodeFee error: %v", err)
		}
		splitSum += nodeAmount.Uint64()
	}

	//fee split of candidate peer
	//cal s of each candidate node
	var length int
	if int(globalParam2.CandidateFeeSplitNum) >= len(peersCandidate) {
		length = len(peersCandidate)
	} else {
		length = int(globalParam2.CandidateFeeSplitNum)
	}
	sum = 0
	for i := int(config.K); i < length; i++ {
		sum += peersCandidate[i].Stake
	}
	if sum == 0 {
		return splitSum, nil
	}
	for i := int(config.K); i < length; i++ {
		//nodeAmount := nodeIncome * uint64(globalParam.B) / 100 * peersCandidate[i].Stake / sum
		//candidateWeight := nodeIncome * uint64(globalParam.B)
		//candidateAmount := candidateWeight / 100
		//nodeWeight := candidateAmount * peersCandidate[i].Stake
		//nodeAmount := nodeWeight / sum
		candidateWeight := new(big.Int).Mul(nodeIncome, new(big.Int).SetUint64(uint64(globalParam.B)))
		candidateAmount := new(big.Int).Div(candidateWeight, new(big.Int).SetUint64(100))
		nodeWeight := new(big.Int).Mul(candidateAmount, new(big.Int).SetUint64(peersCandidate[i].Stake))
		nodeAmount := new(big.Int).Div(nodeWeight, new(big.Int).SetUint64(sum))

		err = splitNodeFee(native, contract, peersCandidate[i].PeerPubkey, peersCandidate[i].Address,
			peerPoolMap.PeerPoolMap[peersCandidate[i].PeerPubkey].Status == ConsensusStatus,
			currentPeerPoolMap.PeerPoolMap[peersCandidate[i].PeerPubkey].Status == ConsensusStatus,
			peerPoolMap.PeerPoolMap[peersCandidate[i].PeerPubkey].InitPos, peerPoolMap.PeerPoolMap[peersCandidate[i].PeerPubkey].TotalPos, nodeAmount.Uint64())
		if err != nil {
			return splitSum, fmt.Errorf("executeSplit2, splitNodeFee error: %v", err)
		}
		splitSum += nodeAmount.Uint64()
	}

	return splitSum, nil
}

func executeAddressSplit(native *native.NativeService, contract common.Address, authorizeInfo *AuthorizeInfo, preIfConsensus, ifConsensus bool, totalPos uint64, totalAmount uint64, peerAddress common.Address) (uint64, error) {
	var validatePos uint64
	if ifConsensus || preIfConsensus {
		validatePos = authorizeInfo.ConsensusPos + authorizeInfo.WithdrawConsensusPos
	} else {
		validatePos = authorizeInfo.CandidatePos + authorizeInfo.WithdrawCandidatePos
	}

	if validatePos == 0 || authorizeInfo.Address == peerAddress {
		return 0, nil
	}
	amount := validatePos * totalAmount / totalPos
	splitFeeAddress, err := getSplitFeeAddress(native, contract, authorizeInfo.Address)
	if err != nil {
		return 0, fmt.Errorf("getSplitFeeAddress, getSplitFeeAddress error: %v", err)
	}
	splitFeeAddress.Amount = splitFeeAddress.Amount + amount
	err = putSplitFeeAddress(native, contract, authorizeInfo.Address, splitFeeAddress)
	if err != nil {
		return 0, fmt.Errorf("putSplitFeeAddress, putSplitFeeAddress error: %v", err)
	}
	return amount, nil
}

func executePeerSplit(native *native.NativeService, contract common.Address, peerAddress common.Address, totalAmount uint64) error {
	splitFeeAddress, err := getSplitFeeAddress(native, contract, peerAddress)
	if err != nil {
		return fmt.Errorf("getSplitFeeAddress, getSplitFeeAddress error: %v", err)
	}
	splitFeeAddress.Amount = splitFeeAddress.Amount + totalAmount
	err = putSplitFeeAddress(native, contract, peerAddress, splitFeeAddress)
	if err != nil {
		return fmt.Errorf("putSplitFeeAddress, putSplitFeeAddress error: %v", err)
	}
	return nil
}

//use consensus gov view when manage peer map
func executeCommitDposEnhance(native *native.NativeService, contract common.Address) error {
	//get governace view
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return fmt.Errorf("getGovernanceView, get GovernanceView error: %v", err)
	}

	//get current consensus gov view
	view := governanceView.View
	newView := view + 1

	//feeSplit first
	err = executeSplitEnhance(native, contract, view)
	if err != nil {
		return fmt.Errorf("executeCommitDposEnhance, executeSplitEnhance error: %v", err)
	}

	//update config
	preConfig, err := getPreConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getPreConfig, get preConfig error: %v", err)
	}
	if preConfig.SetView == view {
		err = putConfig(native, contract, preConfig.Configuration)
		if err != nil {
			return fmt.Errorf("putConfig, put config error: %v", err)
		}
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	var peers []*PeerStakeInfo
	var peersSkip []*PeerStakeInfo
	var items *ConsGroupItems

	govView, err := GetConsGovView(native, contract)
	if err != nil {
		return fmt.Errorf("GetConsGovView error: %v", err)
	}

	last := IsLastViewInConsGovPeriod(govView, view)
	if last {
		//ConsGroupItems will ensure nodes > consensus group size
		items, err = GetConsGroupItems(native, contract, govView.GovView+1)

		if err != nil {
			return fmt.Errorf("GetConsGroupItems, get cons group items error: %v", err)
		}

		for pubkey, _ := range items.ConsGroupItems {
			log.Debugf("executeCommitDposEnhance GetConsGroupItems pubkey:%s", pubkey)
		}
	}

	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == QuitingStatus {
			err = normalQuit(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("normalQuit, normalQuit error: %v", err)
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == BlackStatus {
			err = blackQuit(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("blackQuit, blackQuit error: %v", err)
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == QuitConsensusStatus {
			peerPoolItem.Status = QuitingStatus
			peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
		}

		//only nodes in cons group of next gov view envolve
		if last {
			if _, ok := items.ConsGroupItems[peerPoolItem.PeerPubkey]; !ok {
				if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
					stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
					peersSkip = append(peersSkip, &PeerStakeInfo{
						Index:      peerPoolItem.Index,
						PeerPubkey: peerPoolItem.PeerPubkey,
						Stake:      stake,
					})
				}
				continue
			}
		} else {
			//only item in consensus group of current view take part in consensus switch.
			if peerPoolItem.ConsStatus&InConsensusGroup == 0 {
				log.Debugf("executeCommitDposEnhance item with index %d not in consensus group, skip", peerPoolItem.Index)
				continue
			}
		}

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peers = append(peers, &PeerStakeInfo{
				Index:      peerPoolItem.Index,
				PeerPubkey: peerPoolItem.PeerPubkey,
				Stake:      stake,
			})
		}
	}
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getConfig, get config error: %v", err)
	}
	if len(peers) < int(config.K) {
		return fmt.Errorf("commitDpos, num of peers is less than K")
	}

	// sort peers by stake then pubkey
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].Stake > peers[j].Stake {
			return true
		} else if peers[i].Stake == peers[j].Stake {
			return peers[i].PeerPubkey > peers[j].PeerPubkey
		}
		return false
	})

	//shuffle with latest poc winner info
	log.Debugf("executeCommitDposEnhance finish sorting using deposit info of view:%d", view)
	for i := 0; i < len(peers); i++ {
		log.Debugf("peer[%d] Index %d", i, peers[i].Index)
	}

	//Use dummy winnerInfo, since winner doesn't come from PoC param any more!
	//Try use winner of view-1 later?
	//winnerInfo, err := getWinnerInfo(native, contract, view)
	//if err != nil {
	//	return fmt.Errorf("executeCommitDposEnhance, get winnerInfo for view %d error: %v", view, err)
	//}
	winnerInfo := &WinnerInfo{View: view}

	bf := new(bytes.Buffer)
	if err = winnerInfo.Serialize(bf); err != nil {
		return fmt.Errorf("executeCommitDposEnhance, serialize miningInfo error: %v", err)
	}
	info := bf.Bytes()

	hash := fnv.New64a()

	for i := len(peers) - 1; i > 0; i-- {
		data, err := json.Marshal(struct {
			Info  []byte `json:"info"`
			View  uint32 `json:"view"`
			Index int    `json:"index"`
		}{info, view, i})
		if err != nil {
			return fmt.Errorf("executeCommitDposEnhance, generate random num error: %v", err)
		}

		hash.Reset()
		hash.Write(data)
		h := hash.Sum64()
		j := h % uint64(i)
		peers[i], peers[j] = peers[j], peers[i]
	}
	log.Debugf("executeCommitDposEnhance finish shuffle using poc winner info of view:%d", view)

	for i := 0; i < len(peers); i++ {
		log.Debugf("peer[%d] Index %d\n", i, peers[i].Index)
	}

	// consensus peers
	for i := 0; i < int(config.K); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return fmt.Errorf("executeCommitDposEnhance, peerPubkey is not in peerPoolMap")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("consensusToConsensus, consensusToConsensus error: %v", err)
			}
		} else {
			err = unConsensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("unConsensusToConsensus, unConsensusToConsensus error: %v", err)
			}
		}
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return fmt.Errorf("authorizeForPeer, peerPubkey is not in peerPoolMap")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("consensusToUnConsensus, consensusToUnConsensus error: %v", err)
			}
		} else {
			err = unConsensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("unConsensusToUnConsensus, unConsensusToUnConsensus error: %v", err)
			}
		}
		peerPoolItem.Status = CandidateStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}

	//non consensus peers due to switch consensus gov period
	if last {
		for i := 0; i < len(peersSkip); i++ {
			peerPoolItem, ok := peerPoolMap.PeerPoolMap[peersSkip[i].PeerPubkey]
			if !ok {
				return fmt.Errorf("authorizeForPeer, peerPubkey is not in peerPoolMap")
			}

			if peerPoolItem.Status == ConsensusStatus {
				err = consensusToUnConsensus(native, contract, peerPoolItem)
				if err != nil {
					return fmt.Errorf("consensusToUnConsensus, consensusToUnConsensus error: %v", err)
				}
			} else {
				err = unConsensusToUnConsensus(native, contract, peerPoolItem)
				if err != nil {
					return fmt.Errorf("unConsensusToUnConsensus, unConsensusToUnConsensus error: %v", err)
				}
			}
			peerPoolItem.Status = CandidateStatus
			peerPoolMap.PeerPoolMap[peersSkip[i].PeerPubkey] = peerPoolItem
		}
	}

	//1. adjust the peerMap for last view of cons gov period
	//2. write peerMap for view
	//3. prepare cons map for next gov view
	if last {
		for pubKey, peer := range peerPoolMap.PeerPoolMap {
			if _, ok := items.ConsGroupItems[pubKey]; ok {
				peer.ConsStatus = InConsensusGroup
				log.Debugf("executeCommitDposEnhance set InConsensusGroup for:%s", pubKey)

			} else {
				peer.ConsStatus = EmptyStatus
				log.Debugf("executeCommitDposEnhance clear InConsensusGroup for:%s", pubKey)
			}
		}

		for pubKey, peer := range peerPoolMap.PeerPoolMap {
			log.Debugf("executeCommitDposEnhance pubkey:%s, ConsStatus:%d", pubKey, peer.ConsStatus)
		}
	}

	//overwrite peerMap of view to make ExecuteSplitEnhance easy
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	err = putPeerPoolMap(native, contract, newView, peerPoolMap)
	if err != nil {
		return fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	//need keep old map for dispatch delayed mining bonus
	//oldView := view - 1
	//oldViewBytes := GetUint32Bytes(oldView)
	//native.CacheDB.Delete(utils.ConcatKey(contract, []byte(PEER_POOL), oldViewBytes))

	return nil
}

func executeCommitDpos1(native *native.NativeService, contract common.Address) error {
	//get governace view
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return fmt.Errorf("getGovernanceView, get GovernanceView error: %v", err)
	}

	//get current view
	view := governanceView.View
	newView := view + 1

	//feeSplit first
	err = executeSplit(native, contract, view)
	if err != nil {
		return fmt.Errorf("executeSplit, executeSplit error: %v", err)
	}

	//update config
	preConfig, err := getPreConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getPreConfig, get preConfig error: %v", err)
	}
	if preConfig.SetView == view {
		err = putConfig(native, contract, preConfig.Configuration)
		if err != nil {
			return fmt.Errorf("putConfig, put config error: %v", err)
		}
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	var peers []*PeerStakeInfo
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == QuitingStatus {
			err = normalQuit(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("normalQuit, normalQuit error: %v", err)
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == BlackStatus {
			err = blackQuit(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("blackQuit, blackQuit error: %v", err)
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == QuitConsensusStatus {
			peerPoolItem.Status = QuitingStatus
			peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
		}

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peers = append(peers, &PeerStakeInfo{
				Index:      peerPoolItem.Index,
				PeerPubkey: peerPoolItem.PeerPubkey,
				Stake:      stake,
			})
		}
	}
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getConfig, get config error: %v", err)
	}
	if len(peers) < int(config.K) {
		return fmt.Errorf("commitDpos, num of peers is less than K")
	}

	// sort peers by stake
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].Stake > peers[j].Stake {
			return true
		} else if peers[i].Stake == peers[j].Stake {
			return peers[i].PeerPubkey > peers[j].PeerPubkey
		}
		return false
	})

	// consensus peers
	for i := 0; i < int(config.K); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return fmt.Errorf("commitDpos, peerPubkey is not in peerPoolMap")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("consensusToConsensus, consensusToConsensus error: %v", err)
			}
		} else {
			err = unConsensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("unConsensusToConsensus, unConsensusToConsensus error: %v", err)
			}
		}
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return fmt.Errorf("authorizeForPeer, peerPubkey is not in peerPoolMap")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("consensusToUnConsensus, consensusToUnConsensus error: %v", err)
			}
		} else {
			err = unConsensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("unConsensusToUnConsensus, unConsensusToUnConsensus error: %v", err)
			}
		}
		peerPoolItem.Status = CandidateStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}
	err = putPeerPoolMap(native, contract, newView, peerPoolMap)
	if err != nil {
		return fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}
	oldView := view - 1
	oldViewBytes := GetUint32Bytes(oldView)
	native.CacheDB.Delete(utils.ConcatKey(contract, []byte(PEER_POOL), oldViewBytes))

	return nil
}

func executeCommitDpos2(native *native.NativeService, contract common.Address) error {
	//get governace view
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return fmt.Errorf("getGovernanceView, get GovernanceView error: %v", err)
	}

	//get current view
	view := governanceView.View
	newView := view + 1

	//feeSplit first
	splitSum, err := executeSplit2(native, contract, view)
	if err != nil {
		return fmt.Errorf("executeNodeSplit2, executeNodeSplit2 error: %v", err)
	}

	//update config
	preConfig, err := getPreConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getPreConfig, get preConfig error: %v", err)
	}
	if preConfig.SetView == view {
		err = putConfig(native, contract, preConfig.Configuration)
		if err != nil {
			return fmt.Errorf("putConfig, put config error: %v", err)
		}
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	var peers []*PeerStakeInfo
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		//update peer attributes
		peerAttributes, err := getPeerAttributes(native, contract, peerPoolItem.PeerPubkey)
		if err != nil {
			return fmt.Errorf("getPeerAttributes error: %v", err)
		}
		peerAttributes.TPeerCost = peerAttributes.T1PeerCost
		peerAttributes.T1PeerCost = peerAttributes.T2PeerCost
		peerAttributes.TStakeCost = peerAttributes.T1StakeCost
		peerAttributes.T1StakeCost = peerAttributes.T2StakeCost
		err = putPeerAttributes(native, contract, peerAttributes)
		if err != nil {
			return fmt.Errorf("putPeerAttributes error: %v", err)
		}

		//update peer status
		if peerPoolItem.Status == QuitingStatus {
			err = normalQuit(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("normalQuit, normalQuit error: %v", err)
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == BlackStatus {
			err = blackQuit(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("blackQuit, blackQuit error: %v", err)
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == QuitConsensusStatus {
			peerPoolItem.Status = QuitingStatus
			peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
		}

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peers = append(peers, &PeerStakeInfo{
				Index:      peerPoolItem.Index,
				PeerPubkey: peerPoolItem.PeerPubkey,
				Stake:      stake,
			})
		}
	}
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getConfig, get config error: %v", err)
	}
	if len(peers) < int(config.K) {
		return fmt.Errorf("commitDpos, num of peers is less than K")
	}

	// sort peers by stake
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].Stake > peers[j].Stake {
			return true
		} else if peers[i].Stake == peers[j].Stake {
			return peers[i].PeerPubkey > peers[j].PeerPubkey
		}
		return false
	})

	// consensus peers
	for i := 0; i < int(config.K); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return fmt.Errorf("commitDpos, peerPubkey is not in peerPoolMap")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("consensusToConsensus, consensusToConsensus error: %v", err)
			}
		} else {
			err = unConsensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("unConsensusToConsensus, unConsensusToConsensus error: %v", err)
			}
		}
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return fmt.Errorf("authorizeForPeer, peerPubkey is not in peerPoolMap")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("consensusToUnConsensus, consensusToUnConsensus error: %v", err)
			}
		} else {
			err = unConsensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return fmt.Errorf("unConsensusToUnConsensus, unConsensusToUnConsensus error: %v", err)
			}
		}
		peerPoolItem.Status = CandidateStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}
	err = putPeerPoolMap(native, contract, newView, peerPoolMap)
	if err != nil {
		return fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}
	oldView := view - 1
	oldViewBytes := GetUint32Bytes(oldView)
	native.CacheDB.Delete(utils.ConcatKey(contract, []byte(PEER_POOL), oldViewBytes))

	//update split fee
	splitFee, err := getSplitFee(native, contract)
	if err != nil {
		return fmt.Errorf("getSplitFee, getSplitFee error: %v", err)
	}
	err = putSplitFee(native, contract, splitSum+splitFee)
	if err != nil {
		return fmt.Errorf("putSplitFee, put splitFee error: %v", err)
	}

	return nil
}

func splitNodeFee(native *native.NativeService, contract common.Address, peerPubkey string, peerAddress common.Address,
	preIfConsensus, ifConsensus bool, initPos, totalPos uint64, nodeAmount uint64) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	//fee split of address
	//get peerCost
	peerCost, stakeCost, err := getPeerCost(native, contract, peerPubkey)
	if err != nil {
		return fmt.Errorf("getPeerCost, getPeerCost error: %v", err)
	}
	if stakeCost == 0 {
		stakeCost = peerCost
	}
	if stakeCost == 101 { // in storage, 101 means 0, 0 means null
		stakeCost = 0
	}
	var amount uint64
	if native.Height > config.GetNewPeerCostHeight() {
		stakeFee := new(big.Int).Div(
			new(big.Int).Mul(new(big.Int).SetUint64(nodeAmount), new(big.Int).SetUint64(totalPos)),
			new(big.Int).Add(new(big.Int).SetUint64(initPos), new(big.Int).SetUint64(totalPos))).Uint64()
		nodeFee := nodeAmount - stakeFee
		amount = stakeFee*(100-stakeCost)/100 + nodeFee*(100-peerCost)/100
	} else {
		amount = nodeAmount * (100 - peerCost) / 100
	}
	var sumAmount uint64 = 0
	iter := native.CacheDB.NewIterator(utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	defer iter.Release()
	for has := iter.First(); has; has = iter.Next() {
		authorizeInfoStore, err := cstates.GetValueFromRawStorageItem(iter.Value())
		if err != nil {
			return fmt.Errorf("authorizeInfoStore is not available!:%v", err)
		}
		var authorizeInfo AuthorizeInfo
		if err := authorizeInfo.Deserialization(common.NewZeroCopySource(authorizeInfoStore)); err != nil {
			return fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}

		//fee split
		splitAmount, err := executeAddressSplit(native, contract, &authorizeInfo, preIfConsensus, ifConsensus, totalPos, amount, peerAddress)
		if err != nil {
			return fmt.Errorf("excuteAddressSplit, excuteAddressSplit error: %v", err)
		}
		sumAmount = sumAmount + splitAmount
	}
	if err := iter.Error(); err != nil {
		return err
	}
	//split fee to peer
	remainAmount := nodeAmount - sumAmount
	err = executePeerSplit(native, contract, peerAddress, remainAmount)
	if err != nil {
		return fmt.Errorf("excutePeerSplit, excutePeerSplit error: %v", err)
	}
	return nil
}
