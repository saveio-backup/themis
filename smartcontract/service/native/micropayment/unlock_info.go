package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type UnlockInfo struct {
	ChannelID          uint64
	ParticipantAddress common.Address
	PartnerAddress     common.Address
	MerkleTreeLeaves   []byte
}

func (this *UnlockInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ChannelID); err != nil {
		return fmt.Errorf("[UnlockInfo] [ChannelID:%v] serialize from error:%v", this.ChannelID, err)
	}
	if err := utils.WriteAddress(w, this.ParticipantAddress); err != nil {
		return fmt.Errorf("[UnlockInfo] [ParticipantAddress:%v] serialize from error:%v", this.ParticipantAddress, err)
	}
	if err := utils.WriteAddress(w, this.PartnerAddress); err != nil {
		return fmt.Errorf("[UnlockInfo] [PartnerAddress:%v] serialize from error:%v", this.PartnerAddress, err)
	}
	if err := utils.WriteBytes(w, this.MerkleTreeLeaves); err != nil {
		return fmt.Errorf("[UnlcokInfo] [MerkleTreeLeaves:%v] serialize from error:%v", this.MerkleTreeLeaves, err)
	}
	return nil
}

func (this *UnlockInfo) Deserialize(r io.Reader) error {
	var err error
	if this.ChannelID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UnlockInfo] [ChannelID] deserialize from error:%v", err)
	}
	if this.ParticipantAddress, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[UnlockInfo] [ParticipantAddress] deserialize from error:%v", err)
	}
	if this.PartnerAddress, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[UnlockInfo] [PartnerAddress] deserialize from error:%v", err)
	}
	if this.MerkleTreeLeaves, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[UnlockInfo] [MerkleTreeLeaves] deserialize from error:%v", err)
	}
	return nil
}

func (this *UnlockInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ChannelID)
	utils.EncodeAddress(sink, this.ParticipantAddress)
	utils.EncodeAddress(sink, this.PartnerAddress)
	utils.EncodeBytes(sink, this.MerkleTreeLeaves)
}

func (this *UnlockInfo) Deserialization(source *common.ZeroCopySource) error {
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
	this.MerkleTreeLeaves, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}
