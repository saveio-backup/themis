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

// Package nodeinfo privides functions for nodeinfo server
package nodeinfo

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/core/ledger"
	p2p "github.com/saveio/themis/p2pserver/net/protocol"
)

type Info struct {
	NodeVersion   string
	BlockHeight   uint32
	NeighborCnt   int
	Neighbors     []NgbNodeInfo
	HttpRestPort  int
	HttpWsPort    int
	HttpJsonPort  int
	HttpLocalPort int
	NodePort      uint16
	NodeId        string
	NodeType      string
}

const (
	VERIFYNODE  = "Verify Node"
	SERVICENODE = "Service Node"
)

var node p2p.P2P

var templates = template.Must(template.New("info").Parse(TEMPLATE_PAGE))

func newNgbNodeInfo(ngbId string, ngbType string, ngbAddr string, httpInfoAddr string, httpInfoPort uint16, ngbVersion string) *NgbNodeInfo {
	return &NgbNodeInfo{NgbId: ngbId, NgbType: ngbType, NgbAddr: ngbAddr, HttpInfoAddr: httpInfoAddr,
		HttpInfoPort: httpInfoPort, NgbVersion: ngbVersion}
}

func initPageInfo(blockHeight uint32, curNodeType string, ngbrCnt int, ngbrsInfo []NgbNodeInfo) (*Info, error) {

	id := fmt.Sprintf("0x%x", node.GetID())
	return &Info{NodeVersion: config.Version, BlockHeight: blockHeight,
		NeighborCnt: ngbrCnt, Neighbors: ngbrsInfo,
		HttpRestPort:  int(config.DefConfig.Restful.HttpRestPort),
		HttpWsPort:    int(config.DefConfig.Ws.HttpWsPort),
		HttpJsonPort:  int(config.DefConfig.Rpc.HttpJsonPort),
		HttpLocalPort: int(config.DefConfig.Rpc.HttpLocalPort),
		NodePort:      uint16(config.DefConfig.P2PNode.NodePort),
		NodeId:        id, NodeType: curNodeType}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	var ngbrNodersInfo []NgbNodeInfo
	var ngbId string
	var ngbAddr string
	var ngbType string
	var ngbInfoPort uint16
	var ngbHttpInfoAddr string
	var ngbVersion string

	curNodeType := SERVICENODE

	ngbrNoders := node.GetNeighbors()
	ngbrsLen := len(ngbrNoders)
	for i := 0; i < ngbrsLen; i++ {
		ngbType = SERVICENODE
		ngbAddr = ngbrNoders[i].GetAddr()
		ngbInfoPort = ngbrNoders[i].GetHttpInfoPort()
		ngbHttpInfoAddr = ngbAddr + ":" + strconv.Itoa(int(ngbInfoPort))
		ngbId = fmt.Sprintf("0x%x", ngbrNoders[i].GetID())
		ngbVersion = fmt.Sprintf("%v", ngbrNoders[i].GetVersion())

		ngbrInfo := newNgbNodeInfo(ngbId, ngbType, ngbAddr, ngbHttpInfoAddr, ngbInfoPort, ngbVersion)
		ngbrNodersInfo = append(ngbrNodersInfo, *ngbrInfo)
	}
	sort.Sort(NgbNodeInfoSlice(ngbrNodersInfo))

	blockHeight := ledger.DefLedger.GetCurrentBlockHeight()
	pageInfo, err := initPageInfo(blockHeight, curNodeType, ngbrsLen, ngbrNodersInfo)
	if err != nil {
		http.Redirect(w, r, "/info", http.StatusFound)
		return
	}

	err = templates.ExecuteTemplate(w, "info", pageInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StartServer(n p2p.P2P) {
	node = n
	port := int(config.DefConfig.P2PNode.HttpInfoPort)
	http.HandleFunc("/info", viewHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
