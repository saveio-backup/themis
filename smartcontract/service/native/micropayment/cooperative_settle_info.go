package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type CooperativeSettleInfo struct {
	ChannelID             uint64
	Participant1Address   common.Address
	Participant1Balance   uint64
	Participant2Address   common.Address
	Participant2Balance   uint64
	Participant1Signature []byte
	Participant1PubKey    []byte
	Participant2Signature []byte
	Participant2PubKey    []byte
}

func (this *CooperativeSettleInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChannelID); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [ChannelID:%v] serialize from error:%v", this.ChannelID, err)
	}
	if err := utils.WriteAddress(w, this.Participant1Address); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant1Address:%v] serialize from error:%v", this.Participant1Address, err)
	}
	if err := utils.WriteVarUint(w, this.Participant1Balance); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant1Balance:%v] serialize from error:%v", this.Participant1Balance, err)
	}
	if err := utils.WriteAddress(w, this.Participant2Address); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant2Address:%v] serialize from error:%v", this.Participant2Address, err)
	}
	if err := utils.WriteVarUint(w, this.Participant2Balance); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant2Balance:%v] serialize from error:%v", this.Participant2Balance, err)
	}
	if err := utils.WriteBytes(w, this.Participant1Signature); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant1Signature:%v] serialize from error:%v", this.Participant1Signature, err)
	}
	if err := utils.WriteBytes(w, this.Participant1PubKey); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant1PubKey:%v] serialize from error:%v", this.Participant1Signature, err)
	}
	if err := utils.WriteBytes(w, this.Participant2Signature); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant2Signature:%v] serialize from error:%v", this.Participant2Signature, err)
	}
	if err := utils.WriteBytes(w, this.Participant2PubKey); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant2PubKey:%v] serialize from error:%v", this.Participant1Signature, err)
	}
	return nil
}

func (this *CooperativeSettleInfo) Deserialize(r io.Reader) error {
	var err error
	if this.ChannelID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [ChannelID] deserialize from error:%v", err)
	}
	if this.Participant1Address, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant1Address] deserialize from error:%v", err)
	}
	if this.Participant1Balance, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant1Balance] deserialize from error:%v", err)
	}
	if this.Participant2Address, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant2Address] deserialize from error:%v", err)
	}

	if this.Participant2Balance, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant2Balance] deserialize from error:%v", err)
	}

	if this.Participant1Signature, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant1Signature] deserialize from error:%v", err)
	}
	if this.Participant1PubKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant1PubKey] deserialize from error:%v", err)
	}

	if this.Participant2Signature, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant2Signature] deserialize from error:%v", err)
	}
	if this.Participant2PubKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[CooperativeSettleInfo] [Participant2PubKey] deserialize from error:%v", err)
	}
	return nil
}

func (this *CooperativeSettleInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ChannelID)
	utils.EncodeAddress(sink, this.Participant1Address)
	utils.EncodeVarUint(sink, this.Participant1Balance)
	utils.EncodeAddress(sink, this.Participant2Address)
	utils.EncodeVarUint(sink, this.Participant2Balance)
	utils.EncodeBytes(sink, this.Participant1Signature)
	utils.EncodeBytes(sink, this.Participant1PubKey)
	utils.EncodeBytes(sink, this.Participant2Signature)
	utils.EncodeBytes(sink, this.Participant2PubKey)
}

func (this *CooperativeSettleInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ChannelID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Participant1Address, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Participant1Balance, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Participant2Address, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Participant2Balance, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Participant1Signature, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Participant1PubKey, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Participant2Signature, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Participant2PubKey, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}
