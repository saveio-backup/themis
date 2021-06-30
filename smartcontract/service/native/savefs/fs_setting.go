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

type FsSetting struct {
	FsGasPrice         uint64 // gas price for fs contract
	GasPerGBPerBlock   uint64 // gas for store block
	GasPerKBForRead    uint64 // gas for read file
	GasForChallenge    uint64 // gas for challenge
	MaxProveBlockNum   uint64
	MinVolume          uint64
	DefaultProvePeriod uint64 // default prove interval
	DefaultProveLevel  uint64 // default prove level
	DefaultCopyNum     uint64 // default copy number
}

func (this *FsSetting) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.FsGasPrice); err != nil {
		return fmt.Errorf("[FsSetting] [FsGasPrice:%v] serialize from error:%v", this.FsGasPrice, err)
	}
	if err := utils.WriteVarUint(w, this.GasPerGBPerBlock); err != nil {
		return fmt.Errorf("[FsSetting] [GasPerGBPerBlock:%v] serialize from error:%v", this.GasPerGBPerBlock, err)
	}
	if err := utils.WriteVarUint(w, this.GasPerKBForRead); err != nil {
		return fmt.Errorf("[FsSetting] [GasPerKBForRead:%v] serialize from error:%v", this.GasPerKBForRead, err)
	}
	if err := utils.WriteVarUint(w, this.GasForChallenge); err != nil {
		return fmt.Errorf("[FsSetting] [GasForChallenge:%v] serialize from error:%v", this.GasForChallenge, err)
	}
	if err := utils.WriteVarUint(w, this.MaxProveBlockNum); err != nil {
		return fmt.Errorf("[FsSetting] [MaxProveBlockNum:%v] serialize from error:%v", this.MaxProveBlockNum, err)
	}
	if err := utils.WriteVarUint(w, this.MinVolume); err != nil {
		return fmt.Errorf("[FsSetting] [MinVolume:%v] serialize from error:%v", this.MinVolume, err)
	}
	if err := utils.WriteVarUint(w, this.DefaultProvePeriod); err != nil {
		return fmt.Errorf("[FsSetting] [DefaultProvePeriod:%v] serialize from error:%v", this.DefaultProvePeriod, err)
	}
	if err := utils.WriteVarUint(w, this.DefaultProveLevel); err != nil {
		return fmt.Errorf("[FsSetting] [DefaultProveLevel:%v] serialize from error:%v", this.DefaultProveLevel, err)
	}
	if err := utils.WriteVarUint(w, this.DefaultCopyNum); err != nil {
		return fmt.Errorf("[FsSetting] [DefaultCopyNum:%v] serialize from error:%v", this.DefaultCopyNum, err)
	}
	return nil
}

func (this *FsSetting) Deserialize(r io.Reader) error {
	var err error
	if this.FsGasPrice, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsSetting] [FsGasPrice] Deserialize from error:%v", err)
	}
	if this.GasPerGBPerBlock, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsSetting] [GasPerGBPerBlock] Deserialize from error:%v", err)
	}
	if this.GasPerKBForRead, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsSetting] [GasPerKBForRead] Deserialize from error:%v", err)
	}
	if this.GasForChallenge, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsSetting] [GasForChallenge] Deserialize from error:%v", err)
	}
	if this.MaxProveBlockNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsSetting] [MaxProveBlockNum] Deserialize from error:%v", err)
	}
	if this.MinVolume, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsSetting] [MinVolume] Deserialize from error:%v", err)
	}
	if this.DefaultProvePeriod, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsSetting] [DefaultProvePeriod] Deserialize from error:%v", err)
	}
	if this.DefaultProveLevel, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsSetting] [DefaultProveLevel] Deserialize from error:%v", err)
	}
	if this.DefaultCopyNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FsSetting] [DefaultCopyNum] Deserialize from error:%v", err)
	}
	return nil
}

func (this *FsSetting) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.FsGasPrice)
	utils.EncodeVarUint(sink, this.GasPerGBPerBlock)
	utils.EncodeVarUint(sink, this.GasPerKBForRead)
	utils.EncodeVarUint(sink, this.GasForChallenge)
	utils.EncodeVarUint(sink, this.MaxProveBlockNum)
	utils.EncodeVarUint(sink, this.MinVolume)
	utils.EncodeVarUint(sink, this.DefaultProvePeriod)
	utils.EncodeVarUint(sink, this.DefaultProveLevel)
	utils.EncodeVarUint(sink, this.DefaultCopyNum)
}

func (this *FsSetting) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FsGasPrice, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.GasPerGBPerBlock, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.GasPerKBForRead, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.GasForChallenge, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.MaxProveBlockNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.MinVolume, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.DefaultProvePeriod, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.DefaultProveLevel, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.DefaultCopyNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return err
}

type StorageFee struct {
	TxnFee        uint64
	SpaceFee      uint64
	ValidationFee uint64
}

func (f *StorageFee) Sum() uint64 {
	return f.TxnFee + f.SpaceFee + f.ValidationFee
}

func (this *StorageFee) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.TxnFee); err != nil {
		return fmt.Errorf("[StorageFee] [TxnFee:%v] serialize from error:%v", this.TxnFee, err)
	}
	if err := utils.WriteVarUint(w, this.SpaceFee); err != nil {
		return fmt.Errorf("[StorageFee] [SpaceFee:%v] serialize from error:%v", this.SpaceFee, err)
	}
	if err := utils.WriteVarUint(w, this.ValidationFee); err != nil {
		return fmt.Errorf("[StorageFee] [ValidationFee:%v] serialize from error:%v", this.ValidationFee, err)
	}
	return nil
}

func (this *StorageFee) Deserialize(r io.Reader) error {
	var err error
	if this.TxnFee, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[StorageFee] [TxnFee] Deserialize from error:%v", err)
	}
	if this.SpaceFee, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[StorageFee] [SpaceFee] Deserialize from error:%v", err)
	}
	if this.ValidationFee, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[StorageFee] [ValidationFee] Deserialize from error:%v", err)
	}
	return nil
}
