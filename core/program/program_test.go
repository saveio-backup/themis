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

package program

import (
	"testing"

	"github.com/saveio/themis/crypto/keypair"
	"github.com/stretchr/testify/assert"
)

func TestProgramBuilder_PushBytes(t *testing.T) {
	N := 20000
	builder := NewProgramBuilder()
	for i := 0; i < N; i++ {
		builder.PushNum(uint16(i))
	}
	parser := newProgramParser(builder.Finish())
	for i := 0; i < N; i++ {
		n, err := parser.ReadNum()
		assert.Nil(t, err)
		assert.Equal(t, n, uint16(i))
	}
}

func TestGetProgramInfo(t *testing.T) {
	N := 10
	M := N / 2
	var pubkeys []keypair.PublicKey
	for i := 0; i < N; i++ {
		_, key, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
		pubkeys = append(pubkeys, key)
	}

	pubkeys = keypair.SortPublicKeys(pubkeys)
	progInfo := ProgramInfo{PubKeys: pubkeys, M: uint16(M)}
	prog, err := ProgramFromMultiPubKey(progInfo.PubKeys, int(progInfo.M))
	assert.Nil(t, err)

	info2, err := GetProgramInfo(prog)
	assert.Nil(t, err)
	assert.Equal(t, progInfo, info2)
}
