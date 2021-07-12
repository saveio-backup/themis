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

package governance

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common/serialization"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

//sip
type RegisterSipParam struct {
	Height   uint32
	Detail   []byte
	Default  byte
	MinVotes uint32
	Bonus    uint64
}

func (this *RegisterSipParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.Height)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize Height error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Detail); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize Detail error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Default)); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize Default error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.MinVotes)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize MinVotes error: %v", err)
	}
	if err := utils.WriteVarUint(w, this.Bonus); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize Bonus error: %v", err)
	}
	return nil
}

func (this *RegisterSipParam) Deserialize(r io.Reader) error {
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize height error: %v", err)
	}
	detail, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize detail error: %v", err)
	}
	result, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize default error: %v", err)
	}
	minVotes, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize height error: %v", err)
	}
	bonus, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize bonus error: %v", err)
	}

	this.Height = uint32(height)
	this.Detail = detail
	this.Default = byte(result)
	this.MinVotes = uint32(minVotes)
	this.Bonus = bonus

	return nil
}

type QuerySipParam struct {
	Index uint32
}

func (this *QuerySipParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.Index)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize Index error: %v", err)
	}
	return nil
}

func (this *QuerySipParam) Deserialize(r io.Reader) error {
	index, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize index error: %v", err)
	}

	this.Index = uint32(index)

	return nil
}
