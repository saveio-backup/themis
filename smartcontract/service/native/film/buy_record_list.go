package film

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type BuyRecordList struct {
	Owner     common.Address
	RecordNum uint64
	TxHashes  [][]byte
}

func (this *BuyRecordList) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Owner)
	utils.EncodeVarUint(sink, this.RecordNum)
	for _, id := range this.TxHashes {
		utils.EncodeBytes(sink, id)
	}
}

func (this *BuyRecordList) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Owner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.RecordNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	ids := make([][]byte, 0, this.RecordNum)
	for i := uint64(0); i < this.RecordNum; i++ {
		id, err := utils.DecodeBytes(source)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	this.TxHashes = ids
	return nil
}

func (this *BuyRecordList) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("[BuyRecordList] [Owner:%v] serialize from error:%v", this.Owner, err)
	}
	if err := utils.WriteVarUint(w, this.RecordNum); err != nil {
		return fmt.Errorf("[BuyRecordList] [RecordNum:%v] serialize from error:%v", this.RecordNum, err)
	}
	for _, key := range this.TxHashes {
		if err := utils.WriteBytes(w, key); err != nil {
			return fmt.Errorf("[BuyRecordList] [TxHashes:%v] serialize from error:%v", key, err)
		}
	}
	return nil
}

func (this *BuyRecordList) Deserialize(r io.Reader) error {
	var err error
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[BuyRecordList] [Owner] deserialize from error:%v", err)
	}
	if this.RecordNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[BuyRecordList] [RecordNum] deserialize from error:%v", err)
	}
	data := make([][]byte, 0, this.RecordNum)
	for i := uint64(0); i < this.RecordNum; i++ {
		var d []byte
		if d, err = utils.ReadBytes(r); err != nil {
			return fmt.Errorf("[BuyRecordList] [key] deserialize from error:%v", err)
		}
		data = append(data, d)
	}
	this.TxHashes = data
	return nil
}
