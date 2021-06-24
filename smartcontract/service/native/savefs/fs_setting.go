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

func getFsSetting(native *native.NativeService) (*FsSetting, error) {
	var fsSetting FsSetting
	contract := native.ContextRef.CurrentContext().ContractAddress
	fsSettingKey := GenFsSettingKey(contract)

	item, err := utils.GetStorageItem(native, fsSettingKey)
	if err != nil {
		return nil, errors.NewErr("[FS Init] GetFsSetting error!")
	}
	if item == nil {
		fsSetting = FsSetting{
			FsGasPrice:         FS_GAS_PRICE,
			GasPerGBPerBlock:   GAS_PER_GB_PER_Block,
			GasPerKBForRead:    GAS_PER_KB_FOR_READ,
			GasForChallenge:    GAS_FOR_CHALLENGE,
			MaxProveBlockNum:   MAX_PROVE_BLOCKS,
			MinVolume:          MIN_VOLUME, //1G
			DefaultProvePeriod: DEFAULT_PROVE_PERIOD,
			DefaultProveLevel:  DeFAULT_PROVE_LEVEL,
			DefaultCopyNum:     DEFAULT_COPY_NUM,
		}
		return &fsSetting, nil
	}

	settingSource := common.NewZeroCopySource(item.Value)
	if err := fsSetting.Deserialization(settingSource); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Init] FsSetting Deserialization error!")
	}
	return &fsSetting, nil
}

func setFsSetting(native *native.NativeService, fsSetting FsSetting) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	info := new(bytes.Buffer)
	fsSetting.Serialize(info)

	fsSettingKey := GenFsSettingKey(contract)
	utils.PutBytes(native, fsSettingKey, info.Bytes())
}

// get Fs setting with provided prove level, now prove level only impact the prove interval
func getFsSettingWithProveLevel(native *native.NativeService, proveLevel uint64) (*FsSetting, error) {
	fsSetting, err := getFsSetting(native)
	if err != nil {
		return nil, err
	}

	fsSetting.DefaultProvePeriod = GetProveIntervalByProveLevel(proveLevel)
	return fsSetting, nil
}

func GetProveIntervalByProveLevel(proveLevel uint64) uint64 {
	switch proveLevel {
	case PROVE_LEVEL_HIGH:
		return PROVE_PERIOD_HIGHT
	case PROVE_LEVEL_MEDIEUM:
		return PROVE_PERIOD_MEDIEUM
	case PROVE_LEVEL_LOW:
		return PROVE_PERIOD_LOW
	default:
		return PROVE_PERIOD_HIGHT
	}
}
