package utils

import (
	"encoding/binary"
	"github.com/saveio/themis/common"
)

const (
	SCOOPS_IN_NONCE     = 4096
	SHABAL256_HASH_SIZE = 32
	SCOOP_SIZE          = SHABAL256_HASH_SIZE * 2
	NONCE_SIZE          = SCOOP_SIZE * SCOOPS_IN_NONCE
)

func CalculateScoop(view uint64, gensig []byte) uint32 {
	data := make([]byte, 8)

	binary.BigEndian.PutUint64(data[:], view)
	data = append(data, gensig[:]...)

	md := common.NewShabal256()
	md.Update(data, 0, int64(len(data)))
	newGenSig := md.Digest()

	scoop := (uint32(newGenSig[30]&0x0F) << 8) | uint32(newGenSig[31])
	return scoop
}
