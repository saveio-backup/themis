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
	"bytes"
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type Rule struct {
	Addr         common.Address
	BaseHeight   uint64
	ExpireHeight uint64
}

func (this *Rule) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Addr); err != nil {
		return fmt.Errorf("[Rule] [Addr:%v] serialize from error:%v", this.Addr, err)
	}
	if err := utils.WriteVarUint(w, this.BaseHeight); err != nil {
		return fmt.Errorf("[Rule] [BaseHeight:%v] serialize from error:%v", this.BaseHeight, err)
	}
	if err := utils.WriteVarUint(w, this.ExpireHeight); err != nil {
		return fmt.Errorf("[Rule] [ExpireHeight:%v] serialize from error:%v", this.ExpireHeight, err)
	}
	return nil
}

func (this *Rule) Deserialize(r io.Reader) error {
	var err error
	if this.Addr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[Rule] [Addr] deserialize from error:%v", err)
	}
	if this.BaseHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Rule] [BaseHeight] deserialize from error:%v", err)
	}
	if this.ExpireHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Rule] [ExpireHeight] deserialize from error:%v", err)
	}
	return nil
}

func (this *Rule) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Addr)
	utils.EncodeVarUint(sink, this.BaseHeight)
	utils.EncodeVarUint(sink, this.ExpireHeight)
}

func (this *Rule) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Addr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.BaseHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ExpireHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return err
}

type WhiteList struct {
	Num  uint64
	List []Rule
}

func (this *WhiteList) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.Num)
	for i := uint64(0); i < this.Num; i++ {
		this.List[i].Serialization(sink)
	}
}

func (this *WhiteList) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Num, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	for index := 0; uint64(index) < this.Num; index++ {
		var rule Rule
		err := rule.Deserialization(source)
		if err != nil {
			return err
		}
		this.List[index] = rule
	}
	return err
}

func (this *WhiteList) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.Num); err != nil {
		return fmt.Errorf("[WhiteList] [Num:%v] serialize from error:%v", this.Num, err)
	}

	for index := 0; uint64(index) < this.Num; index++ {
		if err := this.List[index].Serialize(w); err != nil {
			return fmt.Errorf("[WhiteList] [List:%v] serialize from error:%v", this.List[index], err)
		}
	}
	return nil
}

func (this *WhiteList) Deserialize(r io.Reader) error {
	var err error
	if this.Num, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[WhiteList] [Num] deserialize from error:%v", err)
	}
	var tmpRule Rule
	for index := 0; uint64(index) < this.Num; index++ {
		if tmpRule.Deserialize(r); err != nil {
			return fmt.Errorf("[WhiteList] [List] deserialize from error:%v", err)
		}
		this.List = append(this.List, tmpRule)
	}
	return nil
}

func (this *WhiteList) Add(rules []Rule) error {
	for _, rule := range rules {
		if rule.ExpireHeight <= rule.BaseHeight {
			return errors.NewErr("Rule ExpireHeight < BaseHeight")
		}
		flag := false
		for i := uint64(0); i < this.Num; i++ {
			if this.List[i].Addr == rule.Addr {
				this.List[i].BaseHeight = rule.BaseHeight
				this.List[i].ExpireHeight = rule.ExpireHeight
				flag = true
				break
			}
		}
		if !flag {
			this.List = append(this.List, rule)
			this.Num++
		}
	}
	return nil
}

func (this *WhiteList) Del(rules []Rule) {
	for _, rule := range rules {
		if this.Num == 0 {
			return
		}
		for i := uint64(0); i < this.Num; i++ {
			if this.List[i].Addr == rule.Addr {
				this.List = append(this.List[:i], this.List[i+1:]...)
				this.Num -= 1
				break
			}
		}
	}
	return
}

func (this *WhiteList) Check(addr common.Address, curHeight uint64) bool {
	flag := false
	for i := uint64(0); i < this.Num; i++ {
		if this.List[i].Addr == addr && this.List[i].BaseHeight < curHeight &&
			this.List[i].ExpireHeight > curHeight {
			flag = true
			break
		}
	}
	return flag
}

func AddRulesToList(native *native.NativeService, fileHash []byte, rules []Rule) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	var whiteList *WhiteList
	whiteList, err := GetWhiteList(native, fileHash)
	if whiteList == nil {
		whiteList = new(WhiteList)
	}
	if err := whiteList.Add(rules); err != nil {
		return err
	}

	whiteListKey := GenFsWhiteListKey(contract, fileHash)
	whiteListBf := new(bytes.Buffer)
	if err = whiteList.Serialize(whiteListBf); err != nil {
		return errors.NewErr("[FS Profit] AddRulesToList fileList serialize error!")
	}
	utils.PutBytes(native, whiteListKey, whiteListBf.Bytes())
	return nil
}

func DelRulesFromList(native *native.NativeService, fileHash []byte, rules []Rule) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var whiteList *WhiteList
	whiteList, err := GetWhiteList(native, fileHash)
	if err != nil {
		return errors.NewErr("[FS Profit] DelAddrFromList GetFsWhiteList error!")
	}
	if whiteList == nil || whiteList.Num == 0 {
		return nil
	}
	whiteList.Del(rules)

	whiteListKey := GenFsWhiteListKey(contract, fileHash)
	whiteListBf := new(bytes.Buffer)
	if err = whiteList.Serialize(whiteListBf); err != nil {
		return errors.NewErr("[FS Profit] DelRulesFromList whiteList serialize error!")
	}
	utils.PutBytes(native, whiteListKey, whiteListBf.Bytes())
	return nil
}

func CovRulesToList(native *native.NativeService, fileHash []byte, rules []Rule) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	for _, rule := range rules {
		if rule.ExpireHeight <= rule.BaseHeight {
			return errors.NewErr("Rule ExpireHeight < BaseHeight")
		}
	}

	whiteListKey := GenFsWhiteListKey(contract, fileHash)
	utils.DelStorageItem(native, whiteListKey)

	whiteList := WhiteList{Num: uint64(len(rules)), List: rules}
	whiteListBf := new(bytes.Buffer)
	if err := whiteList.Serialize(whiteListBf); err != nil {
		return errors.NewErr("[FS Profit] CovRulesToList fileList serialize error!")
	}
	utils.PutBytes(native, whiteListKey, whiteListBf.Bytes())
	return nil
}

func CleRulesFromList(native *native.NativeService, fileHash []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	whiteListKey := GenFsWhiteListKey(contract, fileHash)
	utils.DelStorageItem(native, whiteListKey)
	return nil
}

func CheckPrivilege(native *native.NativeService, fileHash []byte, addr common.Address) bool {
	var whiteList *WhiteList
	whiteList, err := GetWhiteList(native, fileHash)
	if err != nil {
		return false
	}
	if whiteList == nil || whiteList.Num == 0 {
		return false
	}
	return whiteList.Check(addr, uint64(native.Height))
}

func GetWhiteList(native *native.NativeService, fileHash []byte) (*WhiteList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	whiteListKey := GenFsWhiteListKey(contract, fileHash)
	item, err := utils.GetStorageItem(native, whiteListKey)
	if err != nil {
		return nil, errors.NewErr("[FS Profit] GetWhiteList GetStorageItem error!")
	}
	if item == nil {
		return &WhiteList{0, nil}, nil
	}

	var fsWhiteList WhiteList
	reader := bytes.NewReader(item.Value)
	if err = fsWhiteList.Deserialize(reader); err != nil {
		return nil, errors.NewErr("[FS Profit] GetWhiteList deserialize error!")
	}
	return &fsWhiteList, nil
}
