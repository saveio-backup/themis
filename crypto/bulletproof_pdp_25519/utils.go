package bulletproof_pdp_25519

import (
	"github.com/saveio/themis/crypto/ristretto"
	"golang.org/x/crypto/blake2s"
	"io"
)


type XofExpend struct {
	Xof  blake2s.XOF
	key  [32]byte
	size uint16
}

func NewXofExpend(size uint16,key [32]byte) XofExpend {
	xofExpend := XofExpend{
		key:  key,
		size: size,
	}
	xofExpend.Expend()
	return xofExpend
}

func (self *XofExpend) Read(p []byte) (int, error) {
	l := len(p)
	if len(p) > int(self.size) {
		return 0, io.ErrShortBuffer
	}
	if self.Emty() || self.Xof == nil {
		self.Expend()
	}
	updateKey := make([]byte, 32)
	k, _ := self.Xof.Read(updateKey)
	if k != 32 {
		return 0, io.ErrUnexpectedEOF
	}
	copy(self.key[:], updateKey)
	if self.Emty() || self.Xof == nil {
		self.Expend()
	}
	n, _ := self.Xof.Read(p)
	if n != l {
		return 0, io.ErrUnexpectedEOF
	}
	self.Expend()
	return n, nil
}

func (self *XofExpend) Emty() bool {
	check := make([]byte, 0)
	_, err := self.Xof.Read(check)
	return err == io.EOF
}

func (self *XofExpend) Expend() {
	self.Xof, _ = blake2s.NewXOF(self.size+32, self.key[:])
}

func InnerProduct(a, b []*ristretto255.Scalar) *ristretto255.Scalar {
	product := new(ristretto255.Scalar).Zero()
	for i := 0; i < len(a); i++ {
		product.Add(product, new(ristretto255.Scalar).Multiply(a[i], b[i]))
	}
	return product
}

func left(s []*ristretto255.Scalar) []*ristretto255.Scalar {
	var r []*ristretto255.Scalar
	for i := range s {
		if i&1 == 0 { // even
			r = append(r, s[i])
		}
	}
	return r
}

func right(s []*ristretto255.Scalar) []*ristretto255.Scalar {
	var r []*ristretto255.Scalar
	for i := range s {
		if i&1 == 1 { // odd
			r = append(r, s[i])
		}
	}
	return r
}

func leftElements(s []*ristretto255.Element) []*ristretto255.Element {
	var r []*ristretto255.Element
	for i := range s {
		if i&1 == 0 { // even
			r = append(r, s[i])
		}
	}
	return r
}

func rightElements(s []*ristretto255.Element) []*ristretto255.Element {
	var r []*ristretto255.Element
	for i := range s {
		if i&1 == 1 { // odd
			r = append(r, s[i])
		}
	}
	return r
}

func (self *InnerProductPDP)SumMultElements(a, b []*ristretto255.Scalar, G, H []*ristretto255.Element, u *ristretto255.Element) *ristretto255.Element {

	var scalars []*ristretto255.Scalar
	var elements []*ristretto255.Element
	product := InnerProduct(a, b)

	scalars = append(scalars, a...)
	scalars = append(scalars, b...)
	scalars = append(scalars, product)

	elements = append(elements, G...)
	elements = append(elements, H...)
	elements = append(elements, u)
	result := new(ristretto255.Element).VarTimeMultiScalarMult(scalars, elements)
	return result

}

func HadamardElements(a, b []*ristretto255.Element) []*ristretto255.Element {
	result := make([]*ristretto255.Element, len(a))
	for i := range a {
		result[i] = new(ristretto255.Element).Add(a[i], b[i])
	}
	return result
}

func ScalarMultArray(scalar *ristretto255.Scalar, elements []*ristretto255.Element) (result []*ristretto255.Element) {
	for _, e := range elements {
		result = append(result, new(ristretto255.Element).ScalarMult(scalar, e))
	}
	return result
}

func Square(s *ristretto255.Scalar) *ristretto255.Scalar {
	return new(ristretto255.Scalar).Multiply(s, s)
}

func SumElements(elements ...*ristretto255.Element) *ristretto255.Element {
	if len(elements) < 2 {
		return elements[0]
	}
	result := elements[0]
	for i := 1; i < len(elements); i++ {
		result = new(ristretto255.Element).Add(result, elements[i])
	}
	return result
}

func HadamardScalars(a, b []*ristretto255.Scalar) []*ristretto255.Scalar {
	result := make([]*ristretto255.Scalar, len(a))
	for i := range a {
		result[i] = new(ristretto255.Scalar).Add(a[i], b[i])
	}
	return result
}

func scalarMul(a []*ristretto255.Scalar, b *ristretto255.Scalar) (result []*ristretto255.Scalar) {
	for _, s := range a {
		result = append(result, new(ristretto255.Scalar).Multiply(s, b))
	}
	return result
}

func SumScalars(s ...*ristretto255.Scalar) *ristretto255.Scalar {

	sum := new(ristretto255.Scalar).Zero()
	for _, si := range s {
		sum.Add(sum, si)
	}
	return sum
}

func Mul(scalars ...*ristretto255.Scalar) *ristretto255.Scalar {
	if len(scalars) == 1 {
		return scalars[0]
	}
	result := scalars[0]
	for i := 1; i < len(scalars); i++ {
		result = new(ristretto255.Scalar).Multiply(result, scalars[i])
	}
	return result
}

func MakeDualSlice(row, column int) [][]*ristretto255.Scalar {
	var dualSlice [][]*ristretto255.Scalar
	for i := 0; i < row; i++ {
		slice := make([]*ristretto255.Scalar, column)
		dualSlice = append(dualSlice, slice)
	}
	return dualSlice
}

func Uint32ToBytes(n uint32) []byte {
	return []byte{
		byte(n),
		byte(n >> 8),
		byte(n >> 16),
		byte(n >> 24),
	}
}

func deepCopyElement(res []*ristretto255.Element) []*ristretto255.Element {
	dst := make([]*ristretto255.Element, len(res))
	for i := 0; i < len(res); i++ {
		var e ristretto255.Element = *res[i]
		dst[i] = &e
	}
	return dst
}

func deepCopyScalar(s *ristretto255.Scalar)*ristretto255.Scalar  {
	var r ristretto255.Scalar
	r = *s
	return &r
}

//Use the precomputed table to acc multiscalarmult
//Scalars have to be sort by (scalars...,G||H)
func (self *InnerProductPDP)MultiScalarMult_GH(scalars []*ristretto255.Scalar) *ristretto255.Element  {
	return new(ristretto255.Element).MultiScalarMult_GH(scalars,self.table)
}
