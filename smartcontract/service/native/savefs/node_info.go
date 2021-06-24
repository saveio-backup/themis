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

type FsNodesInfo struct {
	NodeNum  uint64
	NodeInfo []FsNodeInfo
}

type FsNodeInfo struct {
	Pledge      uint64
	Profit      uint64
	Volume      uint64
	RestVol     uint64
	ServiceTime uint64
	WalletAddr  common.Address
	NodeAddr    []byte
}

func (this *FsNodeInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.Pledge); err != nil {
		return fmt.Errorf("[FsNodeInfo] [Pledge:%v] serialize from error:%v", this.Pledge, err)
	}
	if err := utils.WriteVarUint(w, this.Profit); err != nil {
		return fmt.Errorf("[FsNodeInfo] [Profit:%v] serialize from error:%v", this.Profit, err)
	}
	if err := utils.WriteVarUint(w, this.Volume); err != nil {
		return fmt.Errorf("[FsNodeInfo] [Volume:%v] serialize from error:%v", this.Volume, err)
	}
	if err := utils.WriteVarUint(w, this.RestVol); err != nil {
		return fmt.Errorf("[FsNodeInfo] [RestVol:%v] serialize from error:%v", this.RestVol, err)
	}
	if err := utils.WriteVarUint(w, this.ServiceTime); err != nil {
		return fmt.Errorf("[FsNodeInfo] [ServiceTime:%v] serialize from error:%v", this.ServiceTime, err)
	}
	if err := utils.WriteAddress(w, this.WalletAddr); err != nil {
		return fmt.Errorf("[FsNodeInfo] [WalletAddr:%v] serialize from error:%v", this.WalletAddr, err)
	}
	if err := utils.WriteBytes(w, this.NodeAddr); err != nil {
		return fmt.Errorf("[FsNodeInfo] [NodeAddr:%v] serialize from error:%v", this.NodeAddr, err)
	}
	return nil
}

func (this *FsNodeInfo) Deserialize(r io.Reader) error {
	var err error
	if this.Pledge, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsNodeInfo] [Pledge] Deserialize from error:%v", err)
	}
	if this.Profit, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsNodeInfo] [Profit] Deserialize from error:%v", err)
	}
	if this.Volume, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsNodeInfo] [Volume] Deserialize from error:%v", err)
	}
	if this.RestVol, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsNodeInfo] [RestVol] Deserialize from error:%v", err)
	}
	if this.ServiceTime, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsNodeInfo] [ServiceTime] Deserialize from error:%v", err)
	}
	if this.WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[FsNodeInfo] [WalletAddr] Deserialize from error:%v", err)
	}
	if this.NodeAddr, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FsNodeInfo] [NodeAddr] Deserialize from error:%v", err)
	}
	return nil
}

func (this *FsNodeInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.Pledge)
	utils.EncodeVarUint(sink, this.Profit)
	utils.EncodeVarUint(sink, this.Volume)
	utils.EncodeVarUint(sink, this.RestVol)
	utils.EncodeVarUint(sink, this.ServiceTime)
	utils.EncodeAddress(sink, this.WalletAddr)
	utils.EncodeBytes(sink, this.NodeAddr)
}

func (this *FsNodeInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Pledge, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Profit, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Volume, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.RestVol, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ServiceTime, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.NodeAddr, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}

func (this *FsNodesInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.NodeNum); err != nil {
		return fmt.Errorf("[FsNodeInfos] [NodeNum:%v] serialize from error:%v", this.NodeNum, err)
	}
	for i := 0; uint64(i) < this.NodeNum; i++ {
		if err := this.NodeInfo[i].Serialize(w); err != nil {
			return fmt.Errorf("[FsNodeInfos] [NodeInfo] serialize from error:%v", err)
		}
	}
	return nil
}

func (this *FsNodesInfo) Deserialize(r io.Reader) error {
	var err error
	if this.NodeNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsNodeInfos] [NodeNum] deserialize from error:%v", err)
	}
	var nodeInfo FsNodeInfo
	for i := 0; uint64(i) < this.NodeNum; i++ {
		if err := nodeInfo.Deserialize(r); err != nil {
			return fmt.Errorf("[FsNodeInfos] [NodeInfo] deserialize from error:%v", err)
		}
		this.NodeInfo = append(this.NodeInfo, nodeInfo)
	}
	return nil
}

func getFsNodeInfo(native *native.NativeService, walletAddr common.Address) (*FsNodeInfo, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	fsNodeInfoKey := GenFsNodeInfoKey(contract, walletAddr)
	item, err := utils.GetStorageItem(native, fsNodeInfoKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] FsNodeInfo GetStorageItem error!")
	}
	if item == nil {
		return nil, errors.NewErr("[FS Govern] FsNodeInfo not found!")
	}
	var fsNodeInfo FsNodeInfo
	fsNodeInfoSource := common.NewZeroCopySource(item.Value)
	err = fsNodeInfo.Deserialization(fsNodeInfoSource)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] FsNodeInfo deserialize error!")
	}
	return &fsNodeInfo, nil
}

func setFsNodeInfo(native *native.NativeService, fsNodeInfo *FsNodeInfo) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	fsNodeInfoKey := GenFsNodeInfoKey(contract, fsNodeInfo.WalletAddr)
	info := new(bytes.Buffer)
	if err := fsNodeInfo.Serialize(info); err != nil {
		return errors.NewErr("[FS Govern] FsNodeInfo serialize error!")
	}
	utils.PutBytes(native, fsNodeInfoKey, info.Bytes())
	return nil
}
