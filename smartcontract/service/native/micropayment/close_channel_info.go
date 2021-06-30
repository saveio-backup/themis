package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type CloseChannelInfo struct {
	ChannelID          uint64
	ParticipantAddress common.Address
	PartnerAddress     common.Address
	BalanceHash        []byte
	Nonce              uint64
	AdditionalHash     []byte
	PartnerSignature   []byte
	PartnerPubKey      []byte
}

func (this *CloseChannelInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChannelID); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [ChannelID:%v] serialize from error:%v", this.ChannelID, err)
	}
	if err := utils.WriteAddress(w, this.ParticipantAddress); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [ParticipantAddress:%v] serialize from error:%v", this.ParticipantAddress, err)
	}
	if err := utils.WriteAddress(w, this.PartnerAddress); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [PartnerAddress:%v] serialize from error:%v", this.PartnerAddress, err)
	}
	if err := utils.WriteBytes(w, this.BalanceHash); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [BalanceHash:%v] serialize from error:%v", this.BalanceHash, err)
	}
	if err := utils.WriteVarUint(w, this.Nonce); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [Nonce:%v] serialize from error:%v", this.Nonce, err)
	}
	if err := utils.WriteBytes(w, this.AdditionalHash); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [AdditionalHash:%v] serialize from error:%v", this.AdditionalHash, err)
	}
	if err := utils.WriteBytes(w, this.PartnerSignature); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [PartnerSignature:%v] serialize from error:%v", this.PartnerSignature, err)
	}
	if err := utils.WriteBytes(w, this.PartnerPubKey); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [PartnerPubKey:%v] serialize from error:%v", this.PartnerPubKey, err)
	}
	return nil
}

func (this *CloseChannelInfo) Deserialize(r io.Reader) error {
	var err error
	if this.ChannelID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [ChannelID] deserialize from error:%v", err)
	}
	if this.ParticipantAddress, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [ParticipantAddress] deserialize from error:%v", err)
	}
	if this.PartnerAddress, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [PartnerAddress] deserialize from error:%v", err)
	}
	if this.BalanceHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [BalanceHash] deserialize from error:%v", err)
	}
	if this.Nonce, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [Nonce] deserialize from error:%v", err)
	}
	if this.AdditionalHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [AdditionalHash] deserialize from error:%v", err)
	}
	if this.PartnerSignature, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [PartnerSignature] deserialize from error:%v", err)
	}
	if this.PartnerPubKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[CloseChannelInfo] [PartnerPubKey] deserialize from error:%v", err)
	}
	return nil
}

func (this *CloseChannelInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ChannelID)
	utils.EncodeAddress(sink, this.ParticipantAddress)
	utils.EncodeAddress(sink, this.PartnerAddress)
	utils.EncodeBytes(sink, this.BalanceHash)
	utils.EncodeVarUint(sink, this.Nonce)
	utils.EncodeBytes(sink, this.AdditionalHash)
	utils.EncodeBytes(sink, this.PartnerSignature)
	utils.EncodeBytes(sink, this.PartnerPubKey)
}

func (this *CloseChannelInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ChannelID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ParticipantAddress, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.PartnerAddress, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.BalanceHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Nonce, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.AdditionalHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.PartnerSignature, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.PartnerPubKey, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}
