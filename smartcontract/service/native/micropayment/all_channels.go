package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type Participants struct {
	ChannelID uint64
	Part1Addr common.Address
	Part2Addr common.Address
}

type AllChannels struct {
	ParticipantNum uint64
	Participants   []Participants
}

func (this *Participants) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChannelID); err != nil {
		return fmt.Errorf("[Participants] [ChannelID:%v] serialize from error:%v", this.ChannelID, err)
	}
	if err := utils.WriteAddress(w, this.Part1Addr); err != nil {
		return fmt.Errorf("[Participants] [Part1Addr:%v] serialize from error:%v", this.Part1Addr, err)
	}
	if err := utils.WriteAddress(w, this.Part2Addr); err != nil {
		return fmt.Errorf("[Participants] [Part2Addr:%v] serialize from error:%v", this.Part2Addr, err)
	}
	return nil
}

func (this *Participants) Deserialize(r io.Reader) error {
	var err error
	if this.ChannelID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Participants] [ChannelID] deserialize from error:%v", err)
	}
	if this.Part1Addr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[Participants] [Part1Addr] deserialize from error:%v", err)
	}
	if this.Part2Addr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[Participants] [Part2Addr] deserialize from error:%v", err)
	}
	return nil
}

func (this *AllChannels) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ParticipantNum); err != nil {
		return fmt.Errorf("[AllChannels] [ParticipantNum:%v] serialize from error:%v", this.ParticipantNum, err)
	}
	for i := uint64(0); i < this.ParticipantNum; i++ {
		participant := this.Participants[i]
		if err := participant.Serialize(w); err != nil {
			return fmt.Errorf("[AllChannels] [Participants:%v Index:%v] serialize from error:%v", participant, i, err)
		}
	}
	return nil
}

func (this *AllChannels) Deserialize(r io.Reader) error {
	var err error
	if this.ParticipantNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[AllChannels] [ParticipantNum] deserialize from error:%v", err)
	}
	for i := uint64(0); i < this.ParticipantNum; i++ {
		var participants Participants
		if err := participants.Deserialize(r); err != nil {
			return fmt.Errorf("[AllChannels] [Participants] deserialize from error:%v", err)
		}
		this.Participants = append(this.Participants, participants)
	}
	return nil
}
