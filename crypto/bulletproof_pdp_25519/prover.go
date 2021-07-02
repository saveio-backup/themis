package bulletproof_pdp_25519

import (
	"bytes"
	"crypto/sha512"
	"github.com/saveio/themis/crypto/ristretto"
	"github.com/saveio/themis/errors"
)

func (self *InnerProductPDP) ProofGenerate(pdpversion uint64, blocks []Block, fileID [][32]byte, challenges []Challenge) (result []byte, err error) {

	defer func() {
		if e := recover(); e != nil {
			result = nil
			err = errors.NewErr("[PDP ProofGeneratge] panic")
		}
	}()

	batchsize := len(challenges)

	if len(blocks) != batchsize {
		return nil, errors.NewErr("[PDP ProofGenerate] blocks and challenges length not the same")
	}

	if len(fileID) != batchsize {
		return nil, errors.NewErr("[PDP ProofGenerate] fileID and challenges length not the same")
	}

	var proofText []byte

	for i := 0; i < batchsize; i++ {
		id := fileID[i]
		proof := self.SingleProofGenerate(blocks[i], id[:], challenges[i])
		proofText = append(proofText, proof.Serialize()...)
	}
	return proofText, nil
}

func (self *InnerProductPDP) SingleProofGenerate(block Block, fileID []byte, challenge Challenge) Proof {

	G := deepCopyElement(self.G)
	H := deepCopyElement(self.H)

	var a, b []*ristretto255.Scalar

	xofKey := [32]byte{}
	copy(xofKey[:], fileID)
	xof := NewXofExpend(uint16(64), xofKey)
	for i := 0; i < self.n; i++ {
		buf := make([]byte, 64)
		xof.Read(buf)
		b = append(b, new(ristretto255.Scalar).FromUniformBytes(buf))
	}

	r := bytes.NewReader(block.Buf)
	for i := 0; i < self.n; i++ {
		buf := make([]byte, 64)
		n,_ := r.Read(buf)
		if n!=64 {
			panic("block size error")
		}
		a = append(a, new(ristretto255.Scalar).FromUniformBytes(buf))
	}

	trans := sha512.Sum512(Uint32ToBytes(challenge.RandSeed))
	Qseed := sha512.Sum512(fileID)
	u := new(ristretto255.Element).FromUniformBytes(Qseed[:])

	P := self.SumMultElements(a, b, G, H, u)

	round := self.n
	var Ls, Rs []*ristretto255.Element
	var x *ristretto255.Scalar

	for round != 1 {

		round = round / 2

		Li := self.SumMultElements(left(a), right(b), rightElements(G), leftElements(H), u)
		Ri := self.SumMultElements(right(a), left(b), leftElements(G), rightElements(H), u)

		Ls = append(Ls, Li)
		Rs = append(Rs, Ri)

		trans, x = updateTranscript(trans, Li, Ri)
		xInv := new(ristretto255.Scalar).Invert(x)

		G = HadamardElements(
			ScalarMultArray(xInv, leftElements(G)),
			ScalarMultArray(x, rightElements(G)))

		H = HadamardElements(
			ScalarMultArray(x, leftElements(H)),
			ScalarMultArray(xInv, rightElements(H)))

		Lx := new(ristretto255.Element).ScalarMult(Square(x), Li)
		Rx := new(ristretto255.Element).ScalarMult(Square(xInv), Ri)

		P = SumElements(Lx, P, Rx)

		a = HadamardScalars(scalarMul(left(a), x), scalarMul(right(a), xInv))
		b = HadamardScalars(scalarMul(left(b), xInv), scalarMul(right(b), x))

	}

	return Proof{
		iteration: int32(len(Ls)),
		Ls:        Ls,
		Rs:        Rs,
		a:         a[0],
		b:         b[0],
	}

}

func updateTranscript(trans [64]byte, lpt, rpt *ristretto255.Element) ([64]byte, *ristretto255.Scalar) {
	lx, _ := lpt.MarshalText()
	rx, _ := rpt.MarshalText()
	var h []byte
	h = append(h, trans[:]...)
	h = append(h, lx...)
	h = append(h, rx...)
	trans = sha512.Sum512(h)
	return trans, new(ristretto255.Scalar).FromUniformBytes(trans[:])
}
