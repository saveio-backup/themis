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
package testsuite

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/core/payload"
	"github.com/saveio/themis/core/types"
	utils2 "github.com/saveio/themis/core/utils"
	"github.com/saveio/themis/smartcontract"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/storage"
	"github.com/stretchr/testify/assert"
)

func RandomAddress() common.Address {
	var addr common.Address
	_, _ = rand.Read(addr[:])

	return addr
}

func InvokeNativeContract(t *testing.T, addr common.Address, handler native.Handler) {
	buf := make([]byte, 100)
	_, _ = rand.Read(buf)
	method := hex.EncodeToString(buf)
	actions := make(map[string]native.Handler)
	actions[method] = handler
	AppendNativeContract(addr, actions)

	tx := BuildInvokeTx(addr, method, []interface{}{""})
	assert.NotNil(t, tx)

	overlay := NewOverlayDB()
	cache := storage.NewCacheDB(overlay)

	_, err := executeTransaction(tx, cache)

	assert.Nil(t, err)
}

func AppendNativeContract(addr common.Address, actions map[string]native.Handler) {
	origin, ok := native.Contracts[addr]

	contract := func(native *native.NativeService) {
		if ok {
			origin(native)
		}
		for name, fun := range actions {
			native.Register(name, fun)
		}
	}
	native.Contracts[addr] = contract
}

func executeTransaction(tx *types.Transaction, cache *storage.CacheDB) (interface{}, error) {
	config := &smartcontract.Config{
		Time: uint32(time.Now().Unix()),
		Tx:   tx,
	}

	if tx.TxType == types.InvokeNeo {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:  config,
			Store:   nil,
			CacheDB: cache,
			Gas:     100000000000000,
			PreExec: true,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code, tx.TxType)
		res, err := engine.Invoke()
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	panic("unimplemented")
}

func BuildInvokeTx(contractAddress common.Address, method string,
	args []interface{}) *types.Transaction {
	invokCode, err := utils2.BuildNativeInvokeCode(contractAddress, 0, method, args)
	if err != nil {
		return nil
	}
	invokePayload := &payload.InvokeCode{
		Code: invokCode,
	}
	tx := &types.MutableTransaction{
		Version:  0,
		GasPrice: 0,
		GasLimit: 1000000000,
		TxType:   types.InvokeNeo,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     make([]types.Sig, 0, 0),
	}
	res, err := tx.IntoImmutable()
	if err != nil {
		return nil
	}
	return res
}
