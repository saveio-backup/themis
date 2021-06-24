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

type FsProveDetails struct {
	CopyNum        uint64
	ProveDetailNum uint64
	ProveDetails   []ProveDetail
}

type ProveDetail struct {
	NodeAddr    []byte
	WalletAddr  common.Address
	ProveTimes  uint64
	BlockHeight uint64 // block height for first file prove
	Finished    bool
}

func (this *ProveDetail) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.NodeAddr); err != nil {
		return fmt.Errorf("[ProveNode] [NodeAddr:%v] serialize from error:%v", this.NodeAddr, err)
	}
	if err := utils.WriteAddress(w, this.WalletAddr); err != nil {
		return fmt.Errorf("[ProveNode] [WalletAddr:%v] serialize from error:%v", this.WalletAddr, err)
	}
	if err := utils.WriteVarUint(w, this.ProveTimes); err != nil {
		return fmt.Errorf("[ProveNode] [ProveTimes:%v] serialize from error:%v", this.ProveTimes, err)
	}
	if err := utils.WriteVarUint(w, this.BlockHeight); err != nil {
		return fmt.Errorf("[ProveNode] [BlockHeight:%v] serialize from error:%v", this.BlockHeight, err)
	}
	if err := utils.WriteBool(w, this.Finished); err != nil {
		return fmt.Errorf("[ProveNode] [Finished:%v] serialize from error:%v", this.Finished, err)
	}
	return nil
}

func (this *ProveDetail) Deserialize(r io.Reader) error {
	var err error
	if this.NodeAddr, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[ProveNode] [NodeAddr] deserialize from error:%v", err)
	}
	if this.WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[ProveNode] [WalletAddr] deserialize from error:%v", err)
	}
	if this.ProveTimes, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[ProveNode] [ProveTimes] deserialize from error:%v", err)
	}
	if this.BlockHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[ProveNode] [BlockHeight] deserialize from error:%v", err)
	}
	if this.Finished, err = utils.ReadBool(r); err != nil {
		return fmt.Errorf("[ProveNode] [Finished] deserialize from error:%v", err)
	}
	return nil
}

func (this *FsProveDetails) Serialize(w io.Writer) error {
	var err error
	if err = utils.WriteVarUint(w, this.CopyNum); err != nil {
		return fmt.Errorf("[ProveDetail] [CopyNum:%v] serialize from error:%v", this.CopyNum, err)
	}
	if err = utils.WriteVarUint(w, this.ProveDetailNum); err != nil {
		return fmt.Errorf("[ProveDetail] [ProveDetailNum:%v] serialize from error:%v", this.ProveDetailNum, err)
	}
	for _, v := range this.ProveDetails {
		err = v.Serialize(w)
		if err != nil {
			return fmt.Errorf("[ProveDetail] [ProveDetail] serialize from error:%v", err)
		}
	}
	return nil
}

func (this *FsProveDetails) Deserialize(r io.Reader) error {
	var err error
	var tmpProveDetail ProveDetail
	if this.CopyNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[ProveDetail] [CopyNum] deserialize from error:%v", err)
	}
	if this.ProveDetailNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[ProveDetail] [ProveDetailNum] deserialize from error:%v", err)
	}
	for i := 0; uint64(i) < this.ProveDetailNum; i++ {
		if err = tmpProveDetail.Deserialize(r); err != nil {
			return fmt.Errorf("[ProveDetail] [ProveDetail] deserialize from error:%v", err)
		}
		this.ProveDetails = append(this.ProveDetails, tmpProveDetail)
	}
	return nil
}

func getProveDetailsWithNodeAddr(native *native.NativeService, fileHash []byte) (*FsProveDetails, error) {
	proveDetails, err := getProveDetails(native, fileHash)
	if err != nil {
		return nil, err
	}

	for i := uint64(0); i < proveDetails.ProveDetailNum; i++ {
		nodeInfo, err := getFsNodeInfo(native, proveDetails.ProveDetails[i].WalletAddr)
		if err != nil {
			return nil, errors.NewErr("[FS Govern] GetProveDetails GetFsNodeInfo error!")
		}
		proveDetails.ProveDetails[i].NodeAddr = nodeInfo.NodeAddr
	}
	return proveDetails, nil
}

func getProveDetails(native *native.NativeService, fileHash []byte) (*FsProveDetails, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	proveDetailKey := GenFsProveDetailsKey(contract, fileHash)
	item, err := utils.GetStorageItem(native, proveDetailKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FileProveDetails GetStorageItem error!")
	}
	if item == nil {
		return nil, errors.NewErr("[FS Profit] FileProveDetails not found!")
	}

	var proveDetails FsProveDetails
	reader := bytes.NewReader(item.Value)
	err = proveDetails.Deserialize(reader)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] GetProveDetails deserialize error!")
	}
	return &proveDetails, nil
}

func setProveDetails(native *native.NativeService, fileHash []byte, proveDetails *FsProveDetails) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	proveDetailsKey := GenFsProveDetailsKey(contract, fileHash)
	proveDetailsBuff := new(bytes.Buffer)
	if err := proveDetails.Serialize(proveDetailsBuff); err != nil {
		return errors.NewErr("[FS Govern] ProveDetails serialize error!")
	}
	utils.PutBytes(native, proveDetailsKey, proveDetailsBuff.Bytes())
	return nil
}

func deleteProveDetails(native *native.NativeService, fileHash []byte) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	fileInfoKey := GenFsFileInfoKey(contract, fileHash)
	utils.DelStorageItem(native, fileInfoKey)
}
