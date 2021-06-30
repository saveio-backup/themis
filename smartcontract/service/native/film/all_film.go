package film

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type KeyList struct {
	Key  []byte
	Num  uint64
	List [][]byte
}

func (this *KeyList) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.Key)
	utils.EncodeVarUint(sink, this.Num)
	for _, v := range this.List {
		utils.EncodeBytes(sink, v)
	}
}

func (this *KeyList) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Key, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Num, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	data := make([][]byte, 0, this.Num)
	for i := uint64(0); i < this.Num; i++ {
		d, err := utils.DecodeBytes(source)
		if err != nil {
			return err
		}
		data = append(data, d)
	}
	this.List = data
	return nil
}

func (this *KeyList) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.Key); err != nil {
		return fmt.Errorf("[KeyList] [Key:%v] serialize from error:%v", this.Key, err)
	}
	if err := utils.WriteVarUint(w, this.Num); err != nil {
		return fmt.Errorf("[KeyList] [Num:%v] serialize from error:%v", this.Num, err)
	}
	for _, v := range this.List {
		if err := utils.WriteBytes(w, v); err != nil {
			return fmt.Errorf("[KeyList] [v:%v] serialize from error:%v", v, err)
		}
	}
	return nil
}

func (this *KeyList) Deserialize(r io.Reader) error {
	var err error
	if this.Key, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[AllFilms] [Num] deserialize from error:%v", err)
	}
	if this.Num, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[AllFilms] [MinReleasedYear] deserialize from error:%v", err)
	}
	data := make([][]byte, 0, this.Num)
	for i := uint64(0); i < this.Num; i++ {
		var d []byte
		if d, err = utils.ReadBytes(r); err != nil {
			return fmt.Errorf("[AllFilms] [key] deserialize from error:%v", err)
		}
		data = append(data, d)
	}
	this.List = data
	return nil
}

type FilmStats struct {
	Num             uint64
	MinReleasedYear uint64
	MaxReleasedYear uint64
	ReginNum        uint64
	Regions         [][]byte
}

func (this *FilmStats) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.Num)
	utils.EncodeVarUint(sink, this.MinReleasedYear)
	utils.EncodeVarUint(sink, this.MaxReleasedYear)
	utils.EncodeVarUint(sink, this.ReginNum)
	for _, reg := range this.Regions {
		utils.EncodeBytes(sink, reg)
	}
}

func (this *FilmStats) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Num, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.MinReleasedYear, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.MaxReleasedYear, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ReginNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	regs := make([][]byte, 0, this.ReginNum)
	for i := uint64(0); i < this.ReginNum; i++ {
		reg, err := utils.DecodeBytes(source)
		if err != nil {
			return err
		}
		regs = append(regs, reg)
	}
	this.Regions = regs
	return nil
}

func (this *FilmStats) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.Num); err != nil {
		return fmt.Errorf("[AllFilms] [Num:%v] serialize from error:%v", this.Num, err)
	}
	if err := utils.WriteVarUint(w, this.MinReleasedYear); err != nil {
		return fmt.Errorf("[AllFilms] [MinReleasedYear:%v] serialize from error:%v", this.MinReleasedYear, err)
	}
	if err := utils.WriteVarUint(w, this.MaxReleasedYear); err != nil {
		return fmt.Errorf("[AllFilms] [MaxReleasedYear:%v] serialize from error:%v", this.MaxReleasedYear, err)
	}
	if err := utils.WriteVarUint(w, this.ReginNum); err != nil {
		return fmt.Errorf("[AllFilms] [ReginNum:%v] serialize from error:%v", this.ReginNum, err)
	}
	for _, reg := range this.Regions {
		if err := utils.WriteBytes(w, reg); err != nil {
			return fmt.Errorf("[AllFilms] [reg:%v] serialize from error:%v", reg, err)
		}
	}
	return nil
}

func (this *FilmStats) Deserialize(r io.Reader) error {
	var err error
	if this.Num, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[AllFilms] [Num] deserialize from error:%v", err)
	}
	if this.MinReleasedYear, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[AllFilms] [MinReleasedYear] deserialize from error:%v", err)
	}
	if this.MaxReleasedYear, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[AllFilms] [MaxReleasedYear] deserialize from error:%v", err)
	}
	if this.ReginNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[AllFilms] [ReginNum] deserialize from error:%v", err)
	}
	regs := make([][]byte, 0, this.ReginNum)
	for i := uint64(0); i < this.ReginNum; i++ {
		var d []byte
		if d, err = utils.ReadBytes(r); err != nil {
			return fmt.Errorf("[AllFilms] [key] deserialize from error:%v", err)
		}
		regs = append(regs, d)
	}
	this.Regions = regs
	return nil
}
