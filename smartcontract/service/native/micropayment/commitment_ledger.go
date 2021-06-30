package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type CommitmentTxLedger struct {
	FileHash  []byte
	PayFrom   common.Address
	PayTo     common.Address //==receiverAddr
	SliceID   uint64
	SlicePay  uint64
	Nonce     uint64
	SenderSig []byte
	PubKey    []byte
}

func (this *CommitmentTxLedger) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.FileHash); err != nil {
		return fmt.Errorf("[commitmentLedger] [FileHash:%v] serialize from error:%v", this.FileHash, err)
	}
	if err := utils.WriteAddress(w, this.PayFrom); err != nil {
		return fmt.Errorf("[commitmentLedger] [SenderAddr:%v] serialize from error:%v", this.PayFrom, err)
	}
	if err := utils.WriteAddress(w, this.PayTo); err != nil {
		return fmt.Errorf("[commitmentLedger] [ReceiverAddr:%v] serialize from error:%v", this.PayTo, err)
	}
	if err := utils.WriteVarUint(w, this.SliceID); err != nil {
		return fmt.Errorf("[commitmentLedger] [SliceID:%v] serialize from error:%v", this.SliceID, err)
	}
	if err := utils.WriteVarUint(w, this.SlicePay); err != nil {
		return fmt.Errorf("[commitmentLedger] [SlicePay:%v] serialize from error:%v", this.SlicePay, err)
	}
	if err := utils.WriteBytes(w, this.SenderSig); err != nil {
		return fmt.Errorf("[commitmentLedger] [SenderSig:%v] serialize from error:%v", this.SenderSig, err)
	}
	if err := utils.WriteBytes(w, this.PubKey); err != nil {
		return fmt.Errorf("[commitmentLedger] [PubKey:%v] serialize from error:%v", this.PubKey, err)
	}
	return nil
}

func (this *CommitmentTxLedger) Deserialize(r io.Reader) error {
	var err error
	if this.FileHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[commitmentLedger] [FileHash] deserialize from error:%v", err)
	}
	if this.PayFrom, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[commitmentLedger] [PayFrom] deserialize from error:%v", err)
	}
	if this.PayTo, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[commitmentLedger] [PayTo] deserialize from error:%v", err)
	}
	if this.SliceID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[commitmentLedger] [SliceId] deserialize from error:%v", err)
	}
	if this.SlicePay, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[commitmentLedger] [SlicePay] deserialize from error:%v", err)
	}
	if this.SenderSig, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[commitmentLedger] [SenderSig] deserialize from error:%v", err)
	}
	if this.PubKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[commitmentLedger] [PubKey] deserialize from error:%v", err)
	}
	return nil
}

func (this *CommitmentTxLedger) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.FileHash)
	utils.EncodeAddress(sink, this.PayFrom)
	utils.EncodeAddress(sink, this.PayTo)
	utils.EncodeVarUint(sink, this.SliceID)
	utils.EncodeVarUint(sink, this.SlicePay)
	utils.EncodeBytes(sink, this.SenderSig)
	utils.EncodeBytes(sink, this.PubKey)
}

func (this *CommitmentTxLedger) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.PayFrom, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.PayTo, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.SliceID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.SlicePay, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}

	this.SenderSig, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.PubKey, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}
