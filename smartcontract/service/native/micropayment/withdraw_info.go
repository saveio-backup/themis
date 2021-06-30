package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type WithDraw struct {
	ChannelID         uint64
	Participant       common.Address
	Partner           common.Address
	TotalWithdraw     uint64
	ParticipantSig    []byte
	ParticipantPubKey []byte
	PartnerSig        []byte
	PartnerPubKey     []byte
}

func (this *WithDraw) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChannelID); err != nil {
		return fmt.Errorf("[Withdraw] [ChannelID:%v] serialize from error:%v", this.ChannelID, err)
	}
	if err := utils.WriteAddress(w, this.Participant); err != nil {
		return fmt.Errorf("[Withdraw] [Participant:%v] serialize from error:%v", this.Participant, err)
	}
	if err := utils.WriteAddress(w, this.Partner); err != nil {
		return fmt.Errorf("[Withdraw] [withdraw_amount:%v] serialize from error:%v", this.Partner, err)
	}
	if err := utils.WriteVarUint(w, this.TotalWithdraw); err != nil {
		return fmt.Errorf("[Withdraw] [TotalWithdraw:%v] serialize from error:%v", this.TotalWithdraw, err)
	}
	if err := utils.WriteBytes(w, this.ParticipantSig); err != nil {
		return fmt.Errorf("[Withdraw] [Port:%v] serialize from error:%v", this.ParticipantSig, err)
	}
	if err := utils.WriteBytes(w, this.ParticipantPubKey); err != nil {
		return fmt.Errorf("[Withdraw] [ParticipantPubKey:%v] serialize from error:%v", this.ParticipantPubKey, err)
	}
	if err := utils.WriteBytes(w, this.PartnerSig); err != nil {
		return fmt.Errorf("[Withdraw] [PartnerSig:%v] serialize from error:%v", this.PartnerSig, err)
	}
	if err := utils.WriteBytes(w, this.PartnerPubKey); err != nil {
		return fmt.Errorf("[Withdraw] [PartnerPubKey:%v] serialize from error:%v", this.PartnerPubKey, err)
	}

	return nil
}

func (this *WithDraw) Deserialize(r io.Reader) error {
	var err error
	if this.ChannelID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Withdraw] [ChannelID] deserialize from error:%v", err)
	}
	if this.Participant, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[Withdraw] [Participant] deserialize from error:%v", err)
	}
	if this.Partner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[Withdraw] [Partner] deserialize from error:%v", err)
	}
	if this.TotalWithdraw, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Withdraw] [TotalWithdraw] deserialize from error:%v", err)
	}
	if this.ParticipantSig, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[Withdraw] [ParticipantSig] deserialize from error:%v", err)
	}
	if this.ParticipantPubKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[Withdraw] [ParticipantPubKey] deserialize from error:%v", err)
	}
	if this.PartnerSig, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[Withdraw] [PartnerSig] deserialize from error:%v", err)
	}
	if this.PartnerPubKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[Withdraw] [PartnerPubKey] deserialize from error:%v", err)
	}
	return nil
}

func (this *WithDraw) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ChannelID)
	utils.EncodeAddress(sink, this.Participant)
	utils.EncodeAddress(sink, this.Partner)
	utils.EncodeVarUint(sink, this.TotalWithdraw)
	utils.EncodeBytes(sink, this.ParticipantSig)
	utils.EncodeBytes(sink, this.ParticipantPubKey)
	utils.EncodeBytes(sink, this.PartnerSig)
	utils.EncodeBytes(sink, this.PartnerPubKey)
}

func (this *WithDraw) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ChannelID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Participant, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Partner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.TotalWithdraw, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ParticipantSig, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.ParticipantPubKey, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.PartnerSig, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.PartnerPubKey, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}
