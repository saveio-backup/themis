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
	"encoding/binary"
	"github.com/saveio/themis/common"
)

const (
	HASH_SIZE        = 32
	HASHES_PER_SCOOP = 2
	SCOOP_SIZE       = HASHES_PER_SCOOP * HASH_SIZE
	SCOOPS_PER_PLOT  = 4096 // original 1MB/plot = 16384
	PLOT_SIZE        = SCOOPS_PER_PLOT * SCOOP_SIZE
	HASH_CAP         = 4096
)

type MiningPlot struct {
	data []byte
}

//func NewMiningPlot(addr int64, nonce int64) *MiningPlot {
func NewMiningPlot(addr int64, nonce uint64) *MiningPlot {
	self := &MiningPlot{}
	self.data = make([]byte, PLOT_SIZE)

	buf := make([]byte, 16)
	//use big endian to be same with plot program
	binary.BigEndian.PutUint64(buf[:], uint64(addr))
	binary.BigEndian.PutUint64(buf[8:], uint64(nonce))

	gendata := make([]byte, PLOT_SIZE+len(buf))
	gendata = append(gendata[:PLOT_SIZE], buf...)

	md := common.NewShabal256()
	length := len(buf)

	var i int64
	for i = PLOT_SIZE; i > 0; i -= HASH_SIZE {
		md.Reset()
		lens := int64(PLOT_SIZE+length) - i
		if lens > HASH_CAP {
			lens = HASH_CAP
		}
		md.Update(gendata, i, lens)
		tmpHash := md.Digest()
		arraycopy(tmpHash, 0, gendata, i-HASH_SIZE, HASH_SIZE)
	}
	md.Reset()
	md.Update(gendata, 0, int64(len(gendata)))
	finalhash := md.Digest()
	for i = 0; i < PLOT_SIZE; i++ {
		self.data[i] = (byte)(gendata[i] ^ finalhash[i%HASH_SIZE])
	}

	//PoC2 Rearrangement
	var pos, revPos int64
	hashBuffer := make([]byte, HASH_SIZE)
	revPos = PLOT_SIZE - HASH_SIZE                   //Start at second hash in last scoop
	for pos = 32; pos < (PLOT_SIZE / 2); pos += 64 { //Start at second hash in first scoop

		arraycopy(self.data, pos, hashBuffer, 0, HASH_SIZE)     //Copy low scoop second hash to buffer
		arraycopy(self.data, revPos, self.data, pos, HASH_SIZE) //Copy high scoop second hash to low scoop second hash
		arraycopy(hashBuffer, 0, self.data, revPos, HASH_SIZE)

		revPos -= 64 //move backwards
	}

	return self
}

func (self *MiningPlot) GetScoopData(scoop int) []byte {
	return self.data[scoop*SCOOP_SIZE : scoop*SCOOP_SIZE+SCOOP_SIZE]
}

func arraycopy(src []byte, from int64, dst []byte, to int64, count int64) {
	var i int64

	for i = 0; i < count; i++ {
		dst[to+i] = src[from+i]
	}
}
