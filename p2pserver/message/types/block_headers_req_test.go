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
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */

package types

import (
	"testing"

	cm "github.com/saveio/themis/common"
)

func TestBlkHdrReqSerializationDeserialization(t *testing.T) {
	var msg HeadersReq
	msg.Len = 1

	hashstr := "8932da73f52b1e22f30c609988ed1f693b6144f74fed9a2a20869afa7abfdf5e"
	msg.HashStart, _ = cm.Uint256FromHexString(hashstr)

	MessageTest(t, &msg)
}
