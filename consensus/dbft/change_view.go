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

type ChangeView struct {
	msgData       ConsensusMessageData
	NewViewNumber byte
}

func (cv *ChangeView) Serialization(sink *common.ZeroCopySink) {
	cv.msgData.Serialization(sink)
	sink.WriteByte(cv.NewViewNumber)
}

//read data to reader
func (cv *ChangeView) Deserialization(source *common.ZeroCopySource) error {
	err := cv.msgData.Deserialization(source)
	if err != nil {
		return err
	}

	viewNum, eof := source.NextByte()
	if eof {
		return io.ErrUnexpectedEOF
	}
	cv.NewViewNumber = viewNum

	return nil
}

func (cv *ChangeView) Type() ConsensusMessageType {
	return cv.ConsensusMessageData().Type
}

func (cv *ChangeView) ViewNumber() byte {
	return cv.msgData.ViewNumber
}

func (cv *ChangeView) ConsensusMessageData() *ConsensusMessageData {
	return &(cv.msgData)
}
