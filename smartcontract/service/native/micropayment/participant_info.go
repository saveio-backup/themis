package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type Participant struct {
	WalletAddr     common.Address
	Deposit        uint64
	WithDrawAmount uint64
	IP             []byte
	Port           []byte
	Balance        uint64
	BalanceHash    []byte
	Nonce          uint64
	IsCloser       bool
	LocksRoot      []byte
	LockedAmount   uint64
}

func (this *Participant) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.WalletAddr); err != nil {
		return fmt.Errorf("[Participant] [WalletAddr:%v] serialize from error:%v", this.WalletAddr, err)
	}
	if err := utils.WriteVarUint(w, this.Deposit); err != nil {
		return fmt.Errorf("[Participant] [total_deposit:%v] serialize from error:%v", this.Deposit, err)
	}
	if err := utils.WriteVarUint(w, this.WithDrawAmount); err != nil {
		return fmt.Errorf("[Participant] [withdraw_amount:%v] serialize from error:%v", this.WithDrawAmount, err)
	}
	if err := utils.WriteBytes(w, this.IP); err != nil {
		return fmt.Errorf("[Participant] [IP:%v] serialize from error:%v", this.IP, err)
	}
	if err := utils.WriteBytes(w, this.Port); err != nil {
		return fmt.Errorf("[Participant] [Port:%v] serialize from error:%v", this.Port, err)
	}
	if err := utils.WriteVarUint(w, this.Balance); err != nil {
		return fmt.Errorf("[Participant] [Balance:%v] serialize from error:%v", this.Balance, err)
	}
	if err := utils.WriteBytes(w, this.BalanceHash); err != nil {
		return fmt.Errorf("[Participant] [BalanceHash:%v] serialize from error:%v", this.BalanceHash, err)
	}
	if err := utils.WriteVarUint(w, this.Nonce); err != nil {
		return fmt.Errorf("[Participant] [Nonce:%v] serialize from error:%v", this.Nonce, err)
	}
	if err := utils.WriteBool(w, this.IsCloser); err != nil {
		return fmt.Errorf("[Participant] [IsCloser:%v] serialize from error:%v", this.IsCloser, err)
	}
	if err := utils.WriteBytes(w, this.LocksRoot); err != nil {
		return fmt.Errorf("[Participant] [LocksRoot:%v] serialize from error:%v", this.BalanceHash, err)
	}
	if err := utils.WriteVarUint(w, this.LockedAmount); err != nil {
		return fmt.Errorf("[Participant] [LockedAmount:%v] serialize from error:%v", this.Deposit, err)
	}

	return nil
}

func (this *Participant) Deserialize(r io.Reader) error {
	var err error
	if this.WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[Participant] [WalletAddr] deserialize from error:%v", err)
	}
	if this.Deposit, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Participant] [total_deposit] deserialize from error:%v", err)
	}
	if this.WithDrawAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Participant] [withdraw_amount] deserialize from error:%v", err)
	}
	if this.IP, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[Participant] [IP] deserialize from error:%v", err)
	}
	if this.Port, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[Participant] [Port] deserialize from error:%v", err)
	}
	if this.Balance, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Participant] [Balance] deserialize from error:%v", err)
	}
	if this.BalanceHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[Participant] [BalanceHash] deserialize from error:%v", err)
	}
	if this.Nonce, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Participant] [Nonce] deserialize from error:%v", err)
	}
	if this.IsCloser, err = utils.ReadBool(r); err != nil {
		return fmt.Errorf("[Participant] [IsCloser] deserialize from error:%v", err)
	}
	if this.LocksRoot, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[Participant] [LocksRoot] deserialize from error:%v", err)
	}
	if this.LockedAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[Participant] [LockedAmount] deserialize from error:%v", err)
	}
	return nil
}

func (this *Participant) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.WalletAddr)
	utils.EncodeVarUint(sink, this.Deposit)
	utils.EncodeVarUint(sink, this.WithDrawAmount)
	utils.EncodeBytes(sink, this.IP)
	utils.EncodeBytes(sink, this.Port)
	utils.EncodeVarUint(sink, this.Balance)
	utils.EncodeBytes(sink, this.BalanceHash)
	utils.EncodeVarUint(sink, this.Nonce)
	utils.EncodeBool(sink, this.IsCloser)
	utils.EncodeBytes(sink, this.LocksRoot)
	utils.EncodeVarUint(sink, this.LockedAmount)
}

func (this *Participant) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Deposit, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.WithDrawAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.IP, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Port, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Balance, err = utils.DecodeVarUint(source)
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
	this.IsCloser, err = utils.DecodeBool(source)
	if err != nil {
		return err
	}
	this.LocksRoot, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.LockedAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}
