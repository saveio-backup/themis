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

// Package jsonrpc privides a function to start json rpc server
package jsonrpc

import (
	"fmt"
	"net/http"
	"strconv"

	cfg "github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/http/base/rpc"
)

func StartRPCServer() error {
	log.Debug()
	http.HandleFunc("/", rpc.Handle)

	rpc.HandleFunc("getsysstatusscore", rpc.GetSysStatusScore)
	rpc.HandleFunc("getbestblockhash", rpc.GetBestBlockHash)
	rpc.HandleFunc("getblock", rpc.GetBlock)
	rpc.HandleFunc("getblockcount", rpc.GetBlockCount)
	rpc.HandleFunc("getblockhash", rpc.GetBlockHash)
	rpc.HandleFunc("getconnectioncount", rpc.GetConnectionCount)
	rpc.HandleFunc("getsyncstatus", rpc.GetSyncStatus)
	//HandleFunc("getrawmempool", GetRawMemPool)

	rpc.HandleFunc("getrawtransaction", rpc.GetRawTransaction)
	rpc.HandleFunc("sendrawtransaction", rpc.SendRawTransaction)
	rpc.HandleFunc("getstorage", rpc.GetStorage)
	rpc.HandleFunc("getversion", rpc.GetNodeVersion)
	rpc.HandleFunc("getnetworkid", rpc.GetNetworkId)

	rpc.HandleFunc("getcontractstate", rpc.GetContractState)
	rpc.HandleFunc("getmempooltxcount", rpc.GetMemPoolTxCount)
	rpc.HandleFunc("getmempooltxstate", rpc.GetMemPoolTxState)
	rpc.HandleFunc("getmempooltxhashlist", rpc.GetMemPoolTxHashList)
	rpc.HandleFunc("getsmartcodeevent", rpc.GetSmartCodeEvent)
	rpc.HandleFunc("getsmartcodeeventbyeventid", rpc.GetSmartCodeEventByEventId)
	rpc.HandleFunc("getsmartcodeeventbyeventidandheights", rpc.GetSmartCodeEventByEventIdAndHeights)
	rpc.HandleFunc("getblockheightbytxhash", rpc.GetBlockHeightByTxHash)

	rpc.HandleFunc("getbalance", rpc.GetBalance)
	rpc.HandleFunc("getallowance", rpc.GetAllowance)
	rpc.HandleFunc("getmerkleproof", rpc.GetMerkleProof)
	rpc.HandleFunc("getblocktxsbyheight", rpc.GetBlockTxsByHeight)
	rpc.HandleFunc("getgasprice", rpc.GetGasPrice)

	rpc.HandleFunc("getcrosschainmsg", rpc.GetCrossChainMsg)
	rpc.HandleFunc("getcrossstatesproof", rpc.GetCrossStatesProof)

	err := http.ListenAndServe(":"+strconv.Itoa(int(cfg.DefConfig.Rpc.HttpJsonPort)), nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
