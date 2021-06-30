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
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func InitFilm() {
	native.Contracts[utils.FilmContractAddress] = RegisterFilmContract
}

func RegisterFilmContract(native *native.NativeService) {
	native.Register(FILM_PUBLISH, FilmPublish)
	native.Register(FILM_UPDATE, FilmUpdate)
	native.Register(GET_FILM_LIST, GetAllFilmList)
	native.Register(FILM_GETINFO, GetFilmInfo)
	native.Register(GET_USER_FILM_LIST, GetUserFilmList)
	native.Register(BUY_FILM, BuyFilm)
	native.Register(GET_USER_BUY_RECORD_LIST, GetUserBuyRecordList)
	native.Register(GET_USER_PROFIT_RECORD_LIST, GetUserProfitRecordList)
}
