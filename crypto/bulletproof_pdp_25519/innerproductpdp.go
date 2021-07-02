package bulletproof_pdp_25519

import (
	"bufio"
	"fmt"
	ristretto255 "github.com/saveio/themis/crypto/ristretto"
	"os"
)

var GHXOFKey = []byte("saveio innerproduct pdp")

type InnerProductPDP struct {
	basepoint *ristretto255.Element
	n         int
	G, H      []*ristretto255.Element
	table     []ristretto255.NafLookupTable8Pro
}

func NewInnerProductPDP() *InnerProductPDP {

	prover := &InnerProductPDP{
		basepoint: new(ristretto255.Element).Base(),
		n:         pdpn,
		G:         nil,
		H:         nil,
	}

	var GConst, HConst []*ristretto255.Element

	for _, gText := range GSerialized {
		g := ElementFromBytes(gText)
		GConst = append(GConst, g)
	}
	for _, hText := range HSerialized {
		h := ElementFromBytes(hText)
		HConst = append(HConst, h)
	}
	prover.G = GConst
	prover.H = HConst

	var GH []*ristretto255.Element
	copyG := deepCopyElement(prover.G)
	copyH := deepCopyElement(prover.H)
	GH = append(GH, copyG...)
	GH = append(GH, copyH...)
	prover.table = ristretto255.GenGHtable(GH)

	return prover
}

//用于生成InnerProductPdp中G、H固定参数
func GenerateConstants() {
	path := "GH_Elements"
	prover := &InnerProductPDP{
		basepoint: new(ristretto255.Element).Base(),
		n:         pdpn,
		G:         nil,
		H:         nil,
	}
	key := GHXOFKey
	prover.G, prover.H = getGH(pdpn, path, key)
	var GH []*ristretto255.Element
	copyG := deepCopyElement(prover.G)
	copyH := deepCopyElement(prover.H)
	GH = append(GH, copyG...)
	GH = append(GH, copyH...)
	prover.table = ristretto255.GenGHtable(GH)
	gconstant, _ := os.Create("gconstant")
	hconstant, _ := os.Create("hconstant")
	defer gconstant.Close()
	defer hconstant.Close()
	for _, g := range prover.G {
		b := ElementToBytes(g)
		fmt.Fprintf(gconstant, ", [32]byte{ ")
		for i, k := range b {
			if i == 31 {
				fmt.Fprintf(gconstant, "%v}", k)
				break
			}
			fmt.Fprintf(gconstant, "%v, ", k)
		}
	}
	for _, h := range prover.H {
		b := ElementToBytes(h)
		fmt.Fprintf(hconstant, ", [32]byte{ ")
		for i, k := range b {
			if i == 31 {
				fmt.Fprintf(hconstant, "%v}", k)
				break
			}
			fmt.Fprintf(hconstant, "%v, ", k)
		}
	}
}

func getGH(n int, filepath string, key []byte) ([]*ristretto255.Element, []*ristretto255.Element) {
	var Gs []*ristretto255.Element
	fi, err := os.Open(filepath)
	if err != nil {
		return generates(n, filepath, key)
	}
	defer fi.Close()
	r := bufio.NewReader(fi)
	buf := make([]byte, 32)
	for i := 0; i < n; i++ {
		_, err := r.Read(buf)
		if err != nil {
			return generates(n, filepath, key)
		}
		var text [32]byte
		copy(text[:], buf)
		Gs = append(Gs, ElementFromBytes(text))
	}
	var Hs []*ristretto255.Element
	for i := 0; i < n; i++ {
		_, err := r.Read(buf)
		if err != nil {
			return generates(n, filepath, key)
		}
		var text [32]byte
		copy(text[:], buf)
		Hs = append(Hs, ElementFromBytes(text))
	}
	return Gs, Hs
}

func generates(n int, filepath string, seed []byte) ([]*ristretto255.Element, []*ristretto255.Element) {
	fi, err := os.Create(filepath)
	if err != nil {
		return nil, nil
	}
	defer fi.Close()
	var G, H []*ristretto255.Element

	xofKey := [32]byte{}
	copy(xofKey[:], seed)
	xof := NewXofExpend(uint16(64*2), xofKey)
	for i := 0; i < n; i++ {
		seedGH := make([]byte, 64*2)
		xof.Read(seedGH)
		G = append(G, new(ristretto255.Element).FromUniformBytes(seedGH[:64]))
		H = append(H, new(ristretto255.Element).FromUniformBytes(seedGH[64:]))
	}

	for _, g := range G {
		b := ElementToBytes(g)
		fi.Write(b[:])
	}
	for _, h := range H {
		b := ElementToBytes(h)
		fi.Write(b[:])
	}
	fi.Sync()

	return G, H
}
