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

package types

import (
	"math"
	"testing"

	"github.com/saveio/themis/core/payload"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_SigHashForChain(t *testing.T) {
	mutable := &MutableTransaction{
		TxType:  InvokeNeo,
		Payload: &payload.InvokeCode{},
	}

	tx, err := mutable.IntoImmutable()
	assert.Nil(t, err)

	assert.Equal(t, tx.Hash(), tx.SigHashForChain(0))
	assert.NotEqual(t, tx.Hash(), tx.SigHashForChain(1))
	assert.NotEqual(t, tx.Hash(), tx.SigHashForChain(math.MaxUint32))
}
