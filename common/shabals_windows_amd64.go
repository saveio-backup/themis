package common

import (
	"hash"

	"github.com/moonfruit/go-shabal"
)

// Shabal256
type Shabal256 struct {
	hasher hash.Hash
}

func NewShabal256() *Shabal256 {
	result := &Shabal256{}
	result.hasher = shabal.NewShabal256()

	return result
}

func (self *Shabal256) Update(inbuf []byte, off int64, length int64) {
	self.hasher.Write(inbuf[off : off+length])
}

func (self *Shabal256) Digest() []byte {
	return self.hasher.Sum(nil)
}

func (self *Shabal256) Reset() {
	self.hasher.Reset()
}
