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

package governance

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/serialization"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type MiningInfo struct {
	View                uint32
	BaseTarget          int64
	GenerationSignature common.Uint256
}

func (this *MiningInfo) Serialize(w io.Writer) error {
	if err := serialization.WriteUint64(w, uint64(this.View)); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize view error: %v", err)
	}
	if err := serialization.WriteUint64(w, uint64(this.BaseTarget)); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize baseTarget error: %v", err)
	}
	if err := this.GenerationSignature.Serialize(w); err != nil {
		return fmt.Errorf("GenerationSignature.Serialize, serialize generationSignature error: %v", err)
	}
	return nil
}

func (this *MiningInfo) Deserialize(r io.Reader) error {
	view, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize view error: %v", err)
	}
	baseTarget, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize baseTarget error: %v", err)
	}
	generationSignature := new(common.Uint256)
	if err := generationSignature.Deserialize(r); err != nil {
		return fmt.Errorf("lastGenerationSignature.Deserialize, deserialize lastGenerationSignature error: %v", err)
	}

	this.View = uint32(view)
	this.BaseTarget = int64(baseTarget)
	this.GenerationSignature = *generationSignature

	return nil
}

type SubmitNonceParam struct {
	View     uint32
	Address  common.Address
	Id       int64
	Nonce    uint64
	Deadline uint64
	PlotName string

	//vote info
	VoteConsPub []string
	VoteId      []uint32
	VoteInfo    []byte

	//move up consensus elect
	MoveUpElect bool
}

func (this *SubmitNonceParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.View)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize View error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Id)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize Id error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Nonce)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize Nonce error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Deadline)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize Deadline error: %v", err)
	}
	if err := serialization.WriteString(w, this.PlotName); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize PlotName error: %v", err)
	}

	//cons vote by pub key
	if err := utils.WriteVarUint(w, uint64(len(this.VoteConsPub))); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize num VoteCons pubkey error: %v", err)
	}
	for i := 0; i < len(this.VoteConsPub); i++ {
		if err := serialization.WriteString(w, this.VoteConsPub[i]); err != nil {
			return fmt.Errorf("serialization.WriteString, request peerPubkey error: %v", err)
		}
	}

	if err := utils.WriteVarUint(w, uint64(len(this.VoteId))); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize num VoteCons pubkey error: %v", err)
	}
	for i := 0; i < int(len(this.VoteId)); i++ {
		if err := utils.WriteVarUint(w, uint64(this.VoteId[i])); err != nil {
			return fmt.Errorf("utils.WriteVarUint, serialize VoteId error: %v", err)
		}
	}

	if err := serialization.WriteVarBytes(w, this.VoteInfo); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize VoteInfo error: %v", err)
	}

	//move up consensus election
	if err := utils.WriteBool(w, this.MoveUpElect); err != nil {
		return fmt.Errorf("utils.WriteBool, serialize move up elect error:%v", err)
	}

	return nil
}

func (this *SubmitNonceParam) Deserialize(r io.Reader) error {
	view, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize View error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	id, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize Id error: %v", err)
	}
	nonce, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize Nonce error: %v", err)
	}
	deadline, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize Deadline error: %v", err)
	}
	plotName, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize PlotName error: %v", err)
	}

	this.View = uint32(view)
	this.Address = address
	this.Id = int64(id)
	this.Nonce = nonce
	this.Deadline = deadline
	this.PlotName = plotName

	//vote info pubkey
	voteConsPubLen, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize num VoteCons pubkey error: %v", err)
	}
	this.VoteConsPub = make([]string, 0, voteConsPubLen)
	for i := 0; i < int(voteConsPubLen); i++ {
		peerPubkey, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
		}
		this.VoteConsPub = append(this.VoteConsPub, peerPubkey)

	}

	voteIdLen, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize num voteId error: %v", err)
	}

	this.VoteId = make([]uint32, 0, voteIdLen)
	for i := 0; i < int(voteIdLen); i++ {
		value, err := utils.ReadVarUint(r)
		if err != nil {
			return fmt.Errorf("utils.ReadVarUint, deserialize VoteId error: %v", err)
		}
		this.VoteId = append(this.VoteId, uint32(value))
	}

	this.VoteInfo, err = serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize VoteInfo error: %v", err)
	}

	//move up consensus election
	moveup, err := utils.ReadBool(r)
	if err != nil {
		return fmt.Errorf("utils.ReadBool deserialize goverance error:%v", err)
	}
	this.MoveUpElect = moveup

	return nil
}

func (this *SubmitNonceParam) Hash() common.Uint256 {

	buf := new(bytes.Buffer)
	this.Serialize(buf)

	temp := sha256.Sum256(buf.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))

	return hash
}

type WinnerInfoReq struct {
	View uint32
}

func (this *WinnerInfoReq) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.View)); err != nil {
		return fmt.Errorf("serialize view len error:%v", err)
	}
	return nil
}

func (this *WinnerInfoReq) Deserialize(r io.Reader) error {
	view, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize View error: %v", err)
	}

	this.View = uint32(view)

	return nil
}
