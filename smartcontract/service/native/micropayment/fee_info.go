package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type FeeInfo struct {
	WalletAddr   common.Address
	TokenAddr	 common.Address
	Flat         uint64
	Proportional uint64
	PublicKey    []byte
	Signature    []byte
}

func (this *FeeInfo) Serialize(w io.Writer) error {
	var err error
	err = utils.WriteAddress(w, this.WalletAddr)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [WalletAddr:%v] serialize from error:%v", this.WalletAddr, err)
	}
	err = utils.WriteAddress(w, this.TokenAddr)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [TokenAddr:%v] serialize from error:%v", this.TokenAddr, err)
	}
	err = utils.WriteVarUint(w, this.Flat)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [Flat:%v] serialize from error:%v", this.Flat, err)
	}
	err = utils.WriteVarUint(w, this.Proportional)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [Proportional:%v] serialize from error:%v", this.Proportional, err)
	}
	err = utils.WriteBytes(w, this.PublicKey)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [PublicKey:%v] serialize from error:%v", this.PublicKey, err)
	}
	err = utils.WriteBytes(w, this.Signature)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [Signature:%v] serialize from error:%v", this.Signature, err)
	}
	return nil
}

func (this *FeeInfo) Deserialize(r io.Reader) error {
	var err error
	this.WalletAddr, err = utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [WalletAddr] deserialize from error:%v", err)
	}
	this.TokenAddr, err = utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [TokenAddr] deserialize from error:%v", err)
	}
	this.Flat, err = utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [Flat] deserialize from error:%v", err)
	}
	this.Proportional, err = utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [Proportional] deserialize from error:%v", err)
	}
	this.PublicKey, err = utils.ReadBytes(r)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [PublicKey] deserialize from error:%v", err)
	}
	this.Signature, err = utils.ReadBytes(r)
	if err != nil {
		return fmt.Errorf("[FeeInfo] [Signature] deserialize from error:%v", err)
	}
	return nil
}

func (this *FeeInfo) Serialization(sink *common.ZeroCopySink) {
	this.WalletAddr.Serialization(sink)
	this.TokenAddr.Serialization(sink)
	utils.EncodeVarUint(sink, this.Flat)
	utils.EncodeVarUint(sink, this.Proportional)
	utils.EncodeBytes(sink, this.PublicKey)
	utils.EncodeBytes(sink, this.Signature)
}

func (this *FeeInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.TokenAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Flat, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Proportional, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.PublicKey, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Signature, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}
