package poc

import (
	"encoding/binary"
	"math"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/core/utils"
)

func calDeadline(scoop []byte, gensig []byte) (deadline uint64) {
	data := append([]byte{}, gensig[:]...) // gensig 32 bytes
	data = append(data, scoop[:]...)       // scoop 64 bytes

	md := common.NewShabal256()
	md.Update(data, 0, int64(len(data)))
	hash := md.Digest()

	//same with burst calculateHit
	deadline = binary.LittleEndian.Uint64(hash)
	return
}

func findBestDeadline(scoops []byte, gensig []byte) (bestDeadline uint64, offset uint64) {
	numScoops := len(scoops) / utils.SCOOP_SIZE
	bestDeadline = math.MaxUint64
	offset = 0
	for i := 0; i < numScoops; i++ {
		deadline := calDeadline(scoops[i*utils.SCOOP_SIZE:(i+1)*utils.SCOOP_SIZE], gensig)
		if deadline < bestDeadline {
			bestDeadline = deadline
			offset = uint64(i)
		}
	}
	return bestDeadline, offset
}
