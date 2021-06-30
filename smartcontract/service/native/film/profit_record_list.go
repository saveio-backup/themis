package film

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type ProfitRecordList struct {
	Owner    common.Address
	Num      uint64
	TxHashes [][]byte
}

func (this *ProfitRecordList) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Owner)
	utils.EncodeVarUint(sink, this.Num)
	for _, id := range this.TxHashes {
		utils.EncodeBytes(sink, id)
	}
}

func (this *ProfitRecordList) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Owner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Num, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	ids := make([][]byte, 0, this.Num)
	for i := uint64(0); i < this.Num; i++ {
		id, err := utils.DecodeBytes(source)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	this.TxHashes = ids
	return nil
}

func (this *ProfitRecordList) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("[ProfitRecordList] [Owner:%v] serialize from error:%v", this.Owner, err)
	}

	if err := utils.WriteVarUint(w, this.Num); err != nil {
		return fmt.Errorf("[ProfitRecordList] [Num:%v] serialize from error:%v", this.Num, err)
	}
	for _, key := range this.TxHashes {
		if err := utils.WriteBytes(w, key); err != nil {
			return fmt.Errorf("[ProfitRecordList] [TxHashes:%v] serialize from error:%v", key, err)
		}
	}
	return nil
}

func (this *ProfitRecordList) Deserialize(r io.Reader) error {
	var err error
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[ProfitRecordList] [Owner] deserialize from error:%v", err)
	}
	if this.Num, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[ProfitRecordList] [Num] deserialize from error:%v", err)
	}
	data := make([][]byte, 0, this.Num)
	for i := uint64(0); i < this.Num; i++ {
		var d []byte
		if d, err = utils.ReadBytes(r); err != nil {
			return fmt.Errorf("[ProfitRecordList] [key] deserialize from error:%v", err)
		}
		data = append(data, d)
	}
	this.TxHashes = data
	return nil
}
