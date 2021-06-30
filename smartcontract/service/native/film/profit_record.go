package film

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type ProfitRecord struct {
	TxHash      []byte
	FilmHash    []byte
	BuyAt       uint64
	PayAmount   uint64
	Payer       common.Address
	BlockHeight uint64
}

func (this *ProfitRecord) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.TxHash)
	utils.EncodeBytes(sink, this.FilmHash)
	utils.EncodeVarUint(sink, this.BuyAt)
	utils.EncodeVarUint(sink, this.PayAmount)
	utils.EncodeAddress(sink, this.Payer)
	utils.EncodeVarUint(sink, this.BlockHeight)
}

func (this *ProfitRecord) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.TxHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.FilmHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.BuyAt, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.PayAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Payer, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.BlockHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

func (this *ProfitRecord) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.TxHash); err != nil {
		return fmt.Errorf("[FilmInfo] [TxHash:%v] serialize from error:%v", this.TxHash, err)
	}
	if err := utils.WriteBytes(w, this.FilmHash); err != nil {
		return fmt.Errorf("[FilmInfo] [FilmHash:%v] serialize from error:%v", this.FilmHash, err)
	}
	if err := utils.WriteVarUint(w, this.BuyAt); err != nil {
		return fmt.Errorf("[FilmInfo] [CreatedAt:%v] serialize from error:%v", this.BuyAt, err)
	}
	if err := utils.WriteVarUint(w, this.PayAmount); err != nil {
		return fmt.Errorf("[FilmInfo] [PaidCount:%v] serialize from error:%v", this.PayAmount, err)
	}
	if err := utils.WriteAddress(w, this.Payer); err != nil {
		return fmt.Errorf("[FilmInfo] [TotalProfit:%v] serialize from error:%v", this.Payer, err)
	}
	if err := utils.WriteVarUint(w, this.BlockHeight); err != nil {
		return fmt.Errorf("[FilmInfo] [BlockHeight:%v] serialize from error:%v", this.BlockHeight, err)
	}
	return nil
}

func (this *ProfitRecord) Deserialize(r io.Reader) error {
	var err error
	if this.TxHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [TxHash] deserialize from error:%v", err)
	}
	if this.FilmHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [FilmHash] deserialize from error:%v", err)
	}
	if this.BuyAt, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [BuyAt] deserialize from error:%v", err)
	}
	if this.PayAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [PayAmount] deserialize from error:%v", err)
	}
	if this.Payer, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Payer] deserialize from error:%v", err)
	}
	if this.BlockHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [BlockHeight] deserialize from error:%v", err)
	}
	return nil
}
