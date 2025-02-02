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
	"fmt"
	"sync"
	"time"

	"github.com/saveio/themis/common/log"
	vconfig "github.com/saveio/themis/consensus/vbft/config"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/p2pserver/common"
)

type Peer struct {
	Index          uint32
	PubKey         keypair.PublicKey
	handShake      *peerHandshakeMsg
	LatestInfo     *peerHeartbeatMsg // latest heartbeat msg
	LastUpdateTime time.Time         // time received heartbeat from peer
	connected      bool
}

type PeerPool struct {
	lock    sync.RWMutex
	maxSize int

	server  *Server
	configs map[uint32]*vconfig.PeerConfig // peer index to peer
	IDMap   map[string]uint32
	P2pMap  map[uint32]common.PeerId //value: p2p random id

	peers                  map[uint32]*Peer
	peerConnectionWaitings map[uint32]chan struct{}
}

func NewPeerPool(maxSize int, server *Server) *PeerPool {
	return &PeerPool{
		maxSize:                maxSize,
		server:                 server,
		configs:                make(map[uint32]*vconfig.PeerConfig),
		IDMap:                  make(map[string]uint32),
		P2pMap:                 make(map[uint32]common.PeerId),
		peers:                  make(map[uint32]*Peer),
		peerConnectionWaitings: make(map[uint32]chan struct{}),
	}
}

func (pool *PeerPool) clean() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.configs = make(map[uint32]*vconfig.PeerConfig)
	pool.IDMap = make(map[string]uint32)
	pool.P2pMap = make(map[uint32]common.PeerId)
	pool.peers = make(map[uint32]*Peer)
}

// FIXME: should rename to isPeerConnected
func (pool *PeerPool) isNewPeer(peerIdx uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if _, present := pool.peers[peerIdx]; present {
		return !pool.peers[peerIdx].connected
	}

	return true
}

func (pool *PeerPool) addPeer(config *vconfig.PeerConfig) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	peerPK, err := vconfig.Pubkey(config.ID)
	if err != nil {
		return fmt.Errorf("failed to unmarshal peer pubkey: %s", err)
	}
	pool.configs[config.Index] = config
	pool.IDMap[config.ID] = config.Index
	pool.peers[config.Index] = &Peer{
		Index:          config.Index,
		PubKey:         peerPK,
		LastUpdateTime: time.Unix(0, 0),
		connected:      false,
	}
	return nil
}

func (pool *PeerPool) getActivePeerCount() int {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	n := 0
	for _, p := range pool.peers {
		if p.connected {
			n++
		}
	}
	return n
}

func (pool *PeerPool) waitPeerConnected(peerIdx uint32) {
	if !pool.isNewPeer(peerIdx) {
		// peer already connected
		return
	}

	var C chan struct{}
	pool.lock.Lock()
	if _, present := pool.peerConnectionWaitings[peerIdx]; !present {
		C = make(chan struct{})
		pool.peerConnectionWaitings[peerIdx] = C
	} else {
		C = pool.peerConnectionWaitings[peerIdx]
	}
	pool.lock.Unlock()

	t1 := time.NewTimer(time.Second * 5)
	select {
	case <-C:
		log.Debugf("waitPeerConnected peer %d connected", peerIdx)
	case <-t1.C:
		log.Debugf("waitPeerConnected peer %d timeout", peerIdx)
	}
}

func (pool *PeerPool) peerConnected(peerIdx uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	// new peer, rather than modify
	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		LastUpdateTime: time.Now(),
		connected:      true,
	}
	if C, present := pool.peerConnectionWaitings[peerIdx]; present {
		delete(pool.peerConnectionWaitings, peerIdx)
		close(C)
	}
}

func (pool *PeerPool) peerDisconnected(peerIdx uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	var lastUpdateTime time.Time
	if p, present := pool.peers[peerIdx]; present {
		lastUpdateTime = p.LastUpdateTime
	}

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		LastUpdateTime: lastUpdateTime,
		connected:      false,
	}
}

func (pool *PeerPool) peerHandshake(peerIdx uint32, msg *peerHandshakeMsg) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		handShake:      msg,
		LatestInfo:     pool.peers[peerIdx].LatestInfo,
		LastUpdateTime: time.Now(),
		connected:      true,
	}
}

func (pool *PeerPool) peerHeartbeat(peerIdx uint32, msg *peerHeartbeatMsg) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if C, present := pool.peerConnectionWaitings[peerIdx]; present {
		// wake up peer connection waitings
		delete(pool.peerConnectionWaitings, peerIdx)
		close(C)
	}

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		handShake:      pool.peers[peerIdx].handShake,
		LatestInfo:     msg,
		LastUpdateTime: time.Now(),
		connected:      true,
	}
}

func (pool *PeerPool) getNeighbours() []*Peer {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	peers := make([]*Peer, 0)
	for _, p := range pool.peers {
		if p.connected {
			peers = append(peers, p)
		}
	}
	return peers
}

func (pool *PeerPool) GetPeerIndex(nodeId string) (uint32, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	idx, present := pool.IDMap[nodeId]
	return idx, present
}

func (pool *PeerPool) GetPeerPubKey(peerIdx uint32) keypair.PublicKey {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if p, present := pool.peers[peerIdx]; present && p != nil {
		return p.PubKey
	}

	return nil
}

func (pool *PeerPool) GetAllPubKeys() map[uint32]keypair.PublicKey {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	keys := make(map[uint32]keypair.PublicKey)
	for idx, peer := range pool.peers {
		keys[idx] = peer.PubKey
	}
	return keys
}

func (pool *PeerPool) isPeerAlive(peerIdx uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	p := pool.peers[peerIdx]
	if p == nil || !p.connected {
		return false
	}

	// p2pserver keeps peer alive

	return true
}

func (pool *PeerPool) getPeer(idx uint32) *Peer {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	peer := pool.peers[idx]
	if peer != nil {
		if peer.PubKey == nil {
			peer.PubKey, _ = vconfig.Pubkey(pool.configs[idx].ID)
		}
		return peer
	}

	return nil
}

func (pool *PeerPool) addP2pId(peerIdx uint32, p2pId common.PeerId) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.P2pMap[peerIdx] = p2pId
}

func (pool *PeerPool) getP2pId(peerIdx uint32) (common.PeerId, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	p2pid, present := pool.P2pMap[peerIdx]
	return p2pid, present
}

func (pool *PeerPool) RemovePeerIndex(nodeId string) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	delete(pool.IDMap, nodeId)
}
