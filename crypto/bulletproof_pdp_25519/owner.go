package bulletproof_pdp_25519

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	ristretto255 "github.com/saveio/themis/crypto/ristretto"
	"io"
	"os"
)

func (self *InnerProductPDP) GenTag(blocks []Block, fileID [32]byte) [][32]byte {

	Qseed := sha512.Sum512(fileID[:])
	u := new(ristretto255.Element).FromUniformBytes(Qseed[:])

	var b []*ristretto255.Scalar
	xof := NewXofExpend(uint16(64),fileID)
	for i := 0; i < self.n; i++ {
		buf := make([]byte, 64)
		xof.Read(buf)
		b = append(b, new(ristretto255.Scalar).FromUniformBytes(buf))
	}

	var tagsSerialize [][32]byte

	for _, block := range blocks {
		buf := bytes.NewReader(block.Buf)
		text := make([]byte, 64, 64)
		var a []*ristretto255.Scalar
		for i := 0; i < self.n; i++ {
			buf.Read(text)
			a = append(a, new(ristretto255.Scalar).FromUniformBytes(text))
		}
		var scalars []*ristretto255.Scalar
		scalars = append(scalars, a...)
		scalars = append(scalars, b...)
		tag := self.MultiScalarMult_GH(scalars)
		innerproduct := InnerProduct(a, b)
		tag = tag.Add(tag, new(ristretto255.Element).ScalarMult(innerproduct, u))
		tagsSerialize = append(tagsSerialize, ElementToBytes(tag))
	}

	return tagsSerialize
}

func GenerateFileID(path string) ([sha256.Size]byte, error) {
	f, err := os.Open(path)
	if nil != err {
		return [sha256.Size]byte{}, err
	}

	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return [sha256.Size]byte{}, err
	}

	sum := h.Sum(nil)

	var result [sha256.Size]byte
	copy(result[:], sum[:])

	return result, nil
}
