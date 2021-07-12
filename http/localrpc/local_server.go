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

// Package localrpc privides a function to start local rpc server
package localrpc

import (
	"fmt"
	"net/http"
	"strconv"

	cfg "github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/http/base/rpc"
)

const (
	LOCAL_HOST string = "127.0.0.1"
	LOCAL_DIR  string = "/local"
)

func StartLocalServer() error {
	log.Debug()
	http.HandleFunc(LOCAL_DIR, rpc.Handle)

	rpc.HandleFunc("getneighbor", rpc.GetNeighbor)
	rpc.HandleFunc("getnodestate", rpc.GetNodeState)
	rpc.HandleFunc("startconsensus", rpc.StartConsensus)
	rpc.HandleFunc("stopconsensus", rpc.StopConsensus)
	rpc.HandleFunc("setdebuginfo", rpc.SetDebugInfo)
	rpc.HandleFunc("removeplotfile", rpc.RemovePlotFile)
	rpc.HandleFunc("setsipvoteinfo", rpc.SetSipVoteInfo)
	rpc.HandleFunc("setconsvoteinfo", rpc.SetConsVoteInfo)
	rpc.HandleFunc("triggerconselect", rpc.TriggerConsElect)

	// TODO: only listen to local host
	err := http.ListenAndServe(LOCAL_HOST+":"+strconv.Itoa(int(cfg.DefConfig.Rpc.HttpLocalPort)), nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
