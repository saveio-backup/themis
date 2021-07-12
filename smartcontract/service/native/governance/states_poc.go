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

type MiningView struct {
	View   uint32
	Height uint32
}

func (this *MiningView) Serialize(w io.Writer) error {
	if err := serialization.WriteUint64(w, uint64(this.View)); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize view error: %v", err)
	}
	if err := serialization.WriteUint32(w, this.Height); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize height error: %v", err)
	}

	return nil
}

func (this *MiningView) Deserialize(r io.Reader) error {
	view, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize view error: %v", err)
	}
	height, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize height error: %v", err)
	}

	this.View = uint32(view)
	this.Height = height

	return nil
}

type MiningViewInfo struct {
	//Info for mined epoch
	GenerationSignature common.Uint256
	Generator           uint64

	//Info for next mining epoch
	NewGenerationSignature common.Uint256
	Scoop                  uint32
	BaseTarget             int64
}

func (this *MiningViewInfo) Serialize(w io.Writer) error {
	if err := this.GenerationSignature.Serialize(w); err != nil {
		return fmt.Errorf("LastGenerationSignature.Serialize, serialize lastGenerationSignature error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.Generator); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize lastGenerator error: %v", err)
	}

	if err := this.NewGenerationSignature.Serialize(w); err != nil {
		return fmt.Errorf("NewGenerationSignature.Serialize, serialize generationSignature error: %v", err)
	}
	if err := serialization.WriteUint64(w, uint64(this.Scoop)); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize view error: %v", err)
	}
	if err := serialization.WriteUint64(w, uint64(this.BaseTarget)); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize baseTarget error: %v", err)
	}

	return nil
}

func (this *MiningViewInfo) Deserialize(r io.Reader) error {
	generationSignature := new(common.Uint256)
	if err := generationSignature.Deserialize(r); err != nil {
		return fmt.Errorf("generationSignature.Deserialize, deserialize generationSignature error: %v", err)
	}
	generator, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize generator error: %v", err)
	}
	newGenerationSignature := new(common.Uint256)
	if err := newGenerationSignature.Deserialize(r); err != nil {
		return fmt.Errorf("newGenerationSignature.Deserialize, deserialize lastGenerationSignature error: %v", err)
	}
	scoop, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize scoop error: %v", err)
	}
	baseTarget, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize baseTarget error: %v", err)
	}

	this.GenerationSignature = *generationSignature
	this.Generator = generator

	this.NewGenerationSignature = *newGenerationSignature
	this.Scoop = uint32(scoop)
	this.BaseTarget = int64(baseTarget)

	return nil
}

//winner info
type WinnerInfo struct {
	View     uint32
	Address  common.Address
	Deadline uint64

	//vote info
	VoteConsPub []string
	VoteId      []uint32
	VoteInfo    []byte
}

func (this *WinnerInfo) Serialize(w io.Writer) error {
	if err := serialization.WriteUint64(w, uint64(this.View)); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize view error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	if err := serialization.WriteUint64(w, uint64(this.Deadline)); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize difficulty error: %v", err)
	}

	//cons vote by pub key
	if err := serialization.WriteUint64(w, uint64(len(this.VoteConsPub))); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize num VoteCons pubkey error: %v", err)
	}
	for i := 0; i < len(this.VoteConsPub); i++ {
		if err := serialization.WriteString(w, this.VoteConsPub[i]); err != nil {
			return fmt.Errorf("serialization.WriteString, request peerPubkey error: %v", err)
		}
	}

	if err := serialization.WriteUint64(w, uint64(len(this.VoteId))); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize num VoteCons pubkey error: %v", err)
	}
	for i := 0; i < len(this.VoteId); i++ {
		if err := serialization.WriteUint64(w, uint64(this.VoteId[i])); err != nil {
			return fmt.Errorf("serialization.WriteUint64, serialize VoteId error: %v", err)
		}
	}

	if err := serialization.WriteVarBytes(w, this.VoteInfo); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize VoteInfo error: %v", err)
	}

	return nil
}

func (this *WinnerInfo) Deserialize(r io.Reader) error {
	view, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize generator error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	deadline, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize height error: %v", err)
	}

	this.View = uint32(view)
	this.Address = address
	this.Deadline = deadline

	//vote info pubkey
	voteConsPubLen, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize num VoteCons pubkey error: %v", err)
	}
	this.VoteConsPub = make([]string, 0, voteConsPubLen)
	for i := 0; i < int(voteConsPubLen); i++ {
		peerPubkey, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
		}
		this.VoteConsPub = append(this.VoteConsPub, peerPubkey)

	}

	voteIdLen, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize num voteIdLen error: %v", err)
	}
	this.VoteId = make([]uint32, 0, voteIdLen)

	for i := 0; i < int(voteIdLen); i++ {
		value, err := serialization.ReadUint64(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadUint64, deserialize VoteId error: %v", err)
		}
		this.VoteId = append(this.VoteId, uint32(value))
	}
	return nil
}

type PeriodInfo struct {
	Period   uint32 // verification period
	PlotSize uint64
}

func (this *PeriodInfo) Serialize(w io.Writer) error {
	if err := serialization.WriteUint64(w, uint64(this.Period)); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize view error: %v", err)
	}

	if err := serialization.WriteUint64(w, uint64(this.PlotSize)); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize difficulty error: %v", err)
	}
	return nil
}

func (this *PeriodInfo) Deserialize(r io.Reader) error {
	period, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize generator error: %v", err)
	}

	plotSize, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize height error: %v", err)
	}

	this.Period = uint32(period)
	this.PlotSize = plotSize

	return nil
}

type PeriodInfos struct {
	Address   common.Address
	prePeriod PeriodInfo
	curPeriod PeriodInfo
}

func (this *PeriodInfos) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	if err := this.prePeriod.Serialize(w); err != nil {
		return fmt.Errorf("PeriodInfos serialize prePeriod error: %v", err)
	}
	if err := this.curPeriod.Serialize(w); err != nil {
		return fmt.Errorf("PeriodInfos serialize curPeriod error: %v", err)
	}
	return nil
}

func (this *PeriodInfos) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	if err := this.prePeriod.Deserialize(r); err != nil {
		return fmt.Errorf("PeriodInfos deserialize prePeriod error: %v", err)
	}
	if err := this.curPeriod.Deserialize(r); err != nil {
		return fmt.Errorf("PeriodInfos deserialize curPeriod error: %v", err)
	}
	this.Address = address
	return nil
}

type PeriodSummary struct {
	AvgDeadline   uint64
	MinerWinTimes map[common.Address]int64
}

func (this *PeriodSummary) Serialize(w io.Writer) error {
	if err := serialization.WriteUint64(w, uint64(this.AvgDeadline)); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize avg deadline error: %v", err)
	}
	if err := serialization.WriteUint32(w, uint32(len(this.MinerWinTimes))); err != nil {
		return fmt.Errorf("serialization.WriteUint32 serialize win times map length error: %v", err)
	}
	for address, v := range this.MinerWinTimes {
		if err := address.Serialize(w); err != nil {
			return fmt.Errorf("address.Serialize, serialize address error: %v", err)
		}
		if err := serialization.WriteUint64(w, uint64(v)); err != nil {
			return fmt.Errorf("serialization.WriteUint64, serialize times error: %v", err)
		}
	}
	return nil
}

func (this *PeriodSummary) Deserialize(r io.Reader) error {
	avgDeadline, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize avg deadline error: %v", err)
	}

	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize win times map size error: %v", err)
	}

	info := make(map[common.Address]int64)
	for i := 0; uint32(i) < n; i++ {
		address := new(common.Address)
		err := address.Deserialize(r)
		if err != nil {
			return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
		}

		times, err := serialization.ReadUint64(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadUint64, deserialize win times error: %v", err)
		}

		info[*address] = int64(times)
	}

	this.AvgDeadline = avgDeadline
	this.MinerWinTimes = info
	return nil
}

//vote info for consensus group nodes
type ConsVoteItem struct {
	PeerPubkey string //peer pubkey
	NumVotes   uint32
	VoterMap   map[common.Address]uint32
}

func (this *ConsVoteItem) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteUint32(w, this.NumVotes); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize address error: %v", err)
	}

	if err := serialization.WriteUint32(w, uint32(len(this.VoterMap))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize NumVotes error: %v", err)
	}
	for address, v := range this.VoterMap {
		if err := address.Serialize(w); err != nil {
			return fmt.Errorf("address.Serialize, serialize address error: %v", err)
		}
		if err := serialization.WriteUint32(w, v); err != nil {
			return fmt.Errorf("serialization.WriteUint32, serialize num votes error: %v", err)
		}
	}
	return nil
}

func (this *ConsVoteItem) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	numVotes, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize numvote error: %v", err)
	}

	numVoter, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize numVoter error: %v", err)
	}
	voterMap := make(map[common.Address]uint32)
	for i := 0; uint32(i) < numVoter; i++ {
		address := common.Address{}
		err := address.Deserialize(r)
		if err != nil {
			return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
		}

		num, err := serialization.ReadUint32(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadUint32, deserialize num votes error: %v", err)
		}

		voterMap[address] = uint32(num)
	}

	this.PeerPubkey = peerPubkey
	this.NumVotes = numVotes
	this.VoterMap = voterMap

	return nil
}

type ConsVoteMap struct {
	ConsVoteMap map[string]*ConsVoteItem
}

func (this *ConsVoteMap) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.ConsVoteMap))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize ConsVoteMap length error: %v", err)
	}

	for _, v := range this.ConsVoteMap {
		if err := v.Serialize(w); err != nil {
			return fmt.Errorf("serialize consVoteMap error: %v", err)
		}
	}
	return nil
}

func (this *ConsVoteMap) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize ConsVoteMap length error: %v", err)
	}
	consVoteMap := make(map[string]*ConsVoteItem)
	for i := 0; uint32(i) < n; i++ {
		consVoteItem := new(ConsVoteItem)
		if err := consVoteItem.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		consVoteMap[consVoteItem.PeerPubkey] = consVoteItem
	}
	this.ConsVoteMap = consVoteMap
	return nil
}

//vote detail of miner
type ConsVoteDetail struct {
	ConsVoteDetail map[string]int
}

func (this *ConsVoteDetail) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.ConsVoteDetail))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize ConsVoteDetail length error: %v", err)
	}

	for pubkey, _ := range this.ConsVoteDetail {
		if err := serialization.WriteString(w, pubkey); err != nil {
			return fmt.Errorf("serialization.WriteString, serialize votee pubkey error: %v", err)
		}
	}
	return nil
}

func (this *ConsVoteDetail) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize ConsVoteDetail length error: %v", err)
	}
	consVoteDetail := make(map[string]int)
	for i := 0; uint32(i) < n; i++ {
		pubkey, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadString, deserialize votee pubkey error: %v", err)
		}

		consVoteDetail[pubkey] = 1
	}
	this.ConsVoteDetail = consVoteDetail
	return nil
}

//key is pubkey of node included in consensus group
type ConsGroupItems struct {
	ConsGroupItems map[string]int
}

func (this *ConsGroupItems) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.ConsGroupItems))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize ConsGroupItems length error: %v", err)
	}

	for k, _ := range this.ConsGroupItems {
		if err := serialization.WriteString(w, k); err != nil {
			return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
		}
	}
	return nil
}

func (this *ConsGroupItems) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize consGroupItems length error: %v", err)
	}
	consGroupItems := make(map[string]int)
	for i := 0; uint32(i) < n; i++ {
		peerPubkey, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
		}
		consGroupItems[peerPubkey] = 1
	}
	this.ConsGroupItems = consGroupItems
	return nil
}

type DefaultConsNodes struct {
	DefaultConsNodes map[string]int
}

func (this *DefaultConsNodes) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.DefaultConsNodes))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize ConsGroupItems length error: %v", err)
	}

	for k, _ := range this.DefaultConsNodes {
		if err := serialization.WriteString(w, k); err != nil {
			return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
		}
	}
	return nil
}

func (this *DefaultConsNodes) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize consGroupItems length error: %v", err)
	}
	defaultConsNodes := make(map[string]int)
	for i := 0; uint32(i) < n; i++ {
		peerPubkey, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
		}
		defaultConsNodes[peerPubkey] = 1
	}
	this.DefaultConsNodes = defaultConsNodes
	return nil
}

type ConsVoteRevenue struct {
	Total uint64
}

func (this *ConsVoteRevenue) Serialize(w io.Writer) error {
	if err := serialization.WriteUint64(w, this.Total); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize vote revenue error: %v", err)
	}

	return nil
}

func (this *ConsVoteRevenue) Deserialize(r io.Reader) error {
	total, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize vote revenue error: %v", err)
	}

	this.Total = total
	return nil
}
