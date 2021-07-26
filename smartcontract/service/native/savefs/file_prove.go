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
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type FileProve struct {
	FileHash    []byte
	ProveData   []byte
	BlockHeight uint64
	NodeWallet  common.Address
	Profit      uint64
	SectorID    uint64
}

func (this *FileProve) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.FileHash); err != nil {
		return fmt.Errorf("[FileProve] [this.FileHash:%v] serialize from error:%v", this.FileHash, err)
	}
	if err := utils.WriteBytes(w, this.ProveData); err != nil {
		return fmt.Errorf("[FileProve] [this.FileHash:%v] serialize from error:%v", this.FileHash, err)
	}
	if err := utils.WriteVarUint(w, this.BlockHeight); err != nil {
		return fmt.Errorf("[FileProve] [this.BlockHeight:%v] serialize from error:%v", this.BlockHeight, err)
	}
	if err := utils.WriteAddress(w, this.NodeWallet); err != nil {
		return fmt.Errorf("[FileProve] [this.NodeWallet:%v] serialize from error:%v", this.NodeWallet, err)
	}
	if err := utils.WriteVarUint(w, this.Profit); err != nil {
		return fmt.Errorf("[FileProve] [this.Profit:%v] serialize from error:%v", this.Profit, err)
	}
	if err := utils.WriteVarUint(w, this.SectorID); err != nil {
		return fmt.Errorf("[FileProve] [this.SectorID:%v] serialize from error:%v", this.SectorID, err)
	}
	return nil
}

func (this *FileProve) Deserialize(r io.Reader) error {
	var err error
	if this.FileHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FileProve] [FileHash] deserialize from error:%v", err)
	}
	if this.ProveData, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FileProve] [ProveData] deserialize from error:%v", err)
	}
	if this.BlockHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileProve] [BlockHeight] deserialize from error:%v", err)
	}
	if this.NodeWallet, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[FileProve] [NodeWallet] deserialize from error:%v", err)
	}
	if this.Profit, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileProve] [Profit] deserialize from error:%v", err)
	}
	if this.SectorID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileProve] [SectorID] deserialize from error:%v", err)
	}
	return nil
}

func (this *FileProve) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.FileHash)
	utils.EncodeBytes(sink, this.ProveData)
	utils.EncodeVarUint(sink, this.BlockHeight)
	utils.EncodeAddress(sink, this.NodeWallet)
	utils.EncodeVarUint(sink, this.Profit)
	utils.EncodeVarUint(sink, this.SectorID)
}

func (this *FileProve) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.ProveData, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.BlockHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.NodeWallet, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Profit, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.SectorID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}
