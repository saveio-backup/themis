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

package ontfs

import (
	"fmt"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/ontfs/pdp"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type PdpData struct {
	NodeAddr        common.Address
	FileHash        []byte
	ProveData       []byte
	ChallengeHeight uint64
}

func (this *PdpData) Serialization(sink *common.ZeroCopySink) error {
	if len(this.ProveData) < pdp.VersionLength {
		return fmt.Errorf("PdpData Serialization error: ProveData length shorter than 8")
	}
	utils.EncodeAddress(sink, this.NodeAddr)
	sink.WriteVarBytes(this.FileHash)
	sink.WriteVarBytes(this.ProveData)
	utils.EncodeVarUint(sink, this.ChallengeHeight)
	return nil
}

func (this *PdpData) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.NodeAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.FileHash, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.ProveData, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.ChallengeHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}
