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

type PriChange struct {
	FileHash  []byte
	Privilege uint64
}

func (this *PriChange) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.FileHash); err != nil {
		return fmt.Errorf("[PriChange] [FileHash:%v] serialize from error:%v", this.FileHash, err)
	}
	if err := utils.WriteVarUint(w, this.Privilege); err != nil {
		return fmt.Errorf("[PriChange] [Privilege:%v] serialize from error:%v", this.Privilege, err)
	}
	return nil
}

func (this *PriChange) Deserialize(r io.Reader) error {
	var err error
	if this.FileHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[PriChange] [FileHash] deserialize from error:%v", err)
	}
	if this.Privilege, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[PriChange] [Privilege] deserialize from error:%v", err)
	}
	return nil
}

func (this *PriChange) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.FileHash)
	utils.EncodeVarUint(sink, this.Privilege)
}

func (this *PriChange) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Privilege, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}
