package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type UnlockDataInfo struct {
	LocksRoot    []byte
	LockedAmount uint64
}

func (this *UnlockDataInfo) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.LocksRoot); err != nil {
		return fmt.Errorf("[UnlockDataInfo] [LocksRoot:%v] serialize from error:%v", this.LocksRoot, err)
	}
	if err := utils.WriteVarUint(w, this.LockedAmount); err != nil {
		return fmt.Errorf("[UnlockDataInfo] [LockedAmount:%v] serialize from error:%v", this.LockedAmount, err)
	}
	return nil
}

func (this *UnlockDataInfo) Deserialize(r io.Reader) error {
	var err error
	if this.LocksRoot, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[UnlockDataInfo] [LocksRoot] deserialize from error:%v", err)
	}
	if this.LockedAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UnlockDataInfo] [LockedAmount] deserialize from error:%v", err)
	}
	return nil
}

func (this *UnlockDataInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.LocksRoot)
	utils.EncodeVarUint(sink, this.LockedAmount)
}

func (this *UnlockDataInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.LocksRoot, err = utils.DecodeBytes(source)
	if err != nil {
		return fmt.Errorf("[UnlockDataInfo] [LocksRoot] Deserialization from error:%v", err)
	}
	this.LockedAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("[UnlockDataInfo] [LockedAmount] Deserialization from error:%v", err)
	}
	return nil
}
