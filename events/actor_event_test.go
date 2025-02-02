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

package events

import (
	"fmt"
	"testing"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
)

const testTopic = "test"

type testMessage struct {
	Message string
}

func testSubReceive(c actor.Context) {
	switch msg := c.Message().(type) {
	case *testMessage:
		fmt.Printf("PID:%s receive message:%s\n", c.Self().Id, msg.Message)
	}
}

func TestActorEvent(t *testing.T) {
	Init()
	subPID1 := actor.Spawn(actor.FromFunc(testSubReceive))
	subPID2 := actor.Spawn(actor.FromFunc(testSubReceive))
	sub1 := NewActorSubscriber(subPID1)
	sub2 := NewActorSubscriber(subPID2)
	sub1.Subscribe(testTopic)
	sub2.Subscribe(testTopic)
	DefActorPublisher.Publish(testTopic, &testMessage{Message: "Hello"})
	time.Sleep(time.Millisecond)
	DefActorPublisher.Publish(testTopic, &testMessage{Message: "Word"})
}
