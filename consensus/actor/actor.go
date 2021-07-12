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

package actor

import (
	"errors"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/themis/core/types"
	ontErrors "github.com/saveio/themis/errors"

	// netActor "github.com/saveio/themis/network/actor/server"
	// ptypes "github.com/saveio/themis/network/component/wire/old/types"
	gov "github.com/saveio/themis/smartcontract/service/native/governance"
	txpool "github.com/saveio/themis/txnpool/common"
)

type TxPoolActor struct {
	Pool *actor.PID
}

func (self *TxPoolActor) GetTxnPool(byCount bool, height uint32) []*txpool.TXEntry {
	poolmsg := &txpool.GetTxnPoolReq{ByCount: byCount, Height: height}
	future := self.Pool.RequestFuture(poolmsg, time.Second*10)
	entry, err := future.Result()
	if err != nil {
		return nil
	}

	txs := entry.(*txpool.GetTxnPoolRsp).TxnPool
	return txs
}

func (self *TxPoolActor) VerifyBlock(txs []*types.Transaction, height uint32) error {
	poolmsg := &txpool.VerifyBlockReq{Txs: txs, Height: height}
	future := self.Pool.RequestFuture(poolmsg, time.Second*10)
	entry, err := future.Result()
	if err != nil {
		return err
	}

	txentry := entry.(*txpool.VerifyBlockRsp).TxnPool
	for _, entry := range txentry {
		if entry.ErrCode != ontErrors.ErrNoError {
			return errors.New(entry.ErrCode.Error())
		}
	}

	return nil
}
func (self *TxPoolActor) GetPoCParam(view uint32) *gov.SubmitNonceParam {
	poolmsg := &txpool.GetPoCReq{View: view}
	future := self.Pool.RequestFuture(poolmsg, time.Second*10)
	entry, err := future.Result()
	if err != nil {
		return nil
	}

	param := entry.(*txpool.GetPoCRsp).Param
	return param
}

type LedgerActor struct {
	Ledger *actor.PID
}

type PoCPoolActor struct {
	Pool *actor.PID
}

func (self *PoCPoolActor) PushPoCParam(param *gov.SubmitNonceParam) {
	pocReq := &txpool.PoCReq{param, txpool.PoCSender}
	self.Pool.Tell(pocReq)
}
