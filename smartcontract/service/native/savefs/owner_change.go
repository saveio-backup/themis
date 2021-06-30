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

type OwnerChange struct {
	FileHash []byte
	CurOwner common.Address
	NewOwner common.Address
}

func (this *OwnerChange) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.FileHash); err != nil {
		return fmt.Errorf("[OwnerChange] [FileHash:%v] serialize from error:%v", this.FileHash, err)
	}
	if err := utils.WriteAddress(w, this.CurOwner); err != nil {
		return fmt.Errorf("[OwnerChange] [CurOwner:%v] serialize from error:%v", this.CurOwner, err)
	}
	if err := utils.WriteAddress(w, this.NewOwner); err != nil {
		return fmt.Errorf("[OwnerChange] [NewOwner:%v] serialize from error:%v", this.NewOwner, err)
	}
	return nil
}

func (this *OwnerChange) Deserialize(r io.Reader) error {
	var err error
	if this.FileHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[OwnerChange] [FileHash] deserialize from error:%v", err)
	}
	if this.CurOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[OwnerChange] [CurOwner] deserialize from error:%v", err)
	}
	if this.NewOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[OwnerChange] [NewOwner] deserialize from error:%v", err)
	}
	return nil
}

func (this *OwnerChange) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.FileHash)
	utils.EncodeAddress(sink, this.CurOwner)
	utils.EncodeAddress(sink, this.NewOwner)
}

func (this *OwnerChange) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.CurOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.NewOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	return nil
}
