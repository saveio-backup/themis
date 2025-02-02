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

package common

import "fmt"

/* common - print notice when user need to choose or enter something */
func PrintNotice(name string) {
	switch name {
	case "key type":
		fmt.Printf(`
Select a signature algorithm from the following:

  1  ECDSA
  2  SM2
  3  Ed25519

[default is 1]: `)

	case "curve":
		fmt.Printf(`
Select a curve from the following:

    | NAME  | KEY LENGTH (bits)
 ---|-------|------------------
  1 | P-224 | 224
  2 | P-256 | 256

This determines the length of the private key [default is 2]: `)

	case "signature-scheme":
		fmt.Printf(`
Select a signature scheme from the following:

  1  SHA224withECDSA
  2  SHA256withECDSA
  5  SHA3-224withECDSA
  6  SHA3-256withECDSA
  9  RIPEMD160withECDSA

This can be changed later [default is 2]: `)

	default:
	}
}
