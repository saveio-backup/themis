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

package utils

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/serialization"
	"github.com/saveio/themis/vm/neovm/types"
)

func WriteVarUint(w io.Writer, value uint64) error {
	if err := serialization.WriteVarBytes(w, types.BigIntToBytes(big.NewInt(int64(value)))); err != nil {
		return fmt.Errorf("serialize value error:%v", err)
	}
	return nil
}

func ReadVarUint(r io.Reader) (uint64, error) {
	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return 0, fmt.Errorf("deserialize value error:%v", err)
	}
	v := types.BigIntFromBytes(value)
	if v.Cmp(big.NewInt(0)) < 0 {
		return 0, fmt.Errorf("%s", "value should not be a negative number.")
	}
	return v.Uint64(), nil
}

func WriteAddress(w io.Writer, address common.Address) error {
	if err := serialization.WriteVarBytes(w, address[:]); err != nil {
		return fmt.Errorf("serialize value error:%v", err)
	}
	return nil
}

func ReadAddress(r io.Reader) (common.Address, error) {
	from, err := serialization.ReadVarBytes(r)
	if err != nil {
		return common.Address{}, fmt.Errorf("[State] deserialize from error:%v", err)
	}
	return common.AddressParseFromBytes(from)
}

func EncodeAddress(sink *common.ZeroCopySink, addr common.Address) (size uint64) {
	return sink.WriteVarBytes(addr[:])
}

func EncodeVarUint(sink *common.ZeroCopySink, value uint64) (size uint64) {
	return sink.WriteVarBytes(types.BigIntToBytes(big.NewInt(int64(value))))
}

func DecodeVarUint(source *common.ZeroCopySource) (uint64, error) {
	value, _, irregular, eof := source.NextVarBytes()
	if eof {
		return 0, io.ErrUnexpectedEOF
	}
	if irregular {
		return 0, common.ErrIrregularData
	}
	v := types.BigIntFromBytes(value)
	if v.Cmp(big.NewInt(0)) < 0 {
		return 0, fmt.Errorf("%s", "value should not be a negative number.")
	}
	return v.Uint64(), nil
}

func DecodeAddress(source *common.ZeroCopySource) (common.Address, error) {
	from, _, irregular, eof := source.NextVarBytes()
	if eof {
		return common.Address{}, io.ErrUnexpectedEOF
	}
	if irregular {
		return common.Address{}, common.ErrIrregularData
	}

	return common.AddressParseFromBytes(from)
}

func WriteBytes(w io.Writer, b []byte) error {
	if err := serialization.WriteVarBytes(w, b[:]); err != nil {
		return fmt.Errorf("serialize value error:%v", err)
	}
	return nil
}

func ReadBytes(r io.Reader) ([]byte, error) {
	from, err := serialization.ReadVarBytes(r)
	if err != nil {
		return nil, fmt.Errorf("[State] deserialize from error:%v", err)
	}
	return from, nil
}

func EncodeBytes(sink *common.ZeroCopySink, b []byte) (size uint64) {
	return sink.WriteVarBytes(b[:])
}

func DecodeBytes(source *common.ZeroCopySource) ([]byte, error) {
	from, _, irregular, eof := source.NextVarBytes()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}
	if irregular {
		return nil, common.ErrIrregularData
	}

	return from, nil
}

func WriteBool(w io.Writer, b bool) error {
	var val []byte
	if b {
		val = BYTE_TRUE
	} else {
		val = BYTE_FALSE
	}
	if err := serialization.WriteVarBytes(w, val); err != nil {
		return fmt.Errorf("serialize value error:%v", err)
	}
	return nil
}

func ReadBool(r io.Reader) (bool, error) {
	from, err := serialization.ReadVarBytes(r)
	if err != nil {
		return false, fmt.Errorf("[State] deserialize from error:%v", err)
	}
	if bytes.Compare(from, BYTE_TRUE) == 0 {
		return true, nil
	}
	return false, nil
}

func EncodeBool(sink *common.ZeroCopySink, b bool) {
	if b {
		sink.WriteVarBytes(BYTE_TRUE)
	} else {
		sink.WriteVarBytes(BYTE_FALSE)
	}
}

func DecodeBool(source *common.ZeroCopySource) (bool, error) {
	data, _, irregular, eof := source.NextVarBytes()
	var from bool
	if bytes.Compare(data, BYTE_TRUE) == 0 {
		from = true
	}
	if eof {
		return false, io.ErrUnexpectedEOF
	}
	if irregular {
		return false, common.ErrIrregularData
	}

	return from, nil
}
