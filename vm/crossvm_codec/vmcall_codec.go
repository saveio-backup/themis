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

package crossvm_codec

import (
	"bytes"

	"github.com/saveio/themis/common"
)

//input byte array should be the following format
// version(1byte) + type(1byte) + data...
func DeserializeCallParam(input []byte) (interface{}, error) {
	if !bytes.HasPrefix(input, []byte{0}) {
		return nil, ERROR_PARAM_FORMAT
	}

	source := common.NewZeroCopySource(input[1:])
	return DecodeValue(source)
}
