package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type GetChannelId struct {
	Participant1WalletAddr common.Address
	Participant2WalletAddr common.Address
}

func (this *GetChannelId) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Participant1WalletAddr); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant1WalletAddr:%v] serialize from error:%v",
			this.Participant1WalletAddr, err)
	}
	if err := utils.WriteAddress(w, this.Participant2WalletAddr); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant2WalletAddr:%v] serialize from error:%v",
			this.Participant2WalletAddr, err)
	}
	return nil
}

func (this *GetChannelId) Deserialize(r io.Reader) error {
	var err error
	if this.Participant1WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant1WalletAddr] deserialize from error:%v", err)
	}
	if this.Participant2WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[openChannelInfo] [Participant2WalletAddr] deserialize from error:%v", err)
	}
	return nil
}

func (this *GetChannelId) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Participant1WalletAddr)
	utils.EncodeAddress(sink, this.Participant2WalletAddr)
}

func (this *GetChannelId) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Participant1WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Participant2WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	return nil
}
