package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

//ChannelState
const (
	NonExistent = iota
	Opened
	Closed
	Settled
	Removed
)

type ChannelInfo struct {
	ChannelID         uint64
	ChannelState      uint64
	Participant1      Participant
	Participant2      Participant
	SettleBlockHeight uint64
}

func (this *ChannelInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChannelID); err != nil {
		return fmt.Errorf("[ChannelInfo] [ChannelID:%v] serialize from error:%v", this.ChannelID, err)
	}
	if err := utils.WriteVarUint(w, this.ChannelState); err != nil {
		return fmt.Errorf("[ChannelInfo] [ChannelState:%v] serialize from error:%v", this.ChannelState, err)
	}
	if err := this.Participant1.Serialize(w); err != nil {
		return fmt.Errorf("[ChannelInfo] [SenderAddr:%v] serialize from e	rror:%v", this.Participant1, err)
	}
	if err := this.Participant2.Serialize(w); err != nil {
		return fmt.Errorf("[ChannelInfo] [ReceiverAddr:%v] serialize from error:%v", this.Participant2, err)
	}
	if err := utils.WriteVarUint(w, this.SettleBlockHeight); err != nil {
		return fmt.Errorf("[ChannelInfo] [BlockHeight:%v] serialize from error:%v", this.SettleBlockHeight, err)
	}
	return nil
}

func (this *ChannelInfo) Deserialize(r io.Reader) error {
	var err error
	if this.ChannelID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[ChannelInfo] [ChannelID] deserialize from error:%v", err)
	}
	if this.ChannelState, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[ChannelInfo] [ChannelState] deserialize from error:%v", err)
	}
	if err = this.Participant1.Deserialize(r); err != nil {
		return fmt.Errorf("[ChannelInfo] [Participant1] deserialize from error:%v", err)
	}
	if err = this.Participant2.Deserialize(r); err != nil {
		return fmt.Errorf("[ChannelInfo] [Participant2] deserialize from error:%v", err)
	}
	if this.SettleBlockHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[ChannelInfo] [SettleBlockHeight] deserialize from error:%v", err)
	}
	return nil
}

func (this *ChannelInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ChannelID)
	utils.EncodeVarUint(sink, this.ChannelState)
	this.Participant1.Serialization(sink)
	this.Participant2.Serialization(sink)
	utils.EncodeVarUint(sink, this.SettleBlockHeight)
}

func (this *ChannelInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ChannelID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ChannelState, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	err = this.Participant1.Deserialization(source)
	if err != nil {
		return err
	}
	err = this.Participant2.Deserialization(source)
	if err != nil {
		return err
	}
	this.SettleBlockHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}
