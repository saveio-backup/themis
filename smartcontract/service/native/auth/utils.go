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

package auth

import (
	"fmt"
	"sort"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/event"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

var (
	PreAdmin          = []byte{0x01}
	PreRoleFunc       = []byte{0x02}
	PreRoleToken      = []byte{0x03}
	PreDelegateStatus = []byte{0x04}
)

//type(this.contractAddr.Admin) = []byte
func concatContractAdminKey(native *native.NativeService, contractAddr common.Address) []byte {
	this := native.ContextRef.CurrentContext().ContractAddress
	adminKey := append(this[:], contractAddr[:]...)
	adminKey = append(adminKey, PreAdmin...)

	return adminKey
}

func getContractAdmin(native *native.NativeService, contractAddr common.Address) ([]byte, error) {
	key := concatContractAdminKey(native, contractAddr)
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	return item.Value, nil
}

func putContractAdmin(native *native.NativeService, contractAddr common.Address, adminOntID []byte) error {
	key := concatContractAdminKey(native, contractAddr)
	utils.PutBytes(native, key, adminOntID)
	return nil
}

//type(this.contractAddr.RoleFunc.role) = roleFuncs
func concatRoleFuncKey(native *native.NativeService, contractAddr common.Address, role []byte) []byte {
	this := native.ContextRef.CurrentContext().ContractAddress
	roleFuncKey := append(this[:], contractAddr[:]...)
	roleFuncKey = append(roleFuncKey, PreRoleFunc...)
	roleFuncKey = append(roleFuncKey, role...)

	return roleFuncKey
}

func getRoleFunc(native *native.NativeService, contractAddr common.Address, role []byte) (*roleFuncs, error) {
	key := concatRoleFuncKey(native, contractAddr, role)
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	source := common.NewZeroCopySource(item.Value)
	rF := new(roleFuncs)
	err = rF.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize roleFuncs object failed. data: %x", item.Value)
	}
	return rF, nil
}

func putRoleFunc(native *native.NativeService, contractAddr common.Address, role []byte, funcs *roleFuncs) error {
	key := concatRoleFuncKey(native, contractAddr, role)
	utils.PutBytes(native, key, common.SerializeToBytes(funcs))
	return nil
}

//type(this.contractAddr.RoleP.ontID) = roleTokens
func concatOntIDTokenKey(native *native.NativeService, contractAddr common.Address, ontID []byte) []byte {
	this := native.ContextRef.CurrentContext().ContractAddress
	tokenKey := append(this[:], contractAddr[:]...)
	tokenKey = append(tokenKey, PreRoleToken...)
	tokenKey = append(tokenKey, ontID...)

	return tokenKey
}

func getOntIDToken(native *native.NativeService, contractAddr common.Address, ontID []byte) (*roleTokens, error) {
	key := concatOntIDTokenKey(native, contractAddr, ontID)
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	source := common.NewZeroCopySource(item.Value)
	rT := new(roleTokens)
	err = rT.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize roleTokens object failed. data: %x", item.Value)
	}
	return rT, nil
}

func putOntIDToken(native *native.NativeService, contractAddr common.Address, ontID []byte, tokens *roleTokens) error {
	key := concatOntIDTokenKey(native, contractAddr, ontID)
	utils.PutBytes(native, key, common.SerializeToBytes(tokens))
	return nil
}

//type(this.contractAddr.DelegateStatus.ontID)
func concatDelegateStatusKey(native *native.NativeService, contractAddr common.Address, ontID []byte) []byte {
	this := native.ContextRef.CurrentContext().ContractAddress
	key := append(this[:], contractAddr[:]...)
	key = append(key, PreDelegateStatus...)
	key = append(key, ontID...)

	return key
}

func getDelegateStatus(native *native.NativeService, contractAddr common.Address, ontID []byte) (*Status, error) {
	key := concatDelegateStatusKey(native, contractAddr, ontID)
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	status := new(Status)
	source := common.NewZeroCopySource(item.Value)
	err = status.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize Status object failed. data: %x", item.Value)
	}
	return status, nil
}

func putDelegateStatus(native *native.NativeService, contractAddr common.Address, ontID []byte, status *Status) error {
	key := concatDelegateStatusKey(native, contractAddr, ontID)
	utils.PutBytes(native, key, common.SerializeToBytes(status))
	return nil
}

//remove duplicates in the slice of string and sorts the slice in increasing order.
func StringsDedupAndSort(s []string) []string {
	smap := make(map[string]int)
	for i, str := range s {
		if str == "" {
			continue
		}
		smap[str] = i
	}
	ret := make([]string, len(smap))
	i := 0
	for str := range smap {
		ret[i] = str
		i++
	}
	sort.Strings(ret)
	return ret
}

func pushEvent(native *native.NativeService, s interface{}) {
	event := new(event.NotifyEventInfo)
	event.ContractAddress = native.ContextRef.CurrentContext().ContractAddress
	event.States = s
	native.Notifications = append(native.Notifications, event)
}

func serializeAddress(sink *common.ZeroCopySink, addr common.Address) {
	sink.WriteVarBytes(addr[:])
}
