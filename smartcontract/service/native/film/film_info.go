/*
 * Copyright (C) 2019 The themis Authors
 * This file is part of The themis library.
 *
 * The themis is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The themis is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */
package film

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type FilmInfo struct {
	Id           []byte
	Hash         []byte
	Cover        []byte
	Url          []byte
	Name         []byte
	Desc         []byte
	Available    bool
	Type         uint64
	ReleaseYear  uint64
	Language     []byte
	Region       []byte
	Price        uint64
	CreatedAt    uint64
	PaidCount    uint64
	TotalProfit  uint64
	FileSize     uint64
	RealFileSize uint64
	Owner        common.Address
}

func (this *FilmInfo) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.Id); err != nil {
		return fmt.Errorf("[FilmInfo] [Id:%v] serialize from error:%v", this.Id, err)
	}
	if err := utils.WriteBytes(w, this.Hash); err != nil {
		return fmt.Errorf("[FilmInfo] [Hash:%v] serialize from error:%v", this.Hash, err)
	}
	if err := utils.WriteBytes(w, this.Cover); err != nil {
		return fmt.Errorf("[FilmInfo] [Cover:%v] serialize from error:%v", this.Cover, err)
	}
	if err := utils.WriteBytes(w, this.Url); err != nil {
		return fmt.Errorf("[FilmInfo] [Url:%v] serialize from error:%v", this.Url, err)
	}
	if err := utils.WriteBytes(w, this.Name); err != nil {
		return fmt.Errorf("[FilmInfo] [Name:%v] serialize from error:%v", this.Name, err)
	}
	if err := utils.WriteBytes(w, this.Desc); err != nil {
		return fmt.Errorf("[FilmInfo] [Desc:%v] serialize from error:%v", this.Desc, err)
	}
	if err := utils.WriteBool(w, this.Available); err != nil {
		return fmt.Errorf("[FilmInfo] [Available:%v] serialize from error:%v", this.Available, err)
	}
	if err := utils.WriteVarUint(w, this.Type); err != nil {
		return fmt.Errorf("[FilmInfo] [Type:%v] serialize from error:%v", this.Type, err)
	}
	if err := utils.WriteVarUint(w, this.ReleaseYear); err != nil {
		return fmt.Errorf("[FilmInfo] [ReleaseYear:%v] serialize from error:%v", this.ReleaseYear, err)
	}
	if err := utils.WriteBytes(w, this.Language); err != nil {
		return fmt.Errorf("[FilmInfo] [Language:%v] serialize from error:%v", this.Language, err)
	}
	if err := utils.WriteBytes(w, this.Region); err != nil {
		return fmt.Errorf("[FilmInfo] [Region:%v] serialize from error:%v", this.Region, err)
	}
	if err := utils.WriteVarUint(w, this.Price); err != nil {
		return fmt.Errorf("[FilmInfo] [Price:%v] serialize from error:%v", this.Price, err)
	}
	if err := utils.WriteVarUint(w, this.CreatedAt); err != nil {
		return fmt.Errorf("[FilmInfo] [CreatedAt:%v] serialize from error:%v", this.CreatedAt, err)
	}
	if err := utils.WriteVarUint(w, this.PaidCount); err != nil {
		return fmt.Errorf("[FilmInfo] [PaidCount:%v] serialize from error:%v", this.PaidCount, err)
	}
	if err := utils.WriteVarUint(w, this.TotalProfit); err != nil {
		return fmt.Errorf("[FilmInfo] [TotalProfit:%v] serialize from error:%v", this.TotalProfit, err)
	}
	if err := utils.WriteVarUint(w, this.FileSize); err != nil {
		return fmt.Errorf("[FilmInfo] [FileSize:%v] serialize from error:%v", this.FileSize, err)
	}
	if err := utils.WriteVarUint(w, this.RealFileSize); err != nil {
		return fmt.Errorf("[FilmInfo] [RealFileSize:%v] serialize from error:%v", this.RealFileSize, err)
	}
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("[FilmInfo] [Owner:%v] serialize from error:%v", this.Owner, err)
	}
	return nil
}

func (this *FilmInfo) Deserialize(r io.Reader) error {
	var err error
	if this.Id, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Id] deserialize from error:%v", err)
	}
	if this.Hash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Hash] deserialize from error:%v", err)
	}
	if this.Cover, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Cover] deserialize from error:%v", err)
	}
	if this.Url, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Url] deserialize from error:%v", err)
	}
	if this.Name, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Name] deserialize from error:%v", err)
	}
	if this.Desc, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Desc] deserialize from error:%v", err)
	}
	if this.Available, err = utils.ReadBool(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Avaiable] deserialize from error:%v", err)
	}
	if this.Type, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Type] deserialize from error:%v", err)
	}
	if this.ReleaseYear, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [ReleaseYear] deserialize from error:%v", err)
	}
	if this.Language, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Language] deserialize from error:%v", err)
	}
	if this.Region, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Region] deserialize from error:%v", err)
	}
	if this.Price, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [Price] deserialize from error:%v", err)
	}
	if this.CreatedAt, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [CreatedAt] deserialize from error:%v", err)
	}
	if this.PaidCount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [PaidCount] deserialize from error:%v", err)
	}
	if this.TotalProfit, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [TotalProfit] deserialize from error:%v", err)
	}
	if this.FileSize, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [FileSize] deserialize from error:%v", err)
	}
	if this.RealFileSize, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FilmInfo] [RealFileSize] deserialize from error:%v", err)
	}
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[FilmInfo] [ReadAddress] deserialize from error:%v", err)
	}
	return nil
}

func (this *FilmInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.Id)
	utils.EncodeBytes(sink, this.Hash)
	utils.EncodeBytes(sink, this.Cover)
	utils.EncodeBytes(sink, this.Url)
	utils.EncodeBytes(sink, this.Name)
	utils.EncodeBytes(sink, this.Desc)
	utils.EncodeBool(sink, this.Available)
	utils.EncodeVarUint(sink, this.Type)
	utils.EncodeVarUint(sink, this.ReleaseYear)
	utils.EncodeBytes(sink, this.Language)
	utils.EncodeBytes(sink, this.Region)
	utils.EncodeVarUint(sink, this.Price)
	utils.EncodeVarUint(sink, this.CreatedAt)
	utils.EncodeVarUint(sink, this.PaidCount)
	utils.EncodeVarUint(sink, this.TotalProfit)
	utils.EncodeVarUint(sink, this.FileSize)
	utils.EncodeVarUint(sink, this.RealFileSize)
	utils.EncodeAddress(sink, this.Owner)
}

func (this *FilmInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Id, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Hash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Cover, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Url, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Name, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Desc, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Available, err = utils.DecodeBool(source)
	if err != nil {
		return err
	}
	this.Type, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ReleaseYear, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Language, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Region, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Price, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.CreatedAt, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.PaidCount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.TotalProfit, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FileSize, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.RealFileSize, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Owner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	return nil
}
