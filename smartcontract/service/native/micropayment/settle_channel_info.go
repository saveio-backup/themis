package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type SettleChannelInfo struct {
	ChanID              uint64
	Participant1        common.Address
	P1TransferredAmount uint64
	P1LockedAmount      uint64
	P1LocksRoot         []byte
	Participant2        common.Address
	P2TransferredAmount uint64
	P2LockedAmount      uint64
	P2LocksRoot         []byte
}

func (this *SettleChannelInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChanID); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [chaid:%v] serialize from error:%v", this.ChanID, err)
	}
	if err := utils.WriteAddress(w, this.Participant1); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [Participant1:%v] serialize from error:%v", this.Participant1, err)
	}
	if err := utils.WriteVarUint(w, this.P1TransferredAmount); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P1TransferredAmount:%v] serialize from error:%v", this.P1TransferredAmount, err)
	}
	if err := utils.WriteVarUint(w, this.P1LockedAmount); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P1LockedAmount:%v] serialize from error:%v", this.P1LockedAmount, err)
	}
	if err := utils.WriteBytes(w, this.P1LocksRoot); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P1LocksRoot:%v] serialize from error:%v", this.P1LocksRoot, err)
	}
	if err := utils.WriteAddress(w, this.Participant2); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [Participant2:%v] serialize from error:%v", this.Participant2, err)
	}
	if err := utils.WriteVarUint(w, this.P2TransferredAmount); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P2TransferredAmount:%v] serialize from error:%v", this.P2TransferredAmount, err)
	}
	if err := utils.WriteVarUint(w, this.P2LockedAmount); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P2LockedAmount:%v] serialize from error:%v", this.P2LockedAmount, err)
	}
	if err := utils.WriteBytes(w, this.P2LocksRoot); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P2LocksRoot:%v] serialize from error:%v", this.P2LocksRoot, err)
	}
	return nil
}

func (this *SettleChannelInfo) Deserialize(r io.Reader) error {
	var err error
	if this.ChanID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [ChanID] deserialize from error:%v", err)
	}
	if this.Participant1, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [Participant1] deserialize from error:%v", err)
	}
	if this.P1TransferredAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P1TransferredAmount] deserialize from error:%v", err)
	}
	if this.P1LockedAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P1LockedAmount] deserialize from error:%v", err)
	}
	if this.P1LocksRoot, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P1LocksRoot] deserialize from error:%v", err)
	}
	if this.Participant2, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [Participant2] deserialize from error:%v", err)
	}
	if this.P2TransferredAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P2TransferredAmount] deserialize from error:%v", err)
	}
	if this.P2LockedAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P2LockedAmount] deserialize from error:%v", err)
	}
	if this.P2LocksRoot, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P2LocksRoot] deserialize from error:%v", err)
	}
	return nil
}

func (this *SettleChannelInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ChanID)
	utils.EncodeAddress(sink, this.Participant1)
	utils.EncodeVarUint(sink, this.P1TransferredAmount)
	utils.EncodeVarUint(sink, this.P1LockedAmount)
	utils.EncodeBytes(sink, this.P1LocksRoot)
	utils.EncodeAddress(sink, this.Participant2)
	utils.EncodeVarUint(sink, this.P2TransferredAmount)
	utils.EncodeVarUint(sink, this.P2LockedAmount)
	utils.EncodeBytes(sink, this.P2LocksRoot)
}

func (this *SettleChannelInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ChanID, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [ChanID] Deserialization from error:%v", err)
	}
	this.Participant1, err = utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [Participant1] Deserialization from error:%v", err)
	}
	this.P1TransferredAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P1TransferredAmount] Deserialization from error:%v", err)
	}
	this.P1LockedAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P1LockedAmount] Deserialization from error:%v", err)
	}
	this.P1LocksRoot, err = utils.DecodeBytes(source)
	if err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P1LocksRoot] Deserialization from error:%v", err)
	}
	this.Participant2, err = utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [Participant2] Deserialization from error:%v", err)
	}
	this.P2TransferredAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P2TransferredAmount] Deserialization from error:%v", err)
	}
	this.P2LockedAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P2LockedAmount] Deserialization from error:%v", err)
	}
	this.P2LocksRoot, err = utils.DecodeBytes(source)
	if err != nil {
		return fmt.Errorf("[MPsettleChannelInfo] [P2LocksRoot] Deserialization from error:%v", err)
	}
	return nil
}
