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
	"github.com/saveio/themis/common/serialization"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

// UserSpace. used to stored user space
type UserSpace struct {
	Used         uint64 // used space
	Remain       uint64 // remain space, equal to blockNum*blockSize
	ExpireHeight uint64 // expired block height
	Balance      uint64 // balance of asset
	UpdateHeight uint64 // update block height
}

func (this *UserSpace) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.Used); err != nil {
		return fmt.Errorf("[UserSpace] [Used:%v] serialize from error:%v", this.Used, err)
	}
	if err := utils.WriteVarUint(w, this.Remain); err != nil {
		return fmt.Errorf("[UserSpace] [Remain:%v] serialize from error:%v", this.Remain, err)
	}
	if err := utils.WriteVarUint(w, this.ExpireHeight); err != nil {
		return fmt.Errorf("[UserSpace] [ExpireHeight:%v] serialize from error:%v", this.ExpireHeight, err)
	}
	if err := utils.WriteVarUint(w, this.Balance); err != nil {
		return fmt.Errorf("[UserSpace] [Balance:%v] serialize from error:%v", this.Balance, err)
	}
	if err := utils.WriteVarUint(w, this.UpdateHeight); err != nil {
		return fmt.Errorf("[UserSpace] [UpdateHeight:%v] serialize from error:%v", this.UpdateHeight, err)
	}
	return nil
}

func (this *UserSpace) Deserialize(r io.Reader) error {
	var err error
	if this.Used, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UserSpace] [Used] Deserialize from error:%v", err)
	}
	if this.Remain, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UserSpace] [Remain] Deserialize from error:%v", err)
	}
	if this.ExpireHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UserSpace] [ExpireHeight] Deserialize from error:%v", err)
	}
	if this.Balance, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UserSpace] [Balance] Deserialize from error:%v", err)
	}
	if this.UpdateHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UserSpace] [UpdateHeight] Deserialize from error:%v", err)
	}
	return nil
}

func (this *UserSpace) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.Used)
	utils.EncodeVarUint(sink, this.Remain)
	utils.EncodeVarUint(sink, this.ExpireHeight)
	utils.EncodeVarUint(sink, this.Balance)
}

func (this *UserSpace) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Used, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Remain, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ExpireHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Balance, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return err
}

// UserSpaceParams. used to update user spaces
type UserSpaceType uint64

const (
	UserSpaceNone UserSpaceType = iota
	UserSpaceAdd
	UserSpaceRevoke
)

type UserSpaceOperation struct {
	Type  uint64
	Value uint64
}

func (this *UserSpaceOperation) Serialize(w io.Writer) error {
	if err := serialization.WriteUint64(w, uint64(this.Type)); err != nil {
		return fmt.Errorf("[UserSpaceOperation] [Type:%v] serialize from error:%v", this.Type, err)
	}
	if err := serialization.WriteUint64(w, this.Value); err != nil {
		return fmt.Errorf("[UserSpaceOperation] [Value:%v] serialize from error:%v", this.Value, err)
	}
	return nil
}

func (this *UserSpaceOperation) Deserialize(r io.Reader) error {
	var err error
	if this.Type, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UserSpaceOperation] [Type] Deserialize from error:%v", err)
	}
	if this.Value, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UserSpaceOperation] [Value] Deserialize from error:%v", err)
	}
	return nil
}

func (this *UserSpaceOperation) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, uint64(this.Type))
	utils.EncodeVarUint(sink, this.Value)
}
func (this *UserSpaceOperation) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Type, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Value, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

type UserSpaceParams struct {
	WalletAddr common.Address      // depositer address
	Owner      common.Address      // target address
	Size       *UserSpaceOperation // added / revoke size
	BlockCount *UserSpaceOperation // added / revoke block count
}

func (this *UserSpaceParams) Serialize(w io.Writer) error {

	if err := serialization.WriteVarBytes(w, this.WalletAddr[:]); err != nil {
		return fmt.Errorf("[UserSpace] [WalletAddr:%v] serialize from error:%v", this.WalletAddr, err)
	}
	if err := serialization.WriteVarBytes(w, this.Owner[:]); err != nil {
		return fmt.Errorf("[UserSpace] [Owner:%v] serialize from error:%v", this.Owner, err)
	}
	if err := this.Size.Serialize(w); err != nil {
		return fmt.Errorf("[UserSpace] [Size:%v] serialize from error:%v", this.Size, err)
	}
	if err := this.BlockCount.Serialize(w); err != nil {
		return fmt.Errorf("[UserSpace] [BlockCount:%v] serialize from error:%v", this.BlockCount, err)
	}
	return nil
}

func (this *UserSpaceParams) Deserialize(r io.Reader) error {
	var err error
	if this.WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[UserSpace] [WalletAddr] Deserialize from error:%v", err)
	}
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[UserSpace] [Owner] Deserialize from error:%v", err)
	}
	if err := this.Size.Deserialize(r); err != nil {
		return fmt.Errorf("[UserSpace] [Size] Deserialize from error:%v", err)
	}
	if err := this.BlockCount.Deserialize(r); err != nil {
		return fmt.Errorf("[UserSpace] [BlockCount] Deserialize from error:%v", err)
	}
	return nil
}

func (this *UserSpaceParams) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.WalletAddr)
	utils.EncodeAddress(sink, this.Owner)
	this.Size.Serialization(sink)
	this.BlockCount.Serialization(sink)
}

func (this *UserSpaceParams) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Owner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Size = &UserSpaceOperation{}
	err = this.Size.Deserialization(source)
	if err != nil {
		return err
	}
	this.BlockCount = &UserSpaceOperation{}
	err = this.BlockCount.Deserialization(source)
	if err != nil {
		return err
	}
	return err
}
