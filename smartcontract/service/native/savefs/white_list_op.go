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

	"github.com/saveio/themis/smartcontract/service/native/utils"
)

const (
	ADD     = 0 //Will cover rule with same key
	DEL     = 1
	ADD_COV = 2 //Delete all old rules and add new rules
	DEL_ALL = 3 //Delete all rules
	UPDATE  = 4
)

type WhiteListOp struct {
	FileHash []byte
	Op       uint64
	List     WhiteList
}

func (this *WhiteListOp) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.FileHash); err != nil {
		return fmt.Errorf("[WhiteListOp] [FileHash:%v] serialize from error:%v", this.FileHash, err)
	}
	if err := utils.WriteVarUint(w, this.Op); err != nil {
		return fmt.Errorf("[WhiteListOp] [Op:%v] serialize from error:%v", this.Op, err)
	}
	if err := this.List.Serialize(w); err != nil {
		return fmt.Errorf("[WhiteListOp] [List] serialize from error:%v", err)
	}
	return nil
}

func (this *WhiteListOp) Deserialize(r io.Reader) error {
	var err error
	if this.FileHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[WhiteListOp] [FileHash] deserialize from error:%v", err)
	}
	if this.Op, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[WhiteListOp] [Op] deserialize from error:%v", err)
	}
	if err = this.List.Deserialize(r); err != nil {
		return fmt.Errorf("[WhiteListOp] [List] deserialize from error:%v", err)
	}
	return nil
}
