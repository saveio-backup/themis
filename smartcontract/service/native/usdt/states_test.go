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

package usdt

import (
	"bytes"
	"testing"

	"github.com/saveio/themis/common"
	"github.com/stretchr/testify/assert"
)

func TestState_Serialize(t *testing.T) {
	state := State{
		From:  common.AddressFromVmCode([]byte{1, 2, 3}),
		To:    common.AddressFromVmCode([]byte{4, 5, 6}),
		Value: 1,
	}
	bf := new(bytes.Buffer)
	if err := state.Serialize(bf); err != nil {
		t.Fatal("state serialize fail!")
	}

	state2 := State{}
	if err := state2.Deserialize(bf); err != nil {
		t.Fatal("state deserialize fail!")
	}

	assert.Equal(t, state, state2)
}
