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
package common

import (
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/core/payload"
	"github.com/saveio/themis/core/types"
)

type PayloadInfo interface{}

type InvokeCodeInfo struct {
	Code string
}
type DeployCodeInfo struct {
	Code        string
	VmType      byte
	Name        string
	CodeVersion string
	Author      string
	Email       string
	Description string
}

type BookkeeperInfo struct {
	PubKey     string
	Action     string
	Issuer     string
	Controller string
}

//get tranasction payload data
func TransPayloadToHex(p types.Payload) PayloadInfo {
	switch object := p.(type) {
	case *payload.Bookkeeper:
		obj := new(BookkeeperInfo)
		pubKeyBytes := keypair.SerializePublicKey(object.PubKey)
		obj.PubKey = common.ToHexString(pubKeyBytes)
		if object.Action == payload.BookkeeperAction_ADD {
			obj.Action = "add"
		} else if object.Action == payload.BookkeeperAction_SUB {
			obj.Action = "sub"
		} else {
			obj.Action = "nil"
		}
		pubKeyBytes = keypair.SerializePublicKey(object.Issuer)
		obj.Issuer = common.ToHexString(pubKeyBytes)

		return obj
	case *payload.InvokeCode:
		obj := new(InvokeCodeInfo)
		obj.Code = common.ToHexString(object.Code)
		return obj
	case *payload.DeployCode:
		obj := new(DeployCodeInfo)
		obj.Code = common.ToHexString(object.GetRawCode())
		obj.VmType = byte(object.VmType())
		obj.Name = object.Name
		obj.CodeVersion = object.Version
		obj.Author = object.Author
		obj.Email = object.Email
		obj.Description = object.Description
		return obj
	}
	return nil
}
