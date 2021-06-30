package film

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type UserFilmList struct {
	Owner      common.Address
	Num        uint64
	FilmHashes [][]byte
}

func (this *UserFilmList) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Owner)
	utils.EncodeVarUint(sink, this.Num)
	for _, hash := range this.FilmHashes {
		utils.EncodeBytes(sink, hash)
	}
}

func (this *UserFilmList) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Owner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Num, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	hashes := make([][]byte, 0, this.Num)
	for i := uint64(0); i < this.Num; i++ {
		hash, err := utils.DecodeBytes(source)
		if err != nil {
			return err
		}
		hashes = append(hashes, hash)
	}
	this.FilmHashes = hashes
	return nil
}

func (this *UserFilmList) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("[UserFilmList] [Owner:%v] serialize from error:%v", this.Owner, err)
	}

	if err := utils.WriteVarUint(w, this.Num); err != nil {
		return fmt.Errorf("[UserFilmList] [Num:%v] serialize from error:%v", this.Num, err)
	}
	for _, key := range this.FilmHashes {
		if err := utils.WriteBytes(w, key); err != nil {
			return fmt.Errorf("[UserFilmList] [ProfitIds:%v] serialize from error:%v", key, err)
		}
	}
	return nil
}

func (this *UserFilmList) Deserialize(r io.Reader) error {
	var err error
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[UserFilmList] [Owner] deserialize from error:%v", err)
	}
	if this.Num, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[UserFilmList] [Num] deserialize from error:%v", err)
	}
	data := make([][]byte, 0, this.Num)
	for i := uint64(0); i < this.Num; i++ {
		var d []byte
		if d, err = utils.ReadBytes(r); err != nil {
			return fmt.Errorf("[UserFilmList] [key] deserialize from error:%v", err)
		}
		data = append(data, d)
	}
	this.FilmHashes = data
	return nil
}
