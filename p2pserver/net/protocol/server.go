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

// Package p2p provides an network interface
package p2p

import (
	"github.com/saveio/themis/p2pserver/common"
	"github.com/saveio/themis/p2pserver/message/types"
	"github.com/saveio/themis/p2pserver/peer"
)

//P2P represent the net interface of p2p package
type P2P interface {
	Connect(addr string)
	GetHostInfo() *peer.PeerInfo
	GetID() common.PeerId
	GetNeighbors() []*peer.Peer
	GetNeighborAddrs() []common.PeerAddr
	GetConnectionCnt() uint32
	GetMaxPeerBlockHeight() uint64
	GetPeer(id common.PeerId) *peer.Peer
	SetHeight(uint64)
	Send(p *peer.Peer, msg types.Message) error
	SendTo(p common.PeerId, msg types.Message)
	GetOutConnRecordLen() uint
	Broadcast(msg types.Message)
	IsOwnAddress(addr string) bool
}
