/*
 * Copyright (C) 2019 The themis Authors
 * This file is part of The themis library.
 *
 * The themis is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The themis is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */

package merkle

import "math/bits"

// return the number of 1 bit
func countBit(num uint32) uint {
	return uint(bits.OnesCount32(num))
}

func isPower2(num uint32) bool {
	return countBit(num) == 1
}

// return the position of the heightest 1 bit
// 1-based index
func highBit(num uint32) uint {
	return uint(32 - bits.LeadingZeros32(num))
}

// return the position of the lowest 1 bit
// 1-based index
func lowBit(num uint32) uint {
	return highBit(num & -num)
}
