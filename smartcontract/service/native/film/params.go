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

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type Parameter struct {
	Type         []byte
	BytesValue   []byte
	UintValue    uint64
	AddressValue common.Address
	BoolValue    bool
}

func (this *Parameter) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Type, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	switch string(this.Type) {
	case "bytes":
		this.BytesValue, err = utils.DecodeBytes(source)
		if err != nil {
			return err
		}
	case "uint":
		this.UintValue, err = utils.DecodeVarUint(source)
		if err != nil {
			return err
		}
	case "address":
		this.AddressValue, err = utils.DecodeAddress(source)
		if err != nil {
			return err
		}
	case "bool":
		this.BoolValue, err = utils.DecodeBool(source)
		if err != nil {
			return err
		}
	}

	return nil
}

type BaseParams struct {
	Num  uint64
	Args []Parameter
}

func (this *BaseParams) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Num, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	temp := make([]Parameter, 0)
	for i := uint64(0); i < this.Num; i++ {
		// arg := new(Parameter)
		var arg Parameter
		err := arg.Deserialization(source)
		if err != nil {
			return err
		}
		temp = append(temp, arg)
	}
	this.Args = temp
	return nil
}

type SearchFilmParams struct {
	Name        []byte
	Avaiable    uint64
	Type        uint64
	ReleaseYear uint64
	Region      []byte
	WalletAddr  common.Address
	Offset      uint64
	Limit       uint64
}

func (this *SearchFilmParams) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Name, err = utils.DecodeBytes(source)
	if err != nil {
		fmt.Println("decode anme err")
		return err
	}
	this.Avaiable, err = utils.DecodeVarUint(source)
	if err != nil {
		fmt.Println("decode avaliable err")
		return err
	}
	fmt.Printf("Avaiable %v\n", this.Avaiable)
	this.Type, err = utils.DecodeVarUint(source)
	if err != nil {
		fmt.Println("decode type err")
		return err
	}
	this.ReleaseYear, err = utils.DecodeVarUint(source)
	if err != nil {
		fmt.Println("decode year err")
		return err
	}
	fmt.Printf("this.ReleaseYear :%v , Type %v\n", this.ReleaseYear, this.Type)
	this.Region, err = utils.DecodeBytes(source)
	if err != nil {
		fmt.Println("decode region err")
		return err
	}
	this.WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		fmt.Println("decode wallet err")
		return err
	}
	this.Offset, err = utils.DecodeVarUint(source)
	if err != nil {
		fmt.Println("decode offset err")
		return err
	}
	this.Limit, err = utils.DecodeVarUint(source)
	if err != nil {
		fmt.Println("decode limit err")
		return err
	}
	return nil
}

type PublishFilmParam struct {
	Cover       []byte
	Url         []byte
	Name        []byte
	Desc        []byte
	Type        uint64
	ReleaseYear uint64
	Language    []byte
	Region      []byte
	Price       uint64
	CreatedAt   uint64
}

func (this *PublishFilmParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	fmt.Printf("source len %d\n", source.Len())
	this.Cover, err = utils.DecodeBytes(source)
	fmt.Printf("this cover %v, %s\n", this.Cover, err)
	fmt.Printf("source len %d\n", source.Len())
	if err != nil {
		return err
	}
	this.Url, err = utils.DecodeBytes(source)
	fmt.Printf("this url %v, %s\n", this.Url, err)
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
	return nil
}

type FilmHashAddrParams struct {
	Hash       []byte
	WalletAddr common.Address
}

func (this *FilmHashAddrParams) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Hash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	return nil
}
