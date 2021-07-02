package bulletproof_pdp_25519

import (
	"bytes"
	"encoding/binary"
	"errors"
	ristretto255 "github.com/saveio/themis/crypto/ristretto"
	"io"
)

const pdpn = blocksize / 64
const blocksize = 256 * 1024
const prooftextsize = 836

type Block struct {
	Buf []byte //Buf大小为固定参数blocksize(256*1024)字节
}

type Challenge struct {
	Index    uint32
	RandSeed uint32
}

type Proof struct {
	iteration int32
	Ls, Rs    []*ristretto255.Element
	a, b      *ristretto255.Scalar
}

func (self *Proof) Serialize() []byte {
	buf := new(bytes.Buffer)
	writeProof(buf, self)
	return buf.Bytes()
}

func (self *Proof) Deserialize(b []byte) error {
	buf := new(bytes.Buffer)
	buf.Write(b)
	return readProof(buf, self)
}

func ScalarToBytes(i *ristretto255.Scalar) [32]byte {
	b := i.Encode([]byte{})
	var result [32]byte
	copy(result[:], b)
	return result
}

func ScalarFromBytes(b [32]byte) *ristretto255.Scalar {
	var s ristretto255.Scalar
	err := s.Decode(b[:])
	if err != nil {
		panic(err)
	}
	return &s
}

func ElementToBytes(element *ristretto255.Element) [32]byte {
	var buf [32]byte
	b := element.Encode([]byte{})
	copy(buf[:], b)
	return buf
}

func ElementFromBytes(buf [32]byte) *ristretto255.Element {
	var element ristretto255.Element
	element.Decode(buf[:])
	return &element
}

// iteration||a||b||Ls||Rs
func writeProof(w io.Writer, proof *Proof) error {
	err := writeElements(w, proof.iteration, ScalarToBytes(proof.a), ScalarToBytes(proof.b))
	if err != nil {
		return err
	}
	for _, l := range proof.Ls {
		err = writeElement(w, ElementToBytes(l))
	}
	if err != nil {
		return err
	}
	for _, r := range proof.Rs {
		err = writeElement(w, ElementToBytes(r))
	}
	if err != nil {
		return err
	}
	return nil

}

func readProof(r io.Reader, proof *Proof) error {
	var bufa, bufb [32]byte
	err := readElements(r, &proof.iteration, &bufa, &bufb)
	if err != nil {
		return err
	}
	proof.a = ScalarFromBytes(bufa)
	proof.b = ScalarFromBytes(bufb)

	for i := 0; i < int(proof.iteration); i++ {
		var bufPoint [32]byte
		err = readElement(r, &bufPoint)
		if err != nil {
			return err
		}
		l := ElementFromBytes(bufPoint)
		proof.Ls = append(proof.Ls, l)
	}

	for i := 0; i < int(proof.iteration); i++ {
		var bufPoint [32]byte
		err = readElement(r, &bufPoint)
		if err != nil {
			return err
		}
		r := ElementFromBytes(bufPoint)
		proof.Rs = append(proof.Rs, r)
	}
	return nil
}

func writeElements(w io.Writer, elements ...interface{}) error {
	for _, element := range elements {
		err := writeElement(w, element)
		if err != nil {
			return err
		}
	}
	return nil
}
func writeElement(w io.Writer, element interface{}) error {
	var scratch [8]byte
	// Attempt to write the element based on the concrete type via fast
	// type assertions first.
	switch e := element.(type) {
	case int32:
		b := scratch[0:4]
		binary.LittleEndian.PutUint32(b, uint32(e))
		_, err := w.Write(b)
		if err != nil {
			return err
		}
		return nil

	case [32]byte:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	// IP address.
	case [33]byte:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil
	}

	return errors.New("invalid element")
}
func readElements(r io.Reader, elements ...interface{}) error {
	for _, element := range elements {
		err := readElement(r, element)
		if err != nil {
			return err
		}
	}
	return nil
}

// readElement reads the next sequence of bytes from r using little endian
// depending on the concrete type of element pointed to.
func readElement(r io.Reader, element interface{}) error {
	var scratch [8]byte
	// Attempt to read the element based on the concrete type via fast
	// type assertions first.
	switch e := element.(type) {
	case *int32:
		b := scratch[0:4]
		_, err := io.ReadFull(r, b)
		if err != nil {
			return err
		}
		*e = int32(binary.LittleEndian.Uint32(b))
		return nil

	case *[32]byte:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil

	// IP address.
	case *[33]byte:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil
	}

	return errors.New("invalid element")
}
