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

type NodeList struct {
	AddrNum  uint64
	AddrList []common.Address
}

func (this *NodeList) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.AddrNum); err != nil {
		return fmt.Errorf("[NodeList] [AddrNum:%v] serialize from error:%v", this.AddrNum, err)
	}

	for index := 0; uint64(index) < this.AddrNum; index++ {
		if err := utils.WriteAddress(w, this.AddrList[index]); err != nil {
			return fmt.Errorf("[NodeList] [AddrList:%v] serialize from error:%v", this.AddrList[index], err)
		}
	}
	return nil
}

func (this *NodeList) Deserialize(r io.Reader) error {
	var err error
	if this.AddrNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[NodeList] [AddrNum] deserialize from error:%v", err)
	}
	var tmpAddr common.Address
	for index := 0; uint64(index) < this.AddrNum; index++ {
		if tmpAddr, err = utils.ReadAddress(r); err != nil {
			return fmt.Errorf("[NodeList] [AddrList] deserialize from error:%v", err)
		}
		this.AddrList = append(this.AddrList, tmpAddr)
	}
	return nil
}

func (this *NodeList) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.AddrNum)
	for _, addr := range this.AddrList {
		utils.EncodeAddress(sink, addr)
	}
}

func (this *NodeList) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.AddrNum, err = utils.DecodeVarUint(source); err != nil {
		return err
	}
	addrs := make([]common.Address, 0, this.AddrNum)
	for i := uint64(0); i < this.AddrNum; i++ {
		addr, err := utils.DecodeAddress(source)
		if err != nil {
			return err
		}
		addrs = append(addrs, addr)
	}
	this.AddrList = addrs
	return err
}

func (this *NodeList) Add(addr common.Address) error {
	flag := false
	for i := uint64(0); i < this.AddrNum; i++ {
		if this.AddrList[i] == common.ADDRESS_EMPTY {
			this.AddrList[i] = addr
			flag = true
			break
		}
	}
	if !flag {
		this.AddrList = append(this.AddrList, addr)
		this.AddrNum++
	}
	return nil
}

func (this *NodeList) Del(addr common.Address) error {
	for i := uint64(0); i < this.AddrNum; i++ {
		if this.AddrList[i] == addr {
			this.AddrList[i] = common.ADDRESS_EMPTY
		}
	}
	return nil
}

func (this *NodeList) GetList() []common.Address {
	var addrs []common.Address
	for i := uint64(0); i < this.AddrNum; i++ {
		if this.AddrList[i] == common.ADDRESS_EMPTY {
			continue
		}
		addrs = append(addrs, this.AddrList[i])
	}
	return addrs
}

func (this *NodeList) Exist(addr common.Address) bool {
	for i := uint64(0); i < this.AddrNum; i++ {
		if this.AddrList[i] == addr {
			return true
		}
	}
	return false
}

func getFsNodeList(native *native.NativeService) (*NodeList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	nodeSetKey := GenFsNodeSetKey(contract)
	nodeSet, err := utils.GetStorageItem(native, nodeSetKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] GetStorageItem nodeSetKey error!")
	}
	if nodeSet == nil {
		return nil, errors.NewErr("[FS Govern] FsGetNodeList No nodeSet found!")
	}

	var nodeList NodeList
	reader := bytes.NewReader(nodeSet.Value)
	if err = nodeList.Deserialize(reader); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] Set deserialize error!")
	}
	return &nodeList, nil
}

func nodeListOperate(native *native.NativeService, walletAddr common.Address, isAdd bool) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	nodeSetKey := GenFsNodeSetKey(contract)
	nodeSet, err := utils.GetStorageItem(native, nodeSetKey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] GetStorageItem nodeSetKey error!")
	}

	var nodeList NodeList
	if nodeSet != nil {
		reader := bytes.NewReader(nodeSet.Value)
		if err = nodeList.Deserialize(reader); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] Set deserialize error!")
		}
	}

	if isAdd {
		nodeList.Add(walletAddr)
	} else {
		nodeList.Del(walletAddr)
	}
	bf := new(bytes.Buffer)
	err = nodeList.Serialize(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] Put node to set error!")
	}
	utils.PutBytes(native, nodeSetKey, bf.Bytes())
	return nil
}
