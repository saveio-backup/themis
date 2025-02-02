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

// Package txnpool provides a function to start micro service txPool for
// external process
package txnpool

import (
	"fmt"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology-eventbus/mailbox"
	"github.com/saveio/themis/events"
	"github.com/saveio/themis/events/message"
	tc "github.com/saveio/themis/txnpool/common"
	tp "github.com/saveio/themis/txnpool/proc"
)

// startActor starts an actor with the proxy and unique id,
// and return the pid.
func startActor(obj interface{}, id string) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return obj.(actor.Actor)
	})
	props.WithMailbox(mailbox.BoundedDropping(tc.MAX_LIMITATION))

	pid, _ := actor.SpawnNamed(props, id)
	if pid == nil {
		return nil, fmt.Errorf("fail to start actor at props:%v id:%s",
			props, id)
	}
	return pid, nil
}

// StartTxnPoolServer starts the txnpool server and registers
// actors to handle the msgs from the network, http, consensus
// and validators. Meanwhile subscribes the block complete  event.
func StartTxnPoolServer(disablePreExec, disableBroadcastNetTx bool) (*tp.TXPoolServer, error) {
	var s *tp.TXPoolServer

	/* Start txnpool server to receive msgs from p2p,
	 * consensus and valdiators
	 */
	s = tp.NewTxPoolServer(tc.MAX_WORKER_NUM, disablePreExec, disableBroadcastNetTx)

	// Initialize an actor to handle the msgs from valdiators
	rspActor := tp.NewVerifyRspActor(s)
	rspPid, err := startActor(rspActor, "txVerifyRsp")
	if rspPid == nil {
		return nil, err
	}
	s.RegisterActor(tc.VerifyRspActor, rspPid)

	// Initialize an actor to handle the msgs from consensus
	tpa := tp.NewTxPoolActor(s)
	txPoolPid, err := startActor(tpa, "txPool")
	if txPoolPid == nil {
		return nil, err
	}
	s.RegisterActor(tc.TxPoolActor, txPoolPid)

	// Initialize an actor to handle the msgs from p2p and api
	ta := tp.NewTxActor(s)
	txPid, err := startActor(ta, "tx")
	if txPid == nil {
		return nil, err
	}
	s.RegisterActor(tc.TxActor, txPid)

	// Subscribe the block complete event
	var sub = events.NewActorSubscriber(txPoolPid)
	sub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)
	return s, nil
}
