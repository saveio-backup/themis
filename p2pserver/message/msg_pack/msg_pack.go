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

package msgpack

import (
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	ct "github.com/saveio/themis/core/types"
	msgCommon "github.com/saveio/themis/p2pserver/common"
	mt "github.com/saveio/themis/p2pserver/message/types"
	gov "github.com/saveio/themis/smartcontract/service/native/governance"
)

//Peer address package
func NewAddrs(nodeAddrs []msgCommon.PeerAddr) mt.Message {
	log.Trace()
	var addr mt.Addr
	addr.NodeAddrs = nodeAddrs

	return &addr
}

//Peer address request package
func NewAddrReq() mt.Message {
	log.Trace()
	var msg mt.AddrReq
	return &msg
}

///block package
func NewBlock(bk *ct.Block, ccMsg *ct.CrossChainMsg, merkleRoot common.Uint256) mt.Message {
	log.Trace()
	var blk mt.Block
	blk.Blk = bk
	blk.MerkleRoot = merkleRoot
	blk.CCMsg = ccMsg

	return &blk
}

//blk hdr package
func NewHeaders(headers []*ct.RawHeader) mt.Message {
	log.Trace()
	var blkHdr mt.RawBlockHeader
	blkHdr.BlkHdr = headers

	return &blkHdr
}

//blk hdr req package
func NewHeadersReq(curHdrHash common.Uint256) mt.Message {
	log.Trace()
	var h mt.HeadersReq
	h.Len = 1
	h.HashEnd = curHdrHash

	return &h
}

////Consensus info package
func NewConsensus(cp *mt.ConsensusPayload) mt.Message {
	log.Trace()
	var cons mt.Consensus
	cons.Cons = *cp

	return &cons
}

//InvPayload
func NewInvPayload(invType common.InventoryType, msg []common.Uint256) *mt.InvPayload {
	log.Trace()
	return &mt.InvPayload{
		InvType: invType,
		Blk:     msg,
	}
}

//Inv request package
func NewInv(invPayload *mt.InvPayload) mt.Message {
	log.Trace()
	var inv mt.Inv
	inv.P.Blk = invPayload.Blk
	inv.P.InvType = invPayload.InvType

	return &inv
}

//NotFound package
func NewNotFound(hash common.Uint256) mt.Message {
	log.Trace()
	var notFound mt.NotFound
	notFound.Hash = hash

	return &notFound
}

//ping msg package
func NewPingMsg(height uint64) *mt.Ping {
	log.Trace()
	var ping mt.Ping
	ping.Height = uint64(height)

	return &ping
}

//pong msg package
func NewPongMsg(height uint64) *mt.Pong {
	log.Trace()
	var pong mt.Pong
	pong.Height = uint64(height)

	return &pong
}

//Transaction package
func NewTxn(txn *ct.Transaction) mt.Message {
	log.Trace()
	var trn mt.Trn
	trn.Txn = txn

	return &trn
}

//transaction request package
func NewTxnDataReq(hash common.Uint256) mt.Message {
	log.Trace()
	var dataReq mt.DataReq
	dataReq.DataType = common.TRANSACTION
	dataReq.Hash = hash

	return &dataReq
}

//block request package
func NewBlkDataReq(hash common.Uint256) mt.Message {
	log.Trace()
	var dataReq mt.DataReq
	dataReq.DataType = common.BLOCK
	dataReq.Hash = hash

	return &dataReq
}

//consensus request package
func NewConsensusDataReq(hash common.Uint256) mt.Message {
	log.Trace()
	var dataReq mt.DataReq
	dataReq.DataType = common.CONSENSUS
	dataReq.Hash = hash

	return &dataReq
}

func NewFindNodeReq(id msgCommon.PeerId) mt.Message {
	req := mt.FindNodeReq{
		TargetID: id,
	}

	return &req
}

//Peer address request package
func NewSubmitNonce(param *gov.SubmitNonceParam) mt.Message {
	log.Trace()
	var msg mt.SubmitNonceParam
	msg.View = param.View
	msg.Address = param.Address[:]
	msg.Id = param.Id
	msg.Nonce = param.Nonce
	msg.Deadline = param.Deadline
	msg.PlotName = param.PlotName
	msg.VoteConsPub = param.VoteConsPub
	msg.VoteId = param.VoteId
	msg.VoteInfo = param.VoteInfo
	msg.MoveUpElect = param.MoveUpElect

	return &msg
}
