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

package dbft

import (
	"fmt"

	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/ledger"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/core/vote"
	msg "github.com/saveio/themis/p2pserver/message/types"
)

const ContextVersion uint32 = 0

type ConsensusContext struct {
	State           ConsensusState
	PrevHash        common.Uint256
	Height          uint32
	ViewNumber      byte
	Bookkeepers     []keypair.PublicKey
	NextBookkeepers []keypair.PublicKey
	Owner           keypair.PublicKey
	BookkeeperIndex int
	PrimaryIndex    uint32
	Timestamp       uint32
	Nonce           uint64
	NextBookkeeper  common.Address
	Transactions    []*types.Transaction
	Signatures      [][]byte
	ExpectedView    []byte

	header *types.Block

	isBookkeeperChanged bool
	nmChangedblkHeight  uint32
}

func (ctx *ConsensusContext) M() int {
	log.Debug()
	return len(ctx.Bookkeepers) - (len(ctx.Bookkeepers)-1)/3
}

func NewConsensusContext() *ConsensusContext {
	log.Debug()
	return &ConsensusContext{}
}

func (ctx *ConsensusContext) ChangeView(viewNum byte) {
	log.Debug()
	p := (ctx.Height - uint32(viewNum)) % uint32(len(ctx.Bookkeepers))
	ctx.State &= SignatureSent
	ctx.ViewNumber = viewNum
	if p >= 0 {
		ctx.PrimaryIndex = uint32(p)
	} else {
		ctx.PrimaryIndex = uint32(p) + uint32(len(ctx.Bookkeepers))
	}

	if ctx.State == Initial {
		ctx.Transactions = nil
		ctx.Signatures = make([][]byte, len(ctx.Bookkeepers))
		ctx.header = nil
	}
}

func (ctx *ConsensusContext) MakeChangeView() *msg.ConsensusPayload {
	log.Debug()
	cv := &ChangeView{
		NewViewNumber: ctx.ExpectedView[ctx.BookkeeperIndex],
	}
	cv.msgData.Type = ChangeViewMsg
	return ctx.MakePayload(cv)
}

func (ctx *ConsensusContext) MakeHeader() *types.Block {
	if ctx.header == nil {
		txHash := make([]common.Uint256, 0, len(ctx.Transactions))
		for _, t := range ctx.Transactions {
			txHash = append(txHash, t.Hash())
		}
		txRoot := common.ComputeMerkleRoot(txHash)
		blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoots(ctx.Height, []common.Uint256{txRoot})
		header := &types.Header{
			Version:          ContextVersion,
			PrevBlockHash:    ctx.PrevHash,
			TransactionsRoot: txRoot,
			BlockRoot:        blockRoot,
			Timestamp:        ctx.Timestamp,
			Height:           ctx.Height,
			ConsensusData:    ctx.Nonce,
			NextBookkeeper:   ctx.NextBookkeeper,
		}
		ctx.header = &types.Block{
			Header:       header,
			Transactions: []*types.Transaction{},
		}
	}
	return ctx.header
}

func (ctx *ConsensusContext) MakePayload(message ConsensusMessage) *msg.ConsensusPayload {
	log.Debug()
	message.ConsensusMessageData().ViewNumber = ctx.ViewNumber
	sink := common.NewZeroCopySink(nil)
	message.Serialization(sink)
	return &msg.ConsensusPayload{
		Version:         ContextVersion,
		PrevHash:        ctx.PrevHash,
		Height:          ctx.Height,
		BookkeeperIndex: uint16(ctx.BookkeeperIndex),
		Timestamp:       ctx.Timestamp,
		Data:            sink.Bytes(),
		Owner:           ctx.Owner,
	}
}

func (ctx *ConsensusContext) MakePrepareRequest() *msg.ConsensusPayload {
	log.Debug()
	preReq := &PrepareRequest{
		Nonce:          ctx.Nonce,
		NextBookkeeper: ctx.NextBookkeeper,
		Transactions:   ctx.Transactions,
		Signature:      ctx.Signatures[ctx.BookkeeperIndex],
	}
	preReq.msgData.Type = PrepareRequestMsg
	return ctx.MakePayload(preReq)
}

func (ctx *ConsensusContext) MakePrepareResponse(signature []byte) *msg.ConsensusPayload {
	log.Debug()
	preRes := &PrepareResponse{
		Signature: signature,
	}
	preRes.msgData.Type = PrepareResponseMsg
	return ctx.MakePayload(preRes)
}

func (ctx *ConsensusContext) MakeBlockSignatures(signatures []SignaturesData) *msg.ConsensusPayload {
	log.Debug()
	sigs := &BlockSignatures{
		Signatures: signatures,
	}
	sigs.msgData.Type = BlockSignaturesMsg
	return ctx.MakePayload(sigs)
}

func (ctx *ConsensusContext) GetSignaturesCount() (count int) {
	log.Debug()
	count = 0
	for _, sig := range ctx.Signatures {
		if sig != nil {
			count += 1
		}
	}
	return count
}

func (ctx *ConsensusContext) GetStateDetail() string {

	return fmt.Sprintf("Initial: %t, Primary: %t, Backup: %t, RequestSent: %t, RequestReceived: %t, SignatureSent: %t, BlockGenerated: %t, ",
		ctx.State.HasFlag(Initial),
		ctx.State.HasFlag(Primary),
		ctx.State.HasFlag(Backup),
		ctx.State.HasFlag(RequestSent),
		ctx.State.HasFlag(RequestReceived),
		ctx.State.HasFlag(SignatureSent),
		ctx.State.HasFlag(BlockGenerated))

}

func (ctx *ConsensusContext) Reset(bkAccount *account.Account) {
	preHash := ledger.DefLedger.GetCurrentBlockHash()
	height := ledger.DefLedger.GetCurrentBlockHeight()
	header := ctx.MakeHeader()

	if height != ctx.Height || header == nil || header.Hash() != preHash || len(ctx.NextBookkeepers) == 0 {
		log.Info("[ConsensusContext] Calculate Bookkeepers from db")
		var err error
		ctx.Bookkeepers, err = vote.GetValidators([]*types.Transaction{})
		if err != nil {
			log.Error("[ConsensusContext] GetNextBookkeeper failed", err)
		}
	} else {
		ctx.Bookkeepers = ctx.NextBookkeepers
	}

	ctx.State = Initial
	ctx.PrevHash = preHash
	ctx.Height = height + 1
	ctx.ViewNumber = 0
	ctx.BookkeeperIndex = -1
	ctx.NextBookkeepers = nil
	bookkeeperLen := len(ctx.Bookkeepers)
	ctx.PrimaryIndex = ctx.Height % uint32(bookkeeperLen)
	ctx.Transactions = nil
	ctx.header = nil
	ctx.Signatures = make([][]byte, bookkeeperLen)
	ctx.ExpectedView = make([]byte, bookkeeperLen)

	log.Debugf("bookkeepers number: %d", bookkeeperLen)
	for i := 0; i < bookkeeperLen; i++ {
		if keypair.ComparePublicKey(bkAccount.PublicKey, ctx.Bookkeepers[i]) {
			log.Debugf("this node is bookkeeper %d", i)
			ctx.BookkeeperIndex = i
			ctx.Owner = ctx.Bookkeepers[i]
			break
		}
	}

}
