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

package dbft

import (
	"io"

	"github.com/saveio/themis/common"
)

type PrepareResponse struct {
	msgData   ConsensusMessageData
	Signature []byte
}

func (pres *PrepareResponse) Serialization(sink *common.ZeroCopySink) {
	pres.msgData.Serialization(sink)
	sink.WriteVarBytes(pres.Signature)
}

//read data to reader
func (pres *PrepareResponse) Deserialization(source *common.ZeroCopySource) error {
	err := pres.msgData.Deserialization(source)
	if err != nil {
		return err
	}

	sign, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	pres.Signature = sign

	return nil
}

func (pres *PrepareResponse) Type() ConsensusMessageType {
	return pres.ConsensusMessageData().Type
}

func (pres *PrepareResponse) ViewNumber() byte {
	return pres.msgData.ViewNumber
}

func (pres *PrepareResponse) ConsensusMessageData() *ConsensusMessageData {
	return &(pres.msgData)
}
