package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type UpdateNonCloseBalanceProof struct {
	ChanID              uint64
	CloseParticipant    common.Address
	NonCloseParticipant common.Address
	BalanceHash         []byte
	Nonce               uint64
	AdditionalHash      []byte
	CloseSignature      []byte
	NonCloseSignature   []byte
	ClosePubKey         []byte
	NonClosePubKey      []byte
}

func (this *UpdateNonCloseBalanceProof) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChanID); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [chaid:%v] serialize from error:%v", this.ChanID, err)
	}
	if err := utils.WriteAddress(w, this.CloseParticipant); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [closeParticipant:%v] serialize from error:%v", this.CloseParticipant, err)
	}
	if err := utils.WriteAddress(w, this.NonCloseParticipant); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [nonCloseParticipant:%v] serialize from error:%v", this.NonCloseParticipant, err)
	}
	if err := utils.WriteBytes(w, this.BalanceHash); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [balanceHash:%v] serialize from error:%v", this.BalanceHash, err)
	}
	if err := utils.WriteVarUint(w, this.Nonce); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [Nonce:%v] serialize from error:%v", this.Nonce, err)
	}
	if err := utils.WriteBytes(w, this.AdditionalHash); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [additionalHash:%v] serialize from error:%v", this.AdditionalHash, err)
	}
	if err := utils.WriteBytes(w, this.CloseSignature); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [CloseSignature:%v] serialize from error:%v", this.CloseSignature, err)
	}
	if err := utils.WriteBytes(w, this.NonCloseSignature); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [NonCloseSignature:%v] serialize from error:%v", this.NonCloseSignature, err)
	}
	if err := utils.WriteBytes(w, this.ClosePubKey); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [ClosePubKey:%v] serialize from error:%v", this.ClosePubKey, err)
	}
	if err := utils.WriteBytes(w, this.NonClosePubKey); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [NonClosePubKey:%v] serialize from error:%v", this.NonClosePubKey, err)
	}
	return nil
}

func (this *UpdateNonCloseBalanceProof) Deserialize(r io.Reader) error {
	var err error
	if this.ChanID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [ChanID] deserialize from error:%v", err)
	}
	if this.CloseParticipant, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [CloseParticipant] deserialize from error:%v", err)
	}
	if this.NonCloseParticipant, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [NonCloseParticipant] deserialize from error:%v", err)
	}
	if this.BalanceHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [BalanceHash] deserialize from error:%v", err)
	}
	if this.Nonce, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [Nonce] deserialize from error:%v", err)
	}
	if this.AdditionalHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [AdditionalHash] deserialize from error:%v", err)
	}
	if this.CloseSignature, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [CloseSignature] deserialize from error:%v", err)
	}
	if this.NonCloseSignature, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [NonCloseSignature] deserialize from error:%v", err)
	}
	if this.ClosePubKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [ClosePubKey] deserialize from error:%v", err)
	}
	if this.NonClosePubKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [NonClosePubKey] deserialize from error:%v", err)
	}
	return nil
}

func (this *UpdateNonCloseBalanceProof) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ChanID)
	utils.EncodeAddress(sink, this.CloseParticipant)
	utils.EncodeAddress(sink, this.NonCloseParticipant)
	utils.EncodeBytes(sink, this.BalanceHash)
	utils.EncodeVarUint(sink, this.Nonce)
	utils.EncodeBytes(sink, this.AdditionalHash)
	utils.EncodeBytes(sink, this.CloseSignature)
	utils.EncodeBytes(sink, this.NonCloseSignature)
	utils.EncodeBytes(sink, this.ClosePubKey)
	utils.EncodeBytes(sink, this.NonClosePubKey)
}

func (this *UpdateNonCloseBalanceProof) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ChanID, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [ChanID] Deserialization from error:%v", err)
	}
	this.CloseParticipant, err = utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [CloseParticipant] Deserialization from error:%v", err)
	}
	this.NonCloseParticipant, err = utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [NonCloseParticipant] Deserialization from error:%v", err)
	}
	this.BalanceHash, err = utils.DecodeBytes(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [BalanceHash] Deserialization from error:%v", err)
	}
	this.Nonce, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [Nonce] Deserialization from error:%v", err)
	}
	this.AdditionalHash, err = utils.DecodeBytes(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [AdditionalHash] Deserialization from error:%v", err)
	}
	this.CloseSignature, err = utils.DecodeBytes(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [CloseSignature] Deserialization from error:%v", err)
	}
	this.NonCloseSignature, err = utils.DecodeBytes(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [NonCloseSignature] Deserialization from error:%v", err)
	}
	this.ClosePubKey, err = utils.DecodeBytes(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [ClosePubKey] Deserialization from error:%v", err)
	}
	this.NonClosePubKey, err = utils.DecodeBytes(source)
	if err != nil {
		return fmt.Errorf("[MPupdateNoncloseBP] [NonClosePubKey] Deserialization from error:%v", err)
	}
	return nil
}
