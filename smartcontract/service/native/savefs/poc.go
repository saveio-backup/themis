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

package savefs

import (
	"fmt"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type PocProve struct {
	Miner    common.Address
	Height   uint32
	PlotSize uint64
}

func (p *PocProve) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(p.Miner)
	sink.WriteUint32(p.Height)
	sink.WriteUint64(p.PlotSize)
}

func (p *PocProve) Deserialization(source *common.ZeroCopySource) error {
	var err error
	p.Miner, err = utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize Miner error: %v", err)
	}

	p.Height, err = utils.DecodeUint32(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize Height error: %v", err)
	}

	p.PlotSize, err = utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize PlotSize error: %v", err)
	}

	return nil
}

type PocProveList struct {
	Proves []*PocProve
}

func (pl *PocProveList) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(pl.Proves)))
	for _, prove := range pl.Proves {
		prove.Serialization(sink)
	}
}

func (pl *PocProveList) Deserialization(source *common.ZeroCopySource) error {
	length, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeUint64, deserialize prvoce length error: %v", err)
	}
	list := make([]*PocProve, 0)
	for i := uint64(0); i < length; i++ {
		p := &PocProve{}
		if err := p.Deserialization(source); err != nil {
			return fmt.Errorf("utils.Deserialization, deserialize prvoce  error: %v", err)
		}
		list = append(list, p)
	}

	pl.Proves = list

	return nil
}

func FsGetPocProveList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	height, err := utils.DecodeUint32(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FsGetPocProveList] height deserialize error!")
	}
	list := getPocProveList(native, height)

	if list == nil {
		list = &PocProveList{
			Proves: make([]*PocProve, 0),
		}
	}

	sink := common.NewZeroCopySink(nil)

	list.Serialization(sink)

	return EncRet(true, sink.Bytes()), nil

}

func getPocProve(native *native.NativeService, miner common.Address, height uint32) *PocProve {
	contract := native.ContextRef.CurrentContext().ContractAddress
	key := GenPocProveKey(contract, miner, uint64(height))
	item, err := utils.GetStorageItem(native, []byte(key))
	if err != nil {
		log.Errorf("[FS PoC] getPocProve GetStorageItem error %s", err)
		return nil
	}
	if item == nil {
		log.Errorf("[FS PoC] getPocProve not found!")
		return nil
	}

	prove := &PocProve{}
	source := common.NewZeroCopySource(item.Value)
	if err := prove.Deserialization(source); err != nil {
		return nil
	}
	return prove
}

func putPocProve(native *native.NativeService, prove *PocProve) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	key := GenPocProveKey(contract, prove.Miner, uint64(prove.Height))
	sink := common.NewZeroCopySink(nil)
	prove.Serialization(sink)
	utils.PutBytes(native, key, sink.Bytes())

	proveList := getPocProveList(native, prove.Height)

	if proveList == nil {
		proveList = &PocProveList{
			Proves: make([]*PocProve, 0),
		}
	}

	proveList.Proves = append(proveList.Proves, prove)

	return putPocProveList(native, proveList, prove.Height)
}

func getPocProveList(native *native.NativeService, height uint32) *PocProveList {
	contract := native.ContextRef.CurrentContext().ContractAddress
	key := GenPocProveListKey(contract, uint64(height))
	item, err := utils.GetStorageItem(native, []byte(key))
	if err != nil {
		log.Errorf("[FS PoC] getPocProve GetStorageItem error %s", err)
		return nil
	}
	if item == nil {
		log.Errorf("[FS PoC] getPocProve not found!")
		return nil
	}

	prove := &PocProveList{}
	source := common.NewZeroCopySource(item.Value)
	if err := prove.Deserialization(source); err != nil {
		return nil
	}
	return prove
}

func putPocProveList(native *native.NativeService, list *PocProveList, height uint32) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	key := GenPocProveListKey(contract, uint64(height))

	sink := common.NewZeroCopySink(nil)
	list.Serialization(sink)
	utils.PutBytes(native, key, sink.Bytes())
	return nil
}
