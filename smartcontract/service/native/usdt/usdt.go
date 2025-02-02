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

package usdt

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/constants"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

const (
	TRANSFER_FLAG byte = 1
	APPROVE_FLAG  byte = 2
)

func InitUsdt() {
	native.Contracts[utils.UsdtContractAddress] = RegisterOntContract
}

func RegisterOntContract(native *native.NativeService) {
	native.Register(INIT_NAME, UsdtInit)
	native.Register(TRANSFER_NAME, UsdtTransfer)
	native.Register(APPROVE_NAME, UsdtApprove)
	native.Register(TRANSFERFROM_NAME, UsdtTransferFrom)
	native.Register(NAME_NAME, UsdtName)
	native.Register(SYMBOL_NAME, UsdtSymbol)
	native.Register(DECIMALS_NAME, UsdtDecimals)
	native.Register(TOTALSUPPLY_NAME, UsdtTotalSupply)
	native.Register(BALANCEOF_NAME, UsdtBalanceOf)
	native.Register(ALLOWANCE_NAME, UsdtAllowance)
}

func UsdtInit(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount > 0 {
		return utils.BYTE_FALSE, errors.NewErr("Init usdt has been completed!")
	}

	distribute := make(map[common.Address]uint64)
	source := common.NewZeroCopySource(native.Input)
	buf, _, irregular, eof := source.NextVarBytes()
	if eof {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarBytes, contract params deserialize error!")
	}
	if irregular {
		return utils.BYTE_FALSE, common.ErrIrregularData
	}
	input := common.NewZeroCopySource(buf)
	num, err := utils.DecodeVarUint(input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("read number error:%v", err)
	}
	sum := uint64(0)
	overflow := false
	for i := uint64(0); i < num; i++ {
		addr, err := utils.DecodeAddress(input)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("read address error:%v", err)
		}
		value, err := utils.DecodeVarUint(input)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("read value error:%v", err)
		}
		sum, overflow = common.SafeAdd(sum, value)
		if overflow {
			return utils.BYTE_FALSE, errors.NewErr("wrong config. overflow detected")
		}
		distribute[addr] += value
	}
	if sum != constants.USDT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("wrong config. total supply %d != %d", sum, constants.USDT_TOTAL_SUPPLY)
	}

	for addr, val := range distribute {
		balanceKey := GenBalanceKey(contract, addr)
		item := utils.GenUInt64StorageItem(val)
		native.CacheDB.Put(balanceKey, item.ToArray())
		AddNotifications(native, contract, &State{To: addr, Value: val})
	}
	native.CacheDB.Put(GenTotalSupplyKey(contract), utils.GenUInt64StorageItem(constants.USDT_TOTAL_SUPPLY).ToArray())

	return utils.BYTE_TRUE, nil
}

func UsdtTransfer(native *native.NativeService) ([]byte, error) {
	var transfers Transfers
	source := common.NewZeroCopySource(native.Input)
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] Transfers deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	for _, v := range transfers.States {
		if v.Value == 0 {
			continue
		}
		if v.Value > constants.USDT_TOTAL_SUPPLY {
			return utils.BYTE_FALSE, fmt.Errorf("transfer usdt amount:%d over totalSupply:%d", v.Value, constants.USDT_TOTAL_SUPPLY)
		}

		//fromBalance, toBalance, err := Transfer(native, contract, &v)
		_, _, err := Transfer(native, contract, &v)
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		//skip ONG logic
		//if err := grantOng(native, contract, v.From, fromBalance); err != nil {
		//	return utils.BYTE_FALSE, err
		//}

		//if err := grantOng(native, contract, v.To, toBalance); err != nil {
		//	return utils.BYTE_FALSE, err
		//}

		AddNotifications(native, contract, &v)
	}
	return utils.BYTE_TRUE, nil
}

func UsdtTransferFrom(native *native.NativeService) ([]byte, error) {
	var state TransferFrom
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[UsdtTransferFrom] State deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.USDT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("transferFrom usdt amount:%d over totalSupply:%d", state.Value, constants.USDT_TOTAL_SUPPLY)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	_, _, err := TransferedFrom(native, contract, &state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	// if err := grantOng(native, contract, state.From, fromBalance); err != nil {
	// 	return utils.BYTE_FALSE, err
	// }
	// if err := grantOng(native, contract, state.To, toBalance); err != nil {
	// 	return utils.BYTE_FALSE, err
	// }
	AddNotifications(native, contract, &State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}

func UsdtApprove(native *native.NativeService) ([]byte, error) {
	var state State
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.USDT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve usdt amount:%d over totalSupply:%d", state.Value, constants.USDT_TOTAL_SUPPLY)
	}
	if native.ContextRef.CheckWitness(state.From) == false {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CacheDB.Put(GenApproveKey(contract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}

func UsdtName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.USDT_NAME), nil
}

func UsdtDecimals(native *native.NativeService) ([]byte, error) {
	return common.BigIntToNeoBytes(big.NewInt(int64(constants.USDT_DECIMALS))), nil
}

func UsdtSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.USDT_SYMBOL), nil
}

func UsdtTotalSupply(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[UsdtTotalSupply] get totalSupply error!")
	}
	return common.BigIntToNeoBytes(big.NewInt(int64(amount))), nil
}

func UsdtBalanceOf(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, TRANSFER_FLAG)
}

func UsdtAllowance(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, APPROVE_FLAG)
}

func GetBalanceValue(native *native.NativeService, flag byte) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	from, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get from address error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	var key []byte
	if flag == APPROVE_FLAG {
		to, err := utils.DecodeAddress(source)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get from address error!")
		}
		key = GenApproveKey(contract, from, to)
	} else if flag == TRANSFER_FLAG {
		key = GenBalanceKey(contract, from)
	}
	amount, err := utils.GetStorageUInt64(native, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] address parse error!")
	}

	return common.BigIntToNeoBytes(big.NewInt(int64(amount))), nil
}

func grantOng(native *native.NativeService, contract, address common.Address, balance uint64) error {
	// startOffset, err := getUnboundOffset(native, contract, address)
	// if err != nil {
	// 	return err
	// }
	// if native.Time <= constants.GENESIS_BLOCK_TIMESTAMP {
	// 	return nil
	// }
	// endOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP
	// if endOffset < startOffset {
	// 	errstr := fmt.Sprintf("grant Ong error: wrong timestamp endOffset: %d < startOffset: %d", endOffset, startOffset)
	// 	log.Error(errstr)
	// 	return errors.NewErr(errstr)
	// } else if endOffset == startOffset {
	// 	return nil
	// }

	// if balance != 0 {
	// 	value := utils.CalcUnbindOng(balance, startOffset, endOffset)

	// 	args, err := getApproveArgs(native, contract, utils.OngContractAddress, address, value)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if _, err := native.NativeCall(utils.OngContractAddress, "approve", args); err != nil {
	// 		return err
	// 	}
	// }

	// native.CacheDB.Put(genAddressUnboundOffsetKey(contract, address), utils.GenUInt32StorageItem(endOffset).ToArray())
	return nil
}

func getApproveArgs(native *native.NativeService, contract, ongContract, address common.Address, value uint64) ([]byte, error) {
	bf := new(bytes.Buffer)
	approve := State{
		From:  contract,
		To:    address,
		Value: value,
	}

	stateValue, err := utils.GetStorageUInt64(native, GenApproveKey(ongContract, approve.From, approve.To))
	if err != nil {
		return nil, err
	}

	approve.Value += stateValue

	if err := approve.Serialize(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
