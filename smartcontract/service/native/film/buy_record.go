package film

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type BuyRecord struct {
	TxHash      []byte
	FilmHash    []byte
	FilmOwner   common.Address
	BuyAt       uint64
	Cost        uint64
	BlockHeight uint64
}

func (this *BuyRecord) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.TxHash)
	utils.EncodeBytes(sink, this.FilmHash)
	utils.EncodeAddress(sink, this.FilmOwner)
	utils.EncodeVarUint(sink, this.BuyAt)
	utils.EncodeVarUint(sink, this.Cost)
	utils.EncodeVarUint(sink, this.BlockHeight)
}

func (this *BuyRecord) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.TxHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.FilmHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.FilmOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.BuyAt, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Cost, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.BlockHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

func (this *BuyRecord) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.TxHash); err != nil {
		return fmt.Errorf("[FilmInfo] [TxHash:%v] serialize from error:%v", this.TxHash, err)
	}
	if err := utils.WriteBytes(w, this.FilmHash); err != nil {
		return fmt.Errorf("[FilmInfo] [FilmHash:%v] serialize from error:%v", this.FilmHash, err)
	}
	if err := utils.WriteAddress(w, this.FilmOwner); err != nil {
		return fmt.Errorf("[FilmInfo] [FilmOwner:%v] serialize from error:%v", this.FilmOwner, err)
	}
	if err := utils.WriteVarUint(w, this.BuyAt); err != nil {
		return fmt.Errorf("[FilmInfo] [BuyAt:%v] serialize from error:%v", this.BuyAt, err)
	}
	if err := utils.WriteVarUint(w, this.Cost); err != nil {
		return fmt.Errorf("[FilmInfo] [Cost:%v] serialize from error:%v", this.Cost, err)
	}
	if err := utils.WriteVarUint(w, this.BlockHeight); err != nil {
		return fmt.Errorf("[FilmInfo] [BlockHeight:%v] serialize from error:%v", this.BlockHeight, err)
	}
	return nil
}

func (this *BuyRecord) Deserialize(r io.Reader) error {
	var err error
	if this.TxHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [TxHash] deserialize from error:%v", err)
	}
	if this.FilmHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [FilmHash] deserialize from error:%v", err)
	}
	if this.FilmOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[FilmInfo] [FilmOwner] deserialize from error:%v", err)
	}
	if this.BuyAt, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [BuyAt] deserialize from error:%v", err)
	}
	if this.Cost, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Cost] deserialize from error:%v", err)
	}
	if this.BlockHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [BlockHeight] deserialize from error:%v", err)
	}
	return nil
}
