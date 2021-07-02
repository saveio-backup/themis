package bulletproof_pdp_25519

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	ristretto255 "github.com/saveio/themis/crypto/ristretto"
	"io"
)

func (self *InnerProductPDP) SingleVerify(proofText []byte, challenge Challenge, fileID []byte, blockTag [32]byte) (result bool) {

	defer func() {
		if err := recover(); err != nil {
			result = false
		}
	}()

	//G, H := getGH(self.n, "GH_Elements", GHXOFKey)
	G := deepCopyElement(self.G)
	H := deepCopyElement(self.H)

	var proof Proof
	err := proof.Deserialize(proofText)
	if err != nil {
		return false
	}

	//k rounds of prove iteration
	k := len(proof.Ls)
	if k != len(proof.Rs) {
		return false
	}

	p := ElementFromBytes(blockTag)

	trans := sha512.Sum512(Uint32ToBytes(challenge.RandSeed))
	Qseed := sha512.Sum512(fileID)
	u := new(ristretto255.Element).FromUniformBytes(Qseed[:])

	var challenges []*ristretto255.Scalar
	for i := 0; i < k; i++ {
		x := new(ristretto255.Scalar)
		trans, x = updateTranscript(trans, proof.Ls[i], proof.Rs[i])
		challenges = append(challenges, x)
	}

	var challengesSquare []*ristretto255.Scalar
	for _, c := range challenges {
		challengesSquare = append(challengesSquare, Square(c))
	}

	s := GetS(challenges, challengesSquare, self.n, k)

	var as, bsinv []*ristretto255.Scalar
	for i := 0; i < self.n; i++ {
		as = append(as, new(ristretto255.Scalar).Multiply(proof.a, s[i]))
		bsinv = append(bsinv, new(ristretto255.Scalar).Multiply(proof.b, s[self.n-i-1]))
	}

	a := []*ristretto255.Scalar{proof.a}
	b := []*ristretto255.Scalar{proof.b}

	right := SumElements(new(ristretto255.Element).VarTimeMultiScalarMult(as, G),
		new(ristretto255.Element).VarTimeMultiScalarMult(bsinv, H),
		new(ristretto255.Element).ScalarMult(InnerProduct(a, b), u))

	left := p
	for i := 0; i < k; i++ {
		left = SumElements(left,
			new(ristretto255.Element).ScalarMult(challengesSquare[i], proof.Ls[i]),
			new(ristretto255.Element).ScalarMult(new(ristretto255.Scalar).Invert(challengesSquare[i]), proof.Rs[i]))
	}

	return left.Equal(right) == 1
}

//BatchVerify can verify multi proofs. Return false
// if any proof failed, but which one is unknown.
func (self *InnerProductPDP) ProofVerify(pdpversion uint64, proofsText []byte, fileID [][32]byte, blockTags [][32]byte, challenges []Challenge) (result bool) {

	defer func() {
		if err := recover(); err != nil {
			result = false
		}
	}()

	batchsize := len(challenges)

	if len(blockTags) != batchsize {
		return false
	}
	if len(fileID) != batchsize {
		return false
	}

	Ps := make([]*ristretto255.Element, batchsize)

	for i, tag := range blockTags {
		Ps[i] = ElementFromBytes(tag)
	}

	proofs := make([]Proof, batchsize)
	proofBuf := bytes.NewReader(proofsText)
	for i := 0; i < batchsize; i++ {
		pbuf := make([]byte, prooftextsize)
		_, err := proofBuf.Read(pbuf)
		if err != nil && err != io.EOF {
			return false
		}
		err = proofs[i].Deserialize(pbuf)
		if err != nil {
			return false
		}
	}
	//generate u^x with fileID, u^x is the same for different blocks in the same file
	//generate the initial transcript with randseed in challenge
	//transcripts are updated for each proving iteration with xof method
	trans := make([][64]byte, batchsize)
	us := make([]*ristretto255.Element, batchsize)
	for i := 0; i < batchsize; i++ {
		trans[i] = sha512.Sum512(Uint32ToBytes(challenges[i].RandSeed))
		id := fileID[i]
		Qseed := sha512.Sum512(id[:])
		us[i] = new(ristretto255.Element).FromUniformBytes(Qseed[:])
	}
	//check elementright the same length of different proofs
	k := len(proofs[0].Ls)
	for i := 0; i < batchsize; i++ {
		if len(proofs[i].Ls) != k || len(proofs[i].Rs) != k {
			return false
		}
	}
	//generate random scalars to prove g^a==1 & g^b==1 by proving g^(a*r+b) == 1
	var randseed []byte
	for _, cha := range challenges {
		randseed = append(randseed, Uint32ToBytes(cha.RandSeed)...)
	}
	randSeedHash := sha256.Sum256(randseed)
	randList := make([]*ristretto255.Scalar, batchsize)

	xof := NewXofExpend(uint16(64), randSeedHash)
	for i := 0; i < batchsize; i++ {
		seedi := make([]byte, 64)
		xof.Read(seedi)
		randList[i] = new(ristretto255.Scalar).FromUniformBytes(seedi)
	}
	//precompute challenge and its square for later use
	challengesList := MakeDualSlice(batchsize, k)
	for j := 0; j < batchsize; j++ {
		for i := 0; i < k; i++ {
			x := new(ristretto255.Scalar)
			trans[j], x = updateTranscript(trans[j], proofs[j].Ls[i], proofs[j].Rs[i])
			challengesList[j][i] = x
		}
	}

	challengesSquareList := MakeDualSlice(batchsize, k)
	for j := 0; j < batchsize; j++ {
		for i := 0; i < k; i++ {
			challengesSquareList[j][i] = Square(challengesList[j][i])
		}
	}
	sList := MakeDualSlice(batchsize, self.n)
	for j := 0; j < batchsize; j++ {
		sList[j] = GetS(challengesList[j], challengesSquareList[j], self.n, k)
	}
	as := make([]*ristretto255.Scalar, self.n)
	bsinv := make([]*ristretto255.Scalar, self.n)
	for i := 0; i < self.n; i++ {
		asi := Mul(proofs[0].a, sList[0][i], randList[0])
		bsinvi := Mul(proofs[0].b, sList[0][self.n-i-1], randList[0])
		for j := 1; j < batchsize; j++ {
			asi = SumScalars(asi, Mul(proofs[j].a, sList[j][i], randList[j]))
			bsinvi = SumScalars(bsinvi, Mul(proofs[j].b, sList[j][self.n-i-1], randList[j]))
		}
		as[i] = asi
		bsinv[i] = bsinvi
	}

	var scalarright []*ristretto255.Scalar
	var elementright []*ristretto255.Element

	var scalarsGH []*ristretto255.Scalar
	scalarsGH = append(scalarsGH, as...)
	scalarsGH = append(scalarsGH, bsinv...)
	multGH := self.MultiScalarMult_GH(scalarsGH)

	for i := 0; i < batchsize; i++ {
		scalarright = append(scalarright, Mul(proofs[i].a, proofs[i].b, randList[i]))
		elementright = append(elementright, us[i])
	}
	right := new(ristretto255.Element).VarTimeMultiScalarMult(scalarright, elementright)
	right = right.Add(right, multGH)

	left := new(ristretto255.Element).ScalarMult(randList[0], Ps[0])
	for j := 1; j < batchsize; j++ {
		left = SumElements(left, new(ristretto255.Element).ScalarMult(randList[j], Ps[j]))
	}

	for j := 0; j < batchsize; j++ {
		for i := 0; i < k; i++ {
			temp := SumElements(new(ristretto255.Element).ScalarMult(challengesSquareList[j][i], proofs[j].Ls[i]),
				new(ristretto255.Element).ScalarMult(new(ristretto255.Scalar).Invert(challengesSquareList[j][i]), proofs[j].Rs[i]))
			left = SumElements(left, new(ristretto255.Element).ScalarMult(randList[j], temp))
		}
	}

	return left.Equal(right) == 1
}

func GetS(challenges []*ristretto255.Scalar, challengesSquare []*ristretto255.Scalar, n, k int) []*ristretto255.Scalar {
	s := make([]*ristretto255.Scalar, n)
	s[0] = new(ristretto255.Scalar).Invert(Mul(challenges...))

	for i := uint(2); i < uint(n+1); i++ {
		s[i-1] = deepCopyScalar(s[0])
		for j := uint(0); j < uint(k); j++ {
			bit := uint(1) << j
			if bit == (i-1)&bit {
				s[i-1] = new(ristretto255.Scalar).Multiply(s[i-1], challengesSquare[j])
			}
		}
	}

	return s
}
