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

package vbft

import (
	"bytes"
	"encoding/hex"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/core/ledger"
	"github.com/saveio/themis/core/states"
	"github.com/saveio/themis/core/utils"
	httpcom "github.com/saveio/themis/http/base/common"
	gov "github.com/saveio/themis/smartcontract/service/native/governance"
	nutils "github.com/saveio/themis/smartcontract/service/native/utils"
)

func GetConsGovView() (*gov.ConsGovView, error) {
	storageKey := &states.StorageKey{
		ContractAddress: nutils.GovernanceContractAddress,
		Key:             append([]byte(gov.CONS_GOV_VIEW)),
	}
	data, err := ledger.DefLedger.GetStorageItem(storageKey.ContractAddress, storageKey.Key)
	if err != nil {
		return nil, err
	}
	consGovView := new(gov.ConsGovView)
	err = consGovView.Deserialize(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return consGovView, nil
}

func GetMiningInfo() (*gov.MiningInfo, error) {
	mutable := utils.BuildNativeTransaction(nutils.GovernanceContractAddress, gov.QUERY_MINING_INFO, []byte{})

	tx, err := mutable.IntoImmutable()
	if err != nil {
		return nil, err
	}

	result, err := ledger.DefLedger.PreExecuteContract(tx)
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(result.Result.(string))
	if err != nil {
		return nil, err
	}

	miningInfo := &gov.MiningInfo{}
	reader := bytes.NewReader(data)
	err = miningInfo.Deserialize(reader)
	if err != nil {
		return nil, err
	}

	return miningInfo, nil

}

//call FS check if plot file is registered for mining
func GetPlotRegInfo(address common.Address, plot string) (bool, error) {
	return true, nil

	//[TODO] call FS when method ready!
	method := "CheckPlotReg"

	mutable, err := httpcom.NewNativeInvokeTransaction(0, 0, nutils.OntFSContractAddress, 0, method, []interface{}{plot})
	if err != nil {
		return false, err
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return false, err
	}

	_, err = ledger.DefLedger.PreExecuteContract(tx)
	if err != nil {
		return false, err
	}

	return true, nil
}
