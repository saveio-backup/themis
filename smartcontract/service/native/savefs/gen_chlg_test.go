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

package savefs

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/saveio/themis/common"

	"time"
)

func TestGenChallenge(t *testing.T) {
	var hash common.Uint256
	bt := make([]byte, 32)

	for fileBlockNum := 1; fileBlockNum < 12800; fileBlockNum++ {
		rand.Seed(time.Now().Unix())
		rand.Read(bt)
		copy(hash[:], bt)

		challenge := GenChallenge(common.ADDRESS_EMPTY, hash, uint32(fileBlockNum), 32)
		fmt.Printf("==========FileBlockNum: %d, ChallengeLength:%d========\n",
			fileBlockNum, len(challenge))
		fmt.Println("challenge:", challenge)
		for i := 0; i < len(challenge); i++ {
			if challenge[i].Index > uint32(fileBlockNum) {
				fmt.Println("error: ", challenge[i])
				return
			}
		}
	}
}
