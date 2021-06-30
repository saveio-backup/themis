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
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/usdt"
)

const (
	FILM_PUBLISH                = "FilmPublish"
	BUY_FILM                    = "BuyFilm"
	FILM_UPDATE                 = "FilmUpdate"
	GET_FILM_LIST               = "GetFilmList"
	FILM_GETINFO                = "FilmGetInfo"
	GET_USER_FILM_LIST          = "GetUserFilmList"
	GET_USER_BUY_RECORD_LIST    = "GetUserBuyRecordList"
	GET_USER_PROFIT_RECORD_LIST = "GetUserProfitRecordList"
)

const (
	USER_FILM_LIST = "filmuserfilmlist"
	ALL_FILM_LIST  = "filmallfilmlist"
	FILM_INFO      = "filminfo"
	USER_BUY_LIST  = "filmuserfilmbuylist"
	USER_BUY_INFO  = "filmuserfilmbuyinfo"

	USER_PROFIT_LIST = "filmuserfilmprofitlist"
	USER_PROFIT_INFO = "filmuserfilmprofitinfo"

	FILM_COUNT = "filmcount"

	SEARCH_KEY_PATTERN = "type=%d&year=%d&region=%v&available=%v"
)

func GenFilmInfoKey(contract, owner common.Address, fileHash []byte) []byte {
	key := append(contract[:], []byte(FILM_INFO)...)
	key = append(key, owner[:]...)
	return append(key, fileHash...)
}

func GenAllFilmListKey(contract common.Address) []byte {
	return append(contract[:], []byte(ALL_FILM_LIST)...)
}

func GenUserFilmListKey(contract, owner common.Address) []byte {
	key := append(contract[:], []byte(USER_FILM_LIST)...)
	return append(key, owner[:]...)
}

func GenUserFilmBuyListKey(contract, owner common.Address) []byte {
	key := append(contract[:], []byte(USER_BUY_LIST)...)
	return append(key, owner[:]...)
}

func GenUserFilmBuyInfoKey(contract, owner common.Address, id []byte) []byte {
	key := append(contract[:], []byte(USER_BUY_INFO)...)
	key = append(key, owner[:]...)
	return append(key, id...)
}

func GenUserFilmProfitListKey(contract, owner common.Address) []byte {
	key := append(contract[:], []byte(USER_PROFIT_LIST)...)
	return append(key, owner[:]...)
}

func GenUserFilmProfitInfoKey(contract, owner common.Address, id []byte) []byte {
	key := append(contract[:], []byte(USER_PROFIT_INFO)...)
	key = append(key, owner[:]...)
	return append(key, id...)
}

func GetFilmCountKey(contract common.Address) []byte {
	return append(contract[:], []byte(FILM_COUNT)...)
}

func GetFilmKeyAtList(contract common.Address, index uint64) []byte {
	key := append(contract[:], []byte(FILM_COUNT)...)
	key = append(key, []byte(fmt.Sprintf("-%d", index))...)
	return key
}

func getStringValue(value interface{}) string {
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

func getUint64Value(value interface{}) uint64 {
	f64, ok := value.(float64)
	if !ok {
		return 0
	}
	return uint64(f64)
}

func getBoolValue(value interface{}) bool {
	f64, ok := value.(bool)
	if !ok {
		return false
	}
	return bool(f64)
}

func appCallTransfer(native *native.NativeService, contract common.Address, from common.Address, to common.Address, amount uint64) error {
	var sts []usdt.State
	sts = append(sts, usdt.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := usdt.Transfers{
		States: sts,
	}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)

	if _, err := native.NativeCall(contract, "transfer", sink.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransfer, appCall error!")
	}
	return nil
}
