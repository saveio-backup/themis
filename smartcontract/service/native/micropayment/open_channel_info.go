package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type OpenChannelInfo struct {
	Participant1WalletAddr common.Address
	Participant1PubKey     []byte
	Participant2WalletAddr common.Address
	SettleBlockHeight      uint64
}

func (this *OpenChannelInfo) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Participant1WalletAddr); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant1WalletAddr:%v] serialize from error:%v",
			this.Participant1WalletAddr, err)
	}
	if err := utils.WriteBytes(w, this.Participant1PubKey); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant1PubKey:%v] serialize from error:%v",
			this.Participant1PubKey, err)
	}
	if err := utils.WriteAddress(w, this.Participant2WalletAddr); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant2WalletAddr:%v] serialize from error:%v",
			this.Participant2WalletAddr, err)
	}
	if err := utils.WriteVarUint(w, this.SettleBlockHeight); err != nil {
		return fmt.Errorf("[openChannelInfo] [SettleBlockHeight:%v] serialize from error:%v",
			this.SettleBlockHeight, err)
	}
	return nil
}

func (this *OpenChannelInfo) Deserialize(r io.Reader) error {
	var err error
	if this.Participant1WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant1WalletAddr] deserialize from error:%v", err)
	}
	if this.Participant1PubKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant1PubKey] deserialize from error:%v", err)
	}
	if this.Participant2WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant2WalletAddr] deserialize from error:%v", err)
	}
	if this.SettleBlockHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[openChannelInfo] [SettleBlockHeight] deserialize from error:%v", err)
	}
	return nil
}

func (this *OpenChannelInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Participant1WalletAddr)
	utils.EncodeBytes(sink, this.Participant1PubKey)
	utils.EncodeAddress(sink, this.Participant2WalletAddr)
	utils.EncodeVarUint(sink, this.SettleBlockHeight)
}

func (this *OpenChannelInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Participant1WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Participant1PubKey, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Participant2WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.SettleBlockHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}
