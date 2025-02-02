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

	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/signature"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/event"
	vm "github.com/saveio/themis/vm/neovm"
	vmtypes "github.com/saveio/themis/vm/neovm/types"
)

// HeaderGetNextConsensus put current block time to vm stack
func RuntimeGetTime(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushInt64(int64(service.Time))
}

// RuntimeCheckWitness provide check permissions service
// If param address isn't exist in authorization list, check fail
func RuntimeCheckWitness(service *NeoVmService, engine *vm.Executor) error {
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	var result bool
	if len(data) == 20 {
		address, err := common.AddressParseFromBytes(data)
		if err != nil {
			return err
		}
		result = service.ContextRef.CheckWitness(address)
	} else {
		pk, err := keypair.DeserializePublicKey(data)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[RuntimeCheckWitness] data invalid.")
		}
		result = service.ContextRef.CheckWitness(types.AddressFromPubKey(pk))
	}

	return engine.EvalStack.PushBool(result)
}

func RuntimeSerialize(service *NeoVmService, engine *vm.Executor) error {
	val, err := engine.EvalStack.Pop()
	if err != nil {
		return err
	}
	sink := new(common.ZeroCopySink)
	err = val.Serialize(sink)
	if err != nil {
		return err
	}
	return engine.EvalStack.PushBytes(sink.Bytes())
}

//TODO check consistency with original implementation
func RuntimeDeserialize(service *NeoVmService, engine *vm.Executor) error {
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return fmt.Errorf("[RuntimeDeserialize] PopAsBytes error: %s", err)
	}
	source := common.NewZeroCopySource(data)
	vmValue := vmtypes.VmValue{}
	err = vmValue.Deserialize(source)
	if err != nil {
		return fmt.Errorf("[RuntimeDeserialize] Deserialize error: %s", err)
	}
	return engine.EvalStack.Push(vmValue)
}

func RuntimeVerifyMutiSig(service *NeoVmService, engine *vm.Executor) error {
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	arr1, err := engine.EvalStack.PopAsArray()
	if err != nil {
		return err
	}
	pks := make([]keypair.PublicKey, 0, len(arr1.Data))
	for i := 0; i < len(arr1.Data); i++ {
		value, err := arr1.Data[i].AsBytes()
		if err != nil {
			return err
		}
		pk, err := keypair.DeserializePublicKey(value)
		if err != nil {
			return err
		}
		pks = append(pks, pk)
	}

	m, err := engine.EvalStack.PopAsInt64()
	if err != nil {
		return err
	}
	if m > int64(len(pks)) || m < 0 {
		return fmt.Errorf("runtime verify multisig error: wrong m %d", m)

	}
	arr2, err := engine.EvalStack.PopAsArray()
	if err != nil {
		return err
	}
	signs := make([][]byte, 0, len(arr2.Data))
	for i := 0; i < len(arr2.Data); i++ {
		value, err := arr2.Data[i].AsBytes()
		if err != nil {
			return err
		}
		signs = append(signs, value)
	}
	err = signature.VerifyMultiSignature(data, pks, int(m), signs)
	return engine.EvalStack.PushBool(err == nil)
}

// RuntimeNotify put smart contract execute event notify to notifications
func RuntimeNotify(service *NeoVmService, engine *vm.Executor) error {
	item, err := engine.EvalStack.Pop()
	if err != nil {
		return err
	}

	context := service.ContextRef.CurrentContext()
	states, err := item.ConvertNeoVmValueHexString()
	if err != nil {
		return err
	}
	service.Notifications = append(service.Notifications, &event.NotifyEventInfo{ContractAddress: context.ContractAddress, States: states})
	return nil
}

// RuntimeLog push smart contract execute event log to client
func RuntimeLog(service *NeoVmService, engine *vm.Executor) error {
	sitem, err := engine.EvalStack.Peek(0)
	if err != nil {
		return err
	}
	item, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	context := service.ContextRef.CurrentContext()
	txHash := service.Tx.Hash()
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_LOG, &event.LogEventArgs{TxHash: txHash, ContractAddress: context.ContractAddress, Message: string(item)})

	scv := sitem.Dump()
	log.Debugf("[NeoContract]Debug:%s\n", scv)
	return nil
}

func RuntimeGetTrigger(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushInt64(int64(0))
}

func RuntimeBase58ToAddress(service *NeoVmService, engine *vm.Executor) error {
	item, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	address, err := common.AddressFromBase58(string(item))
	if err != nil {
		return err
	}
	return engine.EvalStack.PushBytes(address[:])
}

func RuntimeAddressToBase58(service *NeoVmService, engine *vm.Executor) error {
	item, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	address, err := common.AddressParseFromBytes(item)
	if err != nil {
		return err
	}
	return engine.EvalStack.PushBytes([]byte(address.ToBase58()))
}

func RuntimeGetCurrentBlockHash(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushBytes(service.BlockHash.ToArray())
}
