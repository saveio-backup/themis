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
	"bytes"
	"encoding/binary"

	"crypto/sha256"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/savefs/pdp"
)

func GenChallenge(walletAddr common.Address, hashValue common.Uint256, fileBlockNum, proveNum uint32) []pdp.Challenge {
	blockHashArray := hashValue.ToArray()
	plant := append(walletAddr[:], blockHashArray...)
	hash := sha256.Sum256(plant)

	tmpHash := make([]byte, common.UINT256_SIZE+4)
	copy(tmpHash, hash[:])
	copy(tmpHash[common.UINT256_SIZE:], hash[:4])
	var blockNumPerPart, blockNumLastPart, blockNumOfPart uint32

	if fileBlockNum <= 3 {
		blockNumPerPart = 1
		blockNumLastPart = 1
		blockNumOfPart = 1
		proveNum = fileBlockNum
	} else {
		if fileBlockNum > 3 && fileBlockNum < proveNum {
			proveNum = 3
		}
		blockNumPerPart = fileBlockNum / proveNum
		blockNumLastPart = blockNumPerPart + fileBlockNum%proveNum
		blockNumOfPart = blockNumPerPart
	}

	challenge := make([]pdp.Challenge, proveNum)
	blockHash := hash

	var hashIndex = 0
	for i := uint32(1); i <= proveNum; i++ {
		if i == proveNum {
			blockNumOfPart = blockNumLastPart
		}

		rd := BytesToInt(tmpHash[hashIndex : hashIndex+4])
		//challenge[i-1].Index = (rd+1)%blockNumOfPart + (i-1)*blockNumPerPart + 1
		// index start from 0
		challenge[i-1].Index = (rd+1)%blockNumOfPart + (i-1)*blockNumPerPart
		challenge[i-1].Rand = uint32(blockHash[hashIndex]) + 1

		hashIndex++
		hashIndex = hashIndex % common.UINT256_SIZE
	}
	return challenge
}

func BytesToInt(b []byte) uint32 {
	var tmp uint32
	bytesBuffer := bytes.NewBuffer(b)
	binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return tmp
}
