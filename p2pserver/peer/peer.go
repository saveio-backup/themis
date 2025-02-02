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

package peer

import (
	"container/list"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	comm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/p2pserver/common"
	conn "github.com/saveio/themis/p2pserver/link"
	"github.com/saveio/themis/p2pserver/message/types"
)

const (
	TX_HASH = iota
	POC_HASH
)

const (
	MAX_KNOWN_TX  = 1000
	MAX_KNOWN_POC = 1000
)

type knownInfo struct {
	*sync.RWMutex
	queue *list.List
	index map[interface{}]struct{}
}

func newknownInfo() *knownInfo {
	return &knownInfo{
		RWMutex: &sync.RWMutex{},
		queue:   list.New(),
		index:   make(map[interface{}]struct{}),
	}
}

func (info *knownInfo) Len() int {
	info.RLock()
	defer info.RUnlock()

	return info.queue.Len()
}

func (info *knownInfo) Push(hash interface{}) {
	info.Lock()
	defer info.Unlock()

	if _, ok := info.index[hash]; ok {
		return
	}
	info.queue.PushBack(hash)
	info.index[hash] = struct{}{}
}

func (info *knownInfo) Pop() {
	info.Lock()
	defer info.Unlock()

	if info.queue.Len() == 0 {
		return
	}

	e := info.queue.Front()
	info.queue.Remove(e)

	delete(info.index, e.Value)
}

func (info *knownInfo) Has(hash interface{}) bool {
	info.RLock()
	defer info.RUnlock()

	if _, ok := info.index[hash]; ok {
		return true
	}
	return false
}

// PeerInfo provides the basic information of a peer
type PeerInfo struct {
	Id           common.PeerId
	Version      uint32
	Services     uint64
	Relay        bool
	HttpInfoPort uint16
	Port         uint16
	SoftVersion  string
	Addr         string

	height uint64
}

func NewPeerInfo(id common.PeerId, version uint32, services uint64, relay bool, httpInfoPort uint16,
	port uint16, height uint64, softVersion string, addr string) *PeerInfo {
	return &PeerInfo{
		Id:           id,
		Version:      version,
		Services:     services,
		Relay:        relay,
		HttpInfoPort: httpInfoPort,
		Port:         port,
		height:       height,
		SoftVersion:  softVersion,
		Addr:         addr,
	}
}

// RemoteListen get remote service port
func (pi *PeerInfo) RemoteListenAddress() string {
	host, _, err := net.SplitHostPort(pi.Addr)
	if err != nil {
		return ""
	}

	sb := strings.Builder{}
	sb.WriteString(host)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(int(pi.Port)))

	return sb.String()
}

func (self *PeerInfo) Height() uint64 {
	return atomic.LoadUint64(&self.height)
}

func (self *PeerInfo) SetHeight(height uint64) {
	atomic.StoreUint64(&self.height, height)
}

//Peer represent the node in p2p
type Peer struct {
	Info     *PeerInfo
	Link     *conn.Link
	connLock sync.RWMutex
	knownTx  *knownInfo
	knownPoC *knownInfo
}

//NewPeer return new peer without publickey initial
func NewPeer(info *PeerInfo, c net.Conn, msgChan chan *types.MsgPayload) *Peer {
	return &Peer{
		Info: info,
		Link: conn.NewLink(info.Id, c, msgChan),

		knownTx:  newknownInfo(),
		knownPoC: newknownInfo(),
	}
}

func (self *PeerInfo) String() string {
	return fmt.Sprintf("id=%s, version=%s", self.Id.ToHexString(), self.SoftVersion)
}

//DumpInfo print all information of peer
func (this *Peer) DumpInfo() {
	log.Debug("[p2p]Node Info:")
	log.Debug("[p2p]\t id = ", this.GetID())
	log.Debug("[p2p]\t addr = ", this.Info.Addr)
	log.Debug("[p2p]\t version = ", this.GetVersion())
	log.Debug("[p2p]\t services = ", this.GetServices())
	log.Debug("[p2p]\t port = ", this.GetPort())
	log.Debug("[p2p]\t relay = ", this.GetRelay())
	log.Debug("[p2p]\t height = ", this.GetHeight())
	log.Debug("[p2p]\t softVersion = ", this.GetSoftVersion())
}

//GetVersion return peer`s version
func (this *Peer) GetVersion() uint32 {
	return this.Info.Version
}

//GetHeight return peer`s block height
func (this *Peer) GetHeight() uint64 {
	return this.Info.Height()
}

//SetHeight set height to peer
func (this *Peer) SetHeight(height uint64) {
	this.Info.SetHeight(height)
}

//GetPort return Peer`s sync port
func (this *Peer) GetPort() uint16 {
	return this.Info.Port
}

//SendTo call sync link to send buffer
func (this *Peer) SendRaw(msgType string, msgPayload []byte) error {
	return this.Link.SendRaw(msgPayload)
}

//Close halt sync connection
func (this *Peer) Close() {
	this.connLock.Lock()
	this.Link.CloseConn()
	this.connLock.Unlock()
}

//GetID return peer`s id
func (this *Peer) GetID() common.PeerId {
	return this.Info.Id
}

//GetRelay return peer`s relay state
func (this *Peer) GetRelay() bool {
	return this.Info.Relay
}

//GetServices return peer`s service state
func (this *Peer) GetServices() uint64 {
	return this.Info.Services
}

//GetTimeStamp return peer`s latest contact time in ticks
func (this *Peer) GetTimeStamp() int64 {
	return this.Link.GetRXTime()
}

//GetContactTime return peer`s latest contact time in Time struct
func (this *Peer) GetContactTime() time.Time {
	unixNano := this.Link.GetRXTime()
	return time.Unix(unixNano/1e9, unixNano%1e9)
}

//GetAddr return peer`s sync link address
func (this *Peer) GetAddr() string {
	return this.Info.Addr
}

//GetAddr16 return peer`s sync link address in []byte
func (this *Peer) GetAddr16() ([16]byte, error) {
	var result [16]byte
	addrIp, err := common.ParseIPAddr(this.GetAddr())
	if err != nil {
		return result, err
	}
	ip := net.ParseIP(addrIp).To16()
	if ip == nil {
		log.Warn("[p2p]parse ip address error\n", this.GetAddr())
		return result, errors.New("[p2p]parse ip address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

func (this *Peer) GetSoftVersion() string {
	return this.Info.SoftVersion
}

//Send transfer buffer by sync or cons link
func (this *Peer) Send(msg types.Message) error {
	sink := comm.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)

	return this.SendRaw(msg.CmdType(), sink.Bytes())
}

//GetHttpInfoPort return peer`s httpinfo port
func (this *Peer) GetHttpInfoPort() uint16 {
	return this.Info.HttpInfoPort
}

func (this *Peer) MarkTx(hash comm.Uint256) {
	this.knownTx.Push(hash)
	if this.knownTx.Len() > MAX_KNOWN_TX {
		this.knownTx.Pop()
	}
}

func (this *Peer) KnowTx(hash comm.Uint256) bool {
	return this.knownTx.Has(hash)
}

func (this *Peer) MarkPoC(hash comm.Uint256) {
	this.knownPoC.Push(hash)
	if this.knownPoC.Len() > MAX_KNOWN_POC {
		this.knownPoC.Pop()
	}
}

func (this *Peer) KnowPoC(hash comm.Uint256) bool {
	return this.knownPoC.Has(hash)
}
