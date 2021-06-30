package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

//ChannelState
type SetTotalDepositInfo struct {
	ChannelID             uint64
	ParticipantWalletAddr common.Address
	PartnerWalletAddr     common.Address
	SetTotalDeposit       uint64
}

func (this *SetTotalDepositInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChannelID); err != nil {
		return fmt.Errorf("[SetTotalDepositInfo] [ChannelID:%v] serialize from error:%v", this.ChannelID, err)
	}
	if err := utils.WriteAddress(w, this.ParticipantWalletAddr); err != nil {
		return fmt.Errorf("[SetTotalDepositInfo] [Participant1WalletAddr:%v] serialize from error:%v",
			this.ParticipantWalletAddr, err)
	}
	if err := utils.WriteAddress(w, this.PartnerWalletAddr); err != nil {
		return fmt.Errorf("[SetTotalDepositInfo] [PartnerWalletAddr:%v] serialize from error:%v",
			this.PartnerWalletAddr, err)
	}
	if err := utils.WriteVarUint(w, this.SetTotalDeposit); err != nil {
		return fmt.Errorf("[SetTotalDepositInfo] [SetTotalDeposit:%v] serialize from error:%v", this.SetTotalDeposit, err)
	}
	return nil
}

func (this *SetTotalDepositInfo) Deserialize(r io.Reader) error {
	var err error
	if this.ChannelID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SetTotalDepositInfo] [ChannelID] deserialize from error:%v", err)
	}
	if this.ParticipantWalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[SetTotalDepositInfo] [ParticipantWalletAddr] deserialize from error:%v", err)
	}
	if this.PartnerWalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[SetTotalDepositInfo] [PartnerWalletAddr] deserialize from error:%v", err)
	}
	if this.SetTotalDeposit, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SetTotalDepositInfo] [SetTotalDeposit] deserialize from error:%v", err)
	}
	return nil
}

func (this *SetTotalDepositInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ChannelID)
	utils.EncodeAddress(sink, this.ParticipantWalletAddr)
	utils.EncodeAddress(sink, this.PartnerWalletAddr)
	utils.EncodeVarUint(sink, this.SetTotalDeposit)
}

func (this *SetTotalDepositInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ChannelID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ParticipantWalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.PartnerWalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.SetTotalDeposit, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

// TransferInfo
type TransferInfo struct {
	PaymentId uint64
	From      common.Address
	To        common.Address
	Amount    uint64
}

func (this *TransferInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.PaymentId); err != nil {
		return fmt.Errorf("[TransferInfo] [PaymentId:%v] serialize from error:%v", this.PaymentId, err)
	}
	if err := utils.WriteAddress(w, this.From); err != nil {
		return fmt.Errorf("[TransferInfo] [From:%v] serialize from error:%v",
			this.From, err)
	}
	if err := utils.WriteAddress(w, this.To); err != nil {
		return fmt.Errorf("[TransferInfo] [To:%v] serialize from error:%v",
			this.To, err)
	}
	if err := utils.WriteVarUint(w, this.Amount); err != nil {
		return fmt.Errorf("[TransferInfo] [Amount:%v] serialize from error:%v", this.Amount, err)
	}
	return nil
}

func (this *TransferInfo) Deserialize(r io.Reader) error {
	var err error
	if this.PaymentId, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[TransferInfo] [PaymentId] deserialize from error:%v", err)
	}
	if this.From, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[TransferInfo] [From] deserialize from error:%v", err)
	}
	if this.To, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[TransferInfo] [To] deserialize from error:%v", err)
	}
	if this.Amount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[TransferInfo] [Amount] deserialize from error:%v", err)
	}
	return nil
}

func (this *TransferInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.PaymentId)
	utils.EncodeAddress(sink, this.From)
	utils.EncodeAddress(sink, this.To)
	utils.EncodeVarUint(sink, this.Amount)
}

func (this *TransferInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.PaymentId, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.From, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.To, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Amount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}
