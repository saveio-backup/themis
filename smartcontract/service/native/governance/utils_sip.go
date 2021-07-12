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
	"fmt"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/serialization"
	cstates "github.com/saveio/themis/core/states"
	"github.com/saveio/themis/smartcontract/service/native"

	//fs "github.com/saveio/themis/smartcontract/service/native/savefs"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

//sip
func getSipIndex(native *native.NativeService, contract common.Address) (uint32, error) {
	sipIndexBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIP_SEQ_INDEX)))
	if err != nil {
		return 0, fmt.Errorf("native.CacheDB.Get, get candidateIndex error: %v", err)
	}
	if sipIndexBytes == nil {
		return 0, fmt.Errorf("getSipIndex, sipIndex is not init")
	} else {
		sipIndexStore, err := cstates.GetValueFromRawStorageItem(sipIndexBytes)
		if err != nil {
			return 0, fmt.Errorf("getSipIndex, deserialize from raw storage item err:%v", err)
		}
		sipIndex, err := GetBytesUint32(sipIndexStore)
		if err != nil {
			return 0, fmt.Errorf("GetBytesUint32, get sipIndex error: %v", err)
		}
		return sipIndex, nil
	}
}

func putSipIndex(native *native.NativeService, contract common.Address, sipIndex uint32) error {
	sipIndexBytes := GetUint32Bytes(sipIndex)

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIP_SEQ_INDEX)), cstates.GenRawStorageItem(sipIndexBytes))
	return nil
}

func GetSipMap(native *native.NativeService, contract common.Address) (*SipMap, error) {
	sipMap := &SipMap{
		SipMap: make(map[uint32]*SIP),
	}

	sipMapBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIP_POOL)))
	if err != nil {
		return nil, fmt.Errorf("getSipMap, get all sipMap error: %v", err)
	}
	if sipMapBytes == nil {
		return nil, fmt.Errorf("getSipMap, sipMap is nil")
	}
	item := cstates.StorageItem{}
	source := common.NewZeroCopySource(sipMapBytes)
	err = item.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize SipMap error:%v", err)
	}
	sipMapStore := item.Value
	if err := sipMap.Deserialize(bytes.NewBuffer(sipMapStore)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize sipMap error: %v", err)
	}
	return sipMap, nil
}

func putSipMap(native *native.NativeService, contract common.Address, sipMap *SipMap) error {
	bf := new(bytes.Buffer)
	if err := sipMap.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize sipMap error: %v", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIP_POOL)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getSipVoteRevenue(native *native.NativeService, contract common.Address) (*SIPVoteRevenue, error) {
	voteRevenueBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIP_VOTE_REVENUE)))
	if err != nil {
		return nil, fmt.Errorf("getSipVoteRevenue, get voteRevenueBytes error: %v", err)
	}
	voteRevenue := new(SIPVoteRevenue)
	if voteRevenueBytes == nil {
		return nil, fmt.Errorf("getSipVoteRevenue, get nil voteRevenueBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(voteRevenueBytes)
		if err != nil {
			return nil, fmt.Errorf("getSipVoteRevenue, deserialize from raw storage item err:%v", err)
		}
		if err := voteRevenue.Deserialize(bytes.NewBuffer(value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize voteRevenue error: %v", err)
		}
	}
	return voteRevenue, nil
}

func putSipVoteRevenue(native *native.NativeService, contract common.Address, voteRevenue *SIPVoteRevenue) error {
	bf := new(bytes.Buffer)
	if err := voteRevenue.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize voteRevenue error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIP_VOTE_REVENUE)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func increaseSipVoteRevenue(native *native.NativeService, contract common.Address, amount uint64) error {
	voteRevenue, err := getSipVoteRevenue(native, contract)
	if err != nil {
		return fmt.Errorf("increaseSipVoteRevenue, get vote revenue error: %v", err)
	}
	voteRevenue.Total += amount
	putSipVoteRevenue(native, contract, voteRevenue)
	return nil
}

func reserveSipVoteRevenue(native *native.NativeService, contract common.Address, amount uint64) error {
	voteRevenue, err := getSipVoteRevenue(native, contract)
	if err != nil {
		return fmt.Errorf("reserveSipVoteRevenue, get vote revenue error: %v", err)
	}
	free := voteRevenue.Total - voteRevenue.Reserve
	if amount > free {
		return fmt.Errorf("reserveSipVoteRevenue, fail to reserve bonus")
	}
	voteRevenue.Reserve += amount
	putSipVoteRevenue(native, contract, voteRevenue)
	return nil
}

//last height parameter be changed
func getParamChangeHeight(native *native.NativeService, contract common.Address, param string) (uint32, error) {
	heightBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIP_LAST_CHANGE_HEIGHT), []byte(param)))
	if err != nil {
		return 0, fmt.Errorf("getParamChangeHeight, get heightBytes error: %v", err)
	}
	height := uint32(0)
	if heightBytes == nil {
		return 0, fmt.Errorf("GetGasRevenue, get nil voteRevenueBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(heightBytes)
		if err != nil {
			return 0, fmt.Errorf("GetVoteRevenue, deserialize from raw storage item err:%v", err)
		}

		h, err := serialization.ReadUint64(bytes.NewBuffer(value))
		if err != nil {
			return 0, fmt.Errorf("deserialize height vote revenue error: %v", err)
		}
		height = uint32(h)
	}
	return height, nil
}

func putParamChangeHeight(native *native.NativeService, contract common.Address, param string, height uint32) error {
	bf := new(bytes.Buffer)
	if err := serialization.WriteUint64(bf, uint64(height)); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize vote revenue error: %v", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIP_LAST_CHANGE_HEIGHT), []byte(param)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

/*
func appCallChangePDPGas(native *native.NativeService, pdpGas uint64) error {
	param := &fs.ChangePDPGasParam{GasForChallenge: pdpGas}

	bf := new(bytes.Buffer)

	err := param.Serialize(bf)
	if err != nil {
		return fmt.Errorf("appCallChangePDPGas, param serialize error: %v", err)
	}

	_, err = native.NativeCall(utils.OntFSContractAddress, "FsChangePDPGas", bf.Bytes())

	if err != nil {
		return fmt.Errorf("appCallChangePDPGas, call FsChangePDPGas error: %v", err)
	}

	return nil
}
*/
