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
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/types"
)

type PrepareRequest struct {
	msgData        ConsensusMessageData
	Nonce          uint64
	NextBookkeeper common.Address
	Transactions   []*types.Transaction
	Signature      []byte
}

func (pr *PrepareRequest) Serialization(sink *common.ZeroCopySink) {
	pr.msgData.Serialization(sink)
	sink.WriteVarUint(pr.Nonce)
	sink.WriteAddress(pr.NextBookkeeper)
	sink.WriteVarUint(uint64(len(pr.Transactions)))
	for _, t := range pr.Transactions {
		t.Serialization(sink)
	}
	sink.WriteVarBytes(pr.Signature)
}

func (pr *PrepareRequest) Deserialization(source *common.ZeroCopySource) error {
	pr.msgData = ConsensusMessageData{}
	err := pr.msgData.Deserialization(source)
	if err != nil {
		return err
	}

	nonce, _, irregular, eof := source.NextVarUint()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	pr.Nonce = nonce
	pr.NextBookkeeper, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}

	var length uint64
	length, _, irregular, eof = source.NextVarUint()
	if eof {
		return io.ErrUnexpectedEOF
	}

	if irregular {
		return common.ErrIrregularData
	}

	for i := 0; i < int(length); i++ {
		var t types.Transaction
		if err := t.Deserialization(source); err != nil {
			return fmt.Errorf("[PrepareRequest] transactions deserialization failed: %s", err)
		}
		pr.Transactions = append(pr.Transactions, &t)
	}

	pr.Signature, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}

func (pr *PrepareRequest) Type() ConsensusMessageType {
	log.Debug()
	return pr.ConsensusMessageData().Type
}

func (pr *PrepareRequest) ViewNumber() byte {
	log.Debug()
	return pr.msgData.ViewNumber
}

func (pr *PrepareRequest) ConsensusMessageData() *ConsensusMessageData {
	log.Debug()
	return &(pr.msgData)
}
