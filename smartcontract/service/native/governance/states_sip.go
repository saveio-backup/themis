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
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/serialization"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

//sip
type SIP struct {
	Digest string
	Index  uint32

	Height   uint32
	Detail   []byte
	Default  byte
	MinVotes uint32
	Bonus    uint64

	RegHeight uint32
	NumVotes  uint32
	Result    byte

	VoterMap  map[common.Address]uint32
	BonusDone bool
}

func (this *SIP) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Digest); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize Digest error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Index)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize Index error: %v", err)
	}

	if err := utils.WriteVarUint(w, uint64(this.Height)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize Height error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Detail); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize Detail error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Default)); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize Default error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.MinVotes)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize MinVotes error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Bonus)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize Bonus error: %v", err)
	}

	if err := utils.WriteVarUint(w, uint64(this.RegHeight)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize RegHeight error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.NumVotes)); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize NumVotes error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Result)); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize Result error: %v", err)
	}

	if err := utils.WriteVarUint(w, uint64(len(this.VoterMap))); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize NumVotes error: %v", err)
	}

	for address, v := range this.VoterMap {
		if err := address.Serialize(w); err != nil {
			return fmt.Errorf("address.Serialize, serialize address error: %v", err)
		}

		if err := utils.WriteVarUint(w, uint64(v)); err != nil {
			return fmt.Errorf("utils.WriteVarUint, serialize RegHeight error: %v", err)
		}
	}
	if err := utils.WriteBool(w, this.BonusDone); err != nil {
		return fmt.Errorf("WriteBool, serialize BonusDone error:%v", err)
	}
	return nil
}

func (this *SIP) Deserialize(r io.Reader) error {
	digest, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize digest error: %v", err)
	}
	index, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize index error: %v", err)
	}

	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize height error: %v", err)
	}
	detail, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize detail error: %v", err)
	}
	defaultVal, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize default error: %v", err)
	}
	minVotes, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.WriteVarUint, deserialize minVotes error: %v", err)
	}
	bonus, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.WriteVarUint, deserialize bonus error: %v", err)
	}

	regHeight, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.WriteVarUint, deserialize regHeight error: %v", err)
	}
	numVotes, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize numVotes error: %v", err)
	}
	result, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize result error: %v", err)
	}

	numVoter, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize numVotes error: %v", err)
	}

	voterMap := make(map[common.Address]uint32)
	for i := 0; uint64(i) < numVoter; i++ {
		address := common.Address{}
		err := address.Deserialize(r)
		if err != nil {
			return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
		}

		num, err := utils.ReadVarUint(r)
		if err != nil {
			return fmt.Errorf("utils.ReadVarUint, deserialize height error: %v", err)
		}

		voterMap[address] = uint32(num)
	}
	bonusDone, err := utils.ReadBool(r)
	if err != nil {
		return fmt.Errorf("[FileInfo] [ValidFlag] deserialize from error:%v", err)
	}

	this.Digest = digest
	this.Index = uint32(index)
	this.Height = uint32(height)
	this.Detail = detail
	this.Default = byte(defaultVal)
	this.MinVotes = uint32(minVotes)
	this.Bonus = bonus

	this.RegHeight = uint32(regHeight)
	this.NumVotes = uint32(numVotes)
	this.Result = byte(result)

	this.VoterMap = voterMap
	this.BonusDone = bonusDone

	return nil
}

type SipMap struct {
	SipMap map[uint32]*SIP
}

func (this *SipMap) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.SipMap))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize PeerPoolMap length error: %v", err)
	}
	var sipList []*SIP
	for _, v := range this.SipMap {
		sipList = append(sipList, v)
	}

	for _, v := range sipList {
		if err := v.Serialize(w); err != nil {
			return fmt.Errorf("serialize SIP error: %v", err)
		}
	}
	return nil
}

func (this *SipMap) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize PeerPoolMap length error: %v", err)
	}
	SipMap := make(map[uint32]*SIP)
	for i := 0; uint32(i) < n; i++ {
		sip := new(SIP)
		if err := sip.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		SipMap[sip.Index] = sip
	}
	this.SipMap = SipMap
	return nil
}

type SIPVoteRevenue struct {
	Total   uint64
	Reserve uint64
}

func (this *SIPVoteRevenue) Serialize(w io.Writer) error {
	if err := serialization.WriteUint64(w, this.Total); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize vote revenue error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.Total); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize vote revenue error: %v", err)
	}

	return nil
}

func (this *SIPVoteRevenue) Deserialize(r io.Reader) error {
	total, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize vote revenue error: %v", err)
	}

	this.Total = total
	return nil
}
