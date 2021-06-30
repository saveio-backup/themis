package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type GetChanInfo struct {
	ChannelID    uint64
	Participant1 common.Address
	Participant2 common.Address
}

func (this *GetChanInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChannelID); err != nil {
		return fmt.Errorf("[GetChanInfo] [ChannelID:%v] serialize from error:%v", this.ChannelID, err)
	}

	if err := utils.WriteAddress(w, this.Participant1); err != nil {
		return fmt.Errorf("[GetChanInfo] [Participant1:%v] serialize from error:%v", this.Participant1, err)
	}
	if err := utils.WriteAddress(w, this.Participant2); err != nil {
		return fmt.Errorf("[GetChanInfo] [Participant2:%v] serialize from error:%v", this.Participant2, err)
	}

	return nil
}

func (this *GetChanInfo) Deserialize(r io.Reader) error {
	var err error
	if this.ChannelID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[GetChanInfo] [ChannelID] deserialize from error:%v", err)
	}
	if this.Participant1, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[GetChanInfo] [Participant1] deserialize from error:%v", err)
	}
	if this.Participant2, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[GetChanInfo] [Participant2] deserialize from error:%v", err)
	}
	return nil
}

func (this *GetChanInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ChannelID)
	utils.EncodeAddress(sink, this.Participant1)
	utils.EncodeAddress(sink, this.Participant2)

}

func (this *GetChanInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ChannelID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Participant1, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Participant2, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	return nil
}
