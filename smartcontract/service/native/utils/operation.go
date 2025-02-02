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

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/serialization"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/event"
	"github.com/saveio/themis/smartcontract/service/native"
)

func AddCommonEvent(native *native.NativeService, contract common.Address, name string, params interface{}) {
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{name, params},
		})
}

func ConcatKey(contract common.Address, args ...[]byte) []byte {
	temp := contract[:]
	for _, arg := range args {
		temp = append(temp, arg...)
	}
	return temp
}

func ConcatBytes(args ...[]byte) []byte {
	temp := []byte{}
	for _, arg := range args {
		temp = append(temp, arg...)
	}
	return temp
}

func ValidateOwner(native *native.NativeService, address common.Address) error {
	if !native.ContextRef.CheckWitness(address) {
		return errors.NewErr("validateOwner, authentication failed!")
	}
	return nil
}

func GetUint32Bytes(num uint32) ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteUint32(bf, num); err != nil {
		return nil, fmt.Errorf("serialization.WriteUint32, serialize uint32 error: %v", err)
	}
	return bf.Bytes(), nil
}

func GetBytesUint32(b []byte) (uint32, error) {
	num, err := serialization.ReadUint32(bytes.NewBuffer(b))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint32, deserialize uint32 error: %v", err)
	}
	return num, nil
}

func GetUint64Bytes(num uint64) ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteUint64(bf, num); err != nil {
		return nil, fmt.Errorf("serialization.WriteUint64, serialize uint64 error: %v", err)
	}
	return bf.Bytes(), nil
}

func GetBytesUint64(b []byte) (uint64, error) {
	num, err := serialization.ReadUint64(bytes.NewBuffer(b))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint64, deserialize uint64 error: %v", err)
	}
	return num, nil
}
