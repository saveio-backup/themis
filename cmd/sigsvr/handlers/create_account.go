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
package handlers

import (
	"encoding/json"

	"github.com/saveio/themis/crypto/keypair"
	s "github.com/saveio/themis/crypto/signature"
	clisvrcom "github.com/saveio/themis/cmd/sigsvr/common"
	"github.com/saveio/themis/common/log"
)

type CreateAccountReq struct {
}

type CreateAccountRsp struct {
	Account string `json:"account"`
}

func CreateAccount(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	pwd := req.Pwd
	if pwd == "" {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		resp.ErrorInfo = "pwd cannot empty"
		return
	}
	accData, err := clisvrcom.DefWalletStore.NewAccountData(keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, []byte(pwd))
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		resp.ErrorInfo = "create wallet failed"
		log.Errorf("CreateAccount Qid:%s NewAccountData error:%s", req.Qid, err)
		return
	}
	_, err = clisvrcom.DefWalletStore.AddAccountData(accData)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		resp.ErrorInfo = "create wallet failed"
		log.Errorf("CreateAccount Qid:%s AddAccountData error:%s", req.Qid, err)
		return
	}
	resp.Result = &CreateAccountRsp{
		Account: accData.Address,
	}

	data, _ := json.Marshal(accData)
	log.Infof("[CreateAccount]%s", data)
}
