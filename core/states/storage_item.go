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

package states

import (
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/errors"
)

type StorageItem struct {
	StateBase
	Value []byte
}

func (this *StorageItem) Serialization(sink *common.ZeroCopySink) {
	this.StateBase.Serialization(sink)
	sink.WriteVarBytes(this.Value)
}

func (this *StorageItem) Deserialization(source *common.ZeroCopySource) error {
	err := this.StateBase.Deserialization(source)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageItem], StateBase Deserialize failed.")
	}
	value, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return errors.NewDetailErr(common.ErrIrregularData, errors.ErrNoCode, "[StorageItem], Value Deserialize failed.")
	}
	if eof {
		return errors.NewDetailErr(io.ErrUnexpectedEOF, errors.ErrNoCode, "[StorageItem], Value Deserialize failed.")
	}
	this.Value = value
	return nil
}

func (storageItem *StorageItem) ToArray() []byte {
	return common.SerializeToBytes(storageItem)
}

func GetValueFromRawStorageItem(raw []byte) ([]byte, error) {
	item := StorageItem{}
	err := item.Deserialization(common.NewZeroCopySource(raw))
	if err != nil {
		return nil, err
	}

	return item.Value, nil
}

func GenRawStorageItem(value []byte) []byte {
	item := StorageItem{}
	item.Value = value
	return item.ToArray()
}
