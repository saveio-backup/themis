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

package neovm

import (
	"fmt"

	"github.com/saveio/themis/common"
	vm "github.com/saveio/themis/vm/neovm"
)

// StorageContext store smart contract address
type StorageContext struct {
	Address    common.Address
	IsReadOnly bool
}

// NewStorageContext return a new smart contract storage context
func NewStorageContext(address common.Address) *StorageContext {
	var storageContext StorageContext
	storageContext.Address = address
	storageContext.IsReadOnly = false
	return &storageContext
}

// ToArray return address byte array
func (this *StorageContext) ToArray() []byte {
	return this.Address[:]
}

func StorageContextAsReadOnly(service *NeoVmService, engine *vm.Executor) error {
	data, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	context, ok := data.Data.(*StorageContext)
	if !ok {
		return fmt.Errorf("%s", "pop storage context type invalid")
	}
	if !context.IsReadOnly {
		context = NewStorageContext(context.Address)
		context.IsReadOnly = true
	}
	return engine.EvalStack.PushAsInteropValue(context)
}
