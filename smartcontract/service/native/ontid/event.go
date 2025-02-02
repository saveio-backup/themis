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
package ontid

import (
	"encoding/hex"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/event"
	"github.com/saveio/themis/smartcontract/service/native"
)

func newEvent(srvc *native.NativeService, st interface{}) {
	e := event.NotifyEventInfo{}
	e.ContractAddress = srvc.ContextRef.CurrentContext().ContractAddress
	e.States = st
	srvc.Notifications = append(srvc.Notifications, &e)
}

func triggerRegisterEvent(srvc *native.NativeService, id []byte) {
	newEvent(srvc, []string{"Register", string(id)})
}

func triggerPublicEvent(srvc *native.NativeService, op string, id, pub []byte, keyID uint32) {
	st := []interface{}{"PublicKey", op, string(id), keyID, hex.EncodeToString(pub)}
	newEvent(srvc, st)
}

func triggerAttributeEvent(srvc *native.NativeService, op string, id []byte, path [][]byte) {
	var attr interface{}
	if op == "remove" {
		attr = hex.EncodeToString(path[0])
	} else {
		t := make([]string, len(path))
		for i, v := range path {
			t[i] = hex.EncodeToString(v)
		}
		attr = t
	}
	st := []interface{}{"Attribute", op, string(id), attr}
	newEvent(srvc, st)
}

func triggerRecoveryEvent(srvc *native.NativeService, op string, id []byte, addr common.Address) {
	st := []string{"Recovery", op, string(id), addr.ToHexString()}
	newEvent(srvc, st)
}

func triggerContextEvent(srvc *native.NativeService, op string, id []byte, contexts [][]byte) {
	t := make([]string, len(contexts))
	var c interface{}
	for i := 0; i < len(contexts); i++ {
		t[i] = hex.EncodeToString(contexts[i])
	}
	c = t
	st := []interface{}{"Context", op, string(id), c}
	newEvent(srvc, st)
}

func triggerServiceEvent(srvc *native.NativeService, op string, id []byte, serviceId []byte) {
	st := []string{"Service", op, string(id), common.ToHexString(serviceId)}
	newEvent(srvc, st)
}

func triggerAuthKeyEvent(srvc *native.NativeService, op string, id []byte, keyID uint32) {
	st := []interface{}{"AuthKey", op, string(id), keyID}
	newEvent(srvc, st)
}
