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
	"github.com/saveio/themis/common/serialization"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
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

func getUserSpace(native *native.NativeService, addr common.Address) (*UserSpace, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	userSpaceKey := GenFsUserSpaceKey(contract, addr)
	userSpaceItem, err := utils.GetStorageItem(native, userSpaceKey)
	if err != nil || userSpaceItem == nil {
		return nil, errors.NewErr("Userspace not found!")
	}

	var userspace UserSpace
	reader := bytes.NewReader(userSpaceItem.Value)
	err = userspace.Deserialize(reader)
	if err != nil {
		return nil, errors.NewErr("GetUserSpace deserialize error!")
	}

	return &userspace, nil
}
func setUserSpace(native *native.NativeService, userspace *UserSpace, addr common.Address) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	bf := new(bytes.Buffer)
	if err := userspace.Serialize(bf); err != nil {
		return errors.NewErr("Userspace serialize error!")
	}
	userSpaceKey := GenFsUserSpaceKey(contract, addr)
	utils.PutBytes(native, userSpaceKey, bf.Bytes())
	return nil
}

// get saved user space, nil if not found
func getOldUserSpace(native *native.NativeService, addr common.Address) (*UserSpace, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	userSpaceKey := GenFsUserSpaceKey(contract, addr)
	item, err := utils.GetStorageItem(native, userSpaceKey)
	if err != nil {
		return nil, errors.NewErr("GetUserSpace GetStorageItem error!")
	}

	if item == nil || len(item.Value) == 0 {
		return nil, nil
	}

	var userSpace UserSpace
	reader := bytes.NewReader(item.Value)
	if err = userSpace.Deserialize(reader); err != nil {
		return nil, errors.NewErr("GetUserSpace deserialize error!")
	}

	return &userSpace, nil
}

const (
	// user space operation for size and block count
	UserSpaceOps_None_None     = uint64(UserSpaceNone<<4 | UserSpaceNone)
	UserspaceOps_None_Add      = uint64(UserSpaceNone<<4 | UserSpaceAdd)
	UserspaceOps_None_Revoke   = uint64(UserSpaceNone<<4 | UserSpaceRevoke)
	UserspaceOps_Add_None      = uint64(UserSpaceAdd<<4 | UserSpaceNone)
	UserspaceOps_Add_Add       = uint64(UserSpaceAdd<<4 | UserSpaceAdd)
	UserspaceOps_Add_Revoke    = uint64(UserSpaceAdd<<4 | UserSpaceRevoke)
	UserspaceOps_Revoke_None   = uint64(UserSpaceRevoke<<4 | UserSpaceNone)
	UserspaceOps_Revoke_Add    = uint64(UserSpaceRevoke<<4 | UserSpaceAdd)
	UserspaceOps_Revoke_Revoke = uint64(UserSpaceRevoke<<4 | UserSpaceRevoke)
)

func isValidUserSpaceOperation(op *UserSpaceOperation) bool {
	switch UserSpaceType(op.Type) {
	case UserSpaceRevoke, UserSpaceAdd:
		if op.Value == 0 {
			return false
		}
		return true
	case UserSpaceNone:
		if op.Value != 0 {
			return false
		}
		return true
	default:
		return false
	}
}

func combineUserSpaceTypes(t1, t2 UserSpaceType) uint64 {
	t := (byte)(t1)<<4 | (byte)(t2)
	return uint64(t)
}

// check if there is revoke operation for revoke
func isRevokeUserSpace(params *UserSpaceParams) bool {
	return UserSpaceType(params.Size.Type) == UserSpaceRevoke ||
		UserSpaceType(params.BlockCount.Type) == UserSpaceRevoke
}

func getUserSpaceOperationsFromParams(params *UserSpaceParams) (uint64, error) {
	if params.Size == nil || params.BlockCount == nil {
		return 0, errors.NewErr("UserSpaceParams size or block count is nil")
	}

	if !isValidUserSpaceOperation(params.Size) || !isValidUserSpaceOperation(params.BlockCount) {
		return 0, errors.NewErr("UserSpaceParams invalid user space operation")
	}

	sizeType := UserSpaceType(params.Size.Type)
	blockCountType := UserSpaceType(params.BlockCount.Type)
	return combineUserSpaceTypes(sizeType, blockCountType), nil
}
