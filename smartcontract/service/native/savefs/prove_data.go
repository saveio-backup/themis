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
package savefs

import (
	"bytes"
	"fmt"
	"io"

	"github.com/saveio/themis/smartcontract/service/native/savefs/pdp"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type ProveData struct {
	Proofs     []byte
	BlockNum   uint64
	Tags       []pdp.Tag         // tags for challenged blocks
	MerklePath []*pdp.MerklePath // merkle path for tags as data
}

func (this *ProveData) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.Proofs); err != nil {
		return fmt.Errorf("[ProveData] [Proof:%v] serialize from error:%v", this.Proofs, err)
	}
	if err := utils.WriteVarUint(w, this.BlockNum); err != nil {
		return fmt.Errorf("[ProveData] [BlockNum:%v] serialize from error:%v", this.BlockNum, err)
	}
	if uint64(len(this.Tags)) != this.BlockNum || uint64(len(this.MerklePath)) != this.BlockNum {
		return fmt.Errorf("[ProveData] [BlockNum:%v] unmatching length", this.BlockNum)
	}
	for _, tag := range this.Tags {
		if err := utils.WriteBytes(w, tag[:]); err != nil {
			return fmt.Errorf("[ProveData] [Tags:%v] serialize from error:%v", tag, err)
		}
	}
	for _, path := range this.MerklePath {
		if err := path.Serialize(w); err != nil {
			return fmt.Errorf("[ProveData] [MerklePath:%v] serialize from error:%v", path, err)
		}
	}
	return nil
}

func (this *ProveData) Deserialize(r io.Reader) error {
	var err error
	if this.Proofs, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[ProveData] [Proofs] deserialize from error:%v", err)
	}
	if this.BlockNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[ProveData] [BlockNum] deserialize from error:%v", err)

	}
	if this.BlockNum == 0 {
		return fmt.Errorf("[ProveData] [BlockNum] BlockNum is 0")
	}
	tags := make([]pdp.Tag, 0)
	path := make([]*pdp.MerklePath, 0)
	for i := uint64(0); i < this.BlockNum; i++ {
		var tag pdp.Tag
		var data []byte
		if data, err = utils.ReadBytes(r); err != nil {
			return fmt.Errorf("[ProveData] [Tag] deserialize from error:%v", err)
		}
		if len(data) != pdp.TAG_LENGTH {
			return fmt.Errorf("[ProveData] [Tag] wrong tag length")
		}
		copy(tag[:], data[:])
		tags = append(tags, tag)
	}
	this.Tags = tags

	for i := uint64(0); i < this.BlockNum; i++ {
		p := new(pdp.MerklePath)
		if err = p.Deserialize(r); err != nil {
			return fmt.Errorf("[ProveData] [MerklePath] deserialize from error:%v", err)
		}
		path = append(path, p)
	}
	this.MerklePath = path
	return nil
}

type ProveParam struct {
	RootHash []byte     // root hash of tag merkle tree
	FileID   pdp.FileID // fileID for pdp proof generation/verification
}

func (this *ProveParam) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.RootHash); err != nil {
		return fmt.Errorf("[ProveParam] [RootHash:%v] serialize from error:%v", this.RootHash, err)
	}
	if err := utils.WriteBytes(w, this.FileID[:]); err != nil {
		return fmt.Errorf("[ProveParam] [FileID:%v] serialize from error:%v", this.FileID, err)
	}
	return nil
}

func (this *ProveParam) Deserialize(r io.Reader) error {
	var err error
	if this.RootHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[ProveParam] [RootHash] deserialize from error:%v", err)
	}

	var fileID pdp.FileID
	var data []byte
	if data, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[ProveParam] [FileID] deserialize from error:%v", err)
	}
	if len(data) != pdp.FILEID_LENGTH {
		return fmt.Errorf("[ProveParam] [FileID] wrong tag length")
	}
	copy(fileID[:], data[:])
	this.FileID = fileID
	return nil
}

func getProveParam(proveParam []byte) (*ProveParam, error) {
	var pp ProveParam
	paramReader := bytes.NewReader(proveParam)
	err := pp.Deserialize(paramReader)
	if err != nil {
		return nil, fmt.Errorf("[GetProveParam] ProveParam deserialize error!")
	}
	return &pp, nil
}
