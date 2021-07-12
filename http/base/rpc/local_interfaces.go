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

package rpc

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saveio/themis/common/log"
	bactor "github.com/saveio/themis/http/base/actor"
	"github.com/saveio/themis/http/base/common"
	berr "github.com/saveio/themis/http/base/error"
)

const (
	RANDBYTELEN = 4
)

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func GetNeighbor(params []interface{}) map[string]interface{} {
	addr := bactor.GetNeighborAddrs()
	return responseSuccess(addr)
}

func GetNodeState(params []interface{}) map[string]interface{} {
	t := time.Now().UnixNano()
	port := bactor.GetNodePort()
	id := bactor.GetID()
	ver := bactor.GetVersion()
	tpe := bactor.GetNodeType()
	relay := bactor.GetRelayState()
	height := bactor.GetCurrentBlockHeight()
	txnCnt, err := bactor.GetTxnCount()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	n := common.NodeInfo{
		NodeTime:    t,
		NodePort:    port,
		ID:          id,
		NodeVersion: ver,
		NodeType:    tpe,
		Relay:       relay,
		Height:      height,
		TxnCnt:      txnCnt,
	}
	return responseSuccess(n)
}

func StartConsensus(params []interface{}) map[string]interface{} {
	if err := bactor.ConsensusSrvStart(); err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	return responsePack(berr.SUCCESS, true)
}

func StopConsensus(params []interface{}) map[string]interface{} {
	if err := bactor.ConsensusSrvHalt(); err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	return responsePack(berr.SUCCESS, true)
}

func SetDebugInfo(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch params[0].(type) {
	case float64:
		level := params[0].(float64)
		if err := log.Log().SetDebugLevel(int(level)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.SUCCESS, true)
}

//remove plot from poc miner
//{"jsonrpc": "2.0", "method": "removeplotfile", "params": ["/plotDir/12345678_1_1024"], "id": 0}
func RemovePlotFile(params []interface{}) map[string]interface{} {
	if len(params) > 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}

	plotfile := ""
	for i := 0; i < len(params); i++ {
		switch params[i].(type) {
		case string:
			plotfile = params[i].(string)
			break
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}

	if err := bactor.RemovePlotFile(plotfile); err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	return responsePack(berr.SUCCESS, true)
}

//set Sip vote decision
//{"jsonrpc": "2.0", "method": "setvoteinfo", "params": ["sipIndex", "agree"], "id": 0}
func SetSipVoteInfo(params []interface{}) map[string]interface{} {
	var sipIndex uint32
	var agree byte

	if len(params) < 2 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch params[0].(type) {
	case float64:
		sipIndex = uint32(params[0].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch params[1].(type) {
	case string:
		str := params[1].(string)

		if strings.ToLower(str) == "agree" {
			agree = 1
		}

	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	_, err := common.GetSipInfo(sipIndex)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if err := bactor.SetSipVoteInfo(sipIndex, agree); err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}

	return responsePack(berr.SUCCESS, true)
}

//set consensus vote decision.
//{"jsonrpc": "2.0", "method": "setvoteinfo", "params": ["node1pubkey", "node1pubkey"], "id": 0}
func SetConsVoteInfo(params []interface{}) map[string]interface{} {
	if len(params) < 1 || len(params) > 3 {
		return responsePack(berr.INVALID_PARAMS, "")
	}

	nodesPubkey := []string{}
	for i := 0; i < len(params); i++ {
		switch params[i].(type) {
		case string:
			pubkey := params[i].(string)
			nodesPubkey = append(nodesPubkey, pubkey)
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}

	if err := bactor.SetConsVoteInfo(nodesPubkey); err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}

	return responsePack(berr.SUCCESS, true)
}

//set consensus vote decision.
//{"jsonrpc": "2.0", "method": "triggerconselect"}
func TriggerConsElect(params []interface{}) map[string]interface{} {

	if err := bactor.TriggerConsElect(); err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}

	return responsePack(berr.SUCCESS, true)
}
