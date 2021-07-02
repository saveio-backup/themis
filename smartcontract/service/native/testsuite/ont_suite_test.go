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
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/constants"
	"github.com/saveio/themis/smartcontract/service/native"
	_ "github.com/saveio/themis/smartcontract/service/native/init"
	"github.com/saveio/themis/smartcontract/service/native/usdt"
	"github.com/saveio/themis/smartcontract/service/native/utils"
	"github.com/saveio/themis/smartcontract/storage"
	"github.com/stretchr/testify/assert"

	"testing"
)

func setOntBalance(db *storage.CacheDB, addr common.Address, value uint64) {
	balanceKey := usdt.GenBalanceKey(utils.UsdtContractAddress, addr)
	item := utils.GenUInt64StorageItem(value)
	db.Put(balanceKey, item.ToArray())
}

func ontBalanceOf(native *native.NativeService, addr common.Address) int {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := usdt.UsdtBalanceOf(native)
	val := common.BigIntFromNeoBytes(buf)
	return int(val.Uint64())
}

func ontTotalAllowance(native *native.NativeService, addr common.Address) int {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := usdt.UsdtAllowance(native)
	val := common.BigIntFromNeoBytes(buf)
	return int(val.Uint64())
}

func ontTransfer(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	state := usdt.State{from, to, value}
	native.Input = common.SerializeToBytes(&usdt.Transfers{States: []usdt.State{state}})

	_, err := usdt.UsdtTransfer(native)
	return err
}

func ontApprove(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	native.Input = common.SerializeToBytes(&usdt.State{from, to, value})

	_, err := usdt.UsdtApprove(native)
	return err
}

func TestTransfer(t *testing.T) {
	InvokeNativeContract(t, utils.UsdtContractAddress, func(native *native.NativeService) ([]byte, error) {
		a := RandomAddress()
		b := RandomAddress()
		c := RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)

		assert.Equal(t, ontBalanceOf(native, a), 10000)
		assert.Equal(t, ontBalanceOf(native, b), 0)
		assert.Equal(t, ontBalanceOf(native, c), 0)

		assert.Nil(t, ontTransfer(native, a, b, 10))
		assert.Equal(t, ontBalanceOf(native, a), 9990)
		assert.Equal(t, ontBalanceOf(native, b), 10)

		assert.Nil(t, ontTransfer(native, b, c, 10))
		assert.Equal(t, ontBalanceOf(native, b), 0)
		assert.Equal(t, ontBalanceOf(native, c), 10)

		return nil, nil
	})
}

func TestTotalAllowance(t *testing.T) {
	InvokeNativeContract(t, utils.UsdtContractAddress, func(native *native.NativeService) ([]byte, error) {
		a := RandomAddress()
		b := RandomAddress()
		c := RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)

		assert.Equal(t, ontBalanceOf(native, a), 10000)
		assert.Equal(t, ontBalanceOf(native, b), 0)
		assert.Equal(t, ontBalanceOf(native, c), 0)

		assert.Nil(t, ontApprove(native, a, b, 10))
		assert.Equal(t, ontTotalAllowance(native, a), 10)
		assert.Equal(t, ontTotalAllowance(native, b), 0)

		assert.Nil(t, ontApprove(native, a, c, 100))
		assert.Equal(t, ontTotalAllowance(native, a), 110)
		assert.Equal(t, ontTotalAllowance(native, c), 0)

		return nil, nil
	})
}

func TestGovernanceUnbound(t *testing.T) {
	InvokeNativeContract(t, utils.UsdtContractAddress, func(native *native.NativeService) ([]byte, error) {
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.USDT_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1

		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))

		return nil, nil
	})

	InvokeNativeContract(t, utils.UsdtContractAddress, func(native *native.NativeService) ([]byte, error) {
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.USDT_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))

		return nil, nil
	})

	InvokeNativeContract(t, utils.UsdtContractAddress, func(native *native.NativeService) ([]byte, error) {
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.USDT_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))

		return nil, nil
	})

	InvokeNativeContract(t, utils.UsdtContractAddress, func(native *native.NativeService) ([]byte, error) {
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.USDT_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1
		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))
		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 10000
		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))
		native.Time = config.GetOntHolderUnboundDeadline() - 100
		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))

		return nil, nil
	})
}
