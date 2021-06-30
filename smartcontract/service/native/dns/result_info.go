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
package dns

import (
	"fmt"
	"io"

	"bytes"

	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type RetInfo struct {
	Ret  bool
	Info []byte
}

func (this *RetInfo) Serialize(w io.Writer) error {
	if err := utils.WriteBool(w, this.Ret); err != nil {
		return fmt.Errorf("[RetInfo] [Ret:%v] serialize from error:%v", this.Ret, err)
	}
	if err := utils.WriteBytes(w, this.Info); err != nil {
		return fmt.Errorf("[RetInfo] [Info:%v] serialize from error:%v", this.Info, err)
	}
	return nil
}

func (this *RetInfo) Deserialize(r io.Reader) error {
	var err error
	if this.Ret, err = utils.ReadBool(r); err != nil {
		return fmt.Errorf("[RetInfo] [Ret] deserialize from error:%v", err)
	}
	if this.Info, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[RetInfo] [Info] deserialize from error:%v", err)
	}
	return nil
}

func EncRet(ret bool, info []byte) []byte {
	retInfo := RetInfo{ret, info}
	bf := new(bytes.Buffer)
	retInfo.Serialize(bf)
	return bf.Bytes()
}

func DecRet(ri []byte) *RetInfo {
	var retInfo RetInfo
	reader := bytes.NewReader(ri)
	retInfo.Deserialize(reader)
	return &retInfo
}
