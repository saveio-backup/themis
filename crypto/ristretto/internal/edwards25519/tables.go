// Copyright (c) 2019 Henry de Valence.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edwards25519

import (
	"crypto/subtle"
)

// A dynamic lookup table for variable-base, constant-time scalar muls.
type projLookupTable struct {
	points [8]ProjCached
}

// A precomputed lookup table for fixed-base, constant-time scalar muls.
type affineLookupTable struct {
	points [8]AffineCached
}

// A dynamic lookup table for variable-base, variable-time scalar muls.
type nafLookupTable5 struct {
	points [8]ProjCached
}

type NafLookupTable8Pro struct {
	points [64]ProjCached
}

type NafLookupTable10Pro struct {
	points [16384]ProjCached
}

// A precomputed lookup table for fixed-base, variable-time scalar muls.
type nafLookupTable8 struct {
	points [64]AffineCached
}

// Constructors.

// Builds a lookup table at runtime. Fast.
func (v *projLookupTable) FromP3(q *ProjP3) {
	// Goal: v.points[i] = (i+1)*Q, i.e., Q, 2Q, ..., 8Q
	// This allows lookup of -8Q, ..., -Q, 0, Q, ..., 8Q
	v.points[0].FromP3(q)
	tmpP3 := ProjP3{}
	tmpP1xP1 := ProjP1xP1{}
	for i := 0; i < 7; i++ {
		// Compute (i+1)*Q as Q + i*Q and convert to a ProjCached
		// This is needlessly complicated because the API has explicit
		// recievers instead of creating stack objects and relying on RVO
		v.points[i+1].FromP3(tmpP3.FromP1xP1(tmpP1xP1.Add(q, &v.points[i])))
	}
}

// This is not optimised for speed; affine tables should be precomputed.
func (v *affineLookupTable) FromP3(q *ProjP3) {
	// Goal: v.points[i] = (i+1)*Q, i.e., Q, 2Q, ..., 8Q
	// This allows lookup of -8Q, ..., -Q, 0, Q, ..., 8Q
	v.points[0].FromP3(q)
	tmpP3 := ProjP3{}
	tmpP1xP1 := ProjP1xP1{}
	for i := 0; i < 7; i++ {
		// Compute (i+1)*Q as Q + i*Q and convert to AffineCached
		v.points[i+1].FromP3(tmpP3.FromP1xP1(tmpP1xP1.AddAffine(q, &v.points[i])))
	}
}

// Builds a lookup table at runtime. Fast.
func (v *nafLookupTable5) FromP3(q *ProjP3) {
	// Goal: v.points[i] = (2*i+1)*Q, i.e., Q, 3Q, 5Q, ..., 15Q
	// This allows lookup of -15Q, ..., -3Q, -Q, 0, Q, 3Q, ..., 15Q
	v.points[0].FromP3(q)
	q2 := ProjP3{}
	q2.Add(q, q)
	tmpP3 := ProjP3{}
	tmpP1xP1 := ProjP1xP1{}
	for i := 0; i < 7; i++ {
		v.points[i+1].FromP3(tmpP3.FromP1xP1(tmpP1xP1.Add(&q2, &v.points[i])))
	}
}

func (v *NafLookupTable8Pro)FromP3(q *ProjP3)  {
	v.points[0].FromP3(q)
	q2 := ProjP3{}
	q2.Add(q, q)
	tmpP3 := ProjP3{}
	tmpP1xP1 := ProjP1xP1{}
	for i := 0; i < 63; i++ {
		v.points[i+1].FromP3(tmpP3.FromP1xP1(tmpP1xP1.Add(&q2, &v.points[i])))
	}
}

func (v *NafLookupTable10Pro)FromP3(q *ProjP3)  {
	v.points[0].FromP3(q)
	q2 := ProjP3{}
	q2.Add(q, q)
	tmpP3 := ProjP3{}
	tmpP1xP1 := ProjP1xP1{}
	for i := 0; i < 16383; i++ {
		v.points[i+1].FromP3(tmpP3.FromP1xP1(tmpP1xP1.Add(&q2, &v.points[i])))
	}
}

// This is not optimised for speed; affine tables should be precomputed.
func (v *nafLookupTable8) FromP3(q *ProjP3) {
	v.points[0].FromP3(q)
	q2 := ProjP3{}
	q2.Add(q, q)
	tmpP3 := ProjP3{}
	tmpP1xP1 := ProjP1xP1{}
	for i := 0; i < 63; i++ {
		v.points[i+1].FromP3(tmpP3.FromP1xP1(tmpP1xP1.AddAffine(&q2, &v.points[i])))
	}
}

// Selectors.

// Set dest to x*Q, where -8 <= x <= 8, in constant time.
func (v *projLookupTable) SelectInto(dest *ProjCached, x int8) {
	// Compute xabs = |x|
	xmask := x >> 7
	xabs := uint8((x + xmask) ^ xmask)

	dest.Zero()
	for j := 1; j <= 8; j++ {
		// Set dest = j*Q if |x| = j
		cond := subtle.ConstantTimeByteEq(xabs, uint8(j))
		dest.Select(&v.points[j-1], dest, cond)
	}
	// Now dest = |x|*Q, conditionally negate to get x*Q
	dest.CondNeg(int(xmask & 1))
}

// Set dest to x*Q, where -8 <= x <= 8, in constant time.
func (v *affineLookupTable) SelectInto(dest *AffineCached, x int8) {
	// Compute xabs = |x|
	xmask := x >> 7
	xabs := uint8((x + xmask) ^ xmask)

	dest.Zero()
	for j := 1; j <= 8; j++ {
		// Set dest = j*Q if |x| = j
		cond := subtle.ConstantTimeByteEq(xabs, uint8(j))
		dest.Select(&v.points[j-1], dest, cond)
	}
	// Now dest = |x|*Q, conditionally negate to get x*Q
	dest.CondNeg(int(xmask & 1))
}

// Given odd x with 0 < x < 2^4, return x*Q (in variable time).
func (v *nafLookupTable5) SelectInto(dest *ProjCached, x int8) {
	*dest = v.points[x/2]
}

// Given odd x with 0 < x < 2^7, return x*Q (in variable time).
func (v *NafLookupTable8Pro) SelectInto(dest *ProjCached, x int8) {
	*dest = v.points[x/2]
}

func (v *NafLookupTable10Pro) SelectInto(dest *ProjCached, x int) {
	*dest = v.points[x/2]
}

// Given odd x with 0 < x < 2^7, return x*Q (in variable time).
func (v *nafLookupTable8) SelectInto(dest *AffineCached, x int8) {
	*dest = v.points[x/2]
}
