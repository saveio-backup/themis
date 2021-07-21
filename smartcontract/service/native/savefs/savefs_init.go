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

package savefs

import (
	"bytes"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func InitFs() {
	native.Contracts[utils.OntFSContractAddress] = RegisterFsContract
}

func RegisterFsContract(native *native.NativeService) {
	//native.Register(FS_INIT, FsInit)
	native.Register(FS_GETSETTING, FsGetSetting)
	native.Register(FS_GETSTORAGEFEE, FsGetUploadStorageFee)
	native.Register(FS_NODE_REGISTER, FsNodeRegister)
	native.Register(FS_NODE_QUERY, FsNodeQuery)
	native.Register(FS_NODE_UPDATE, FsNodeUpdate)
	native.Register(FS_NODE_CANCEL, FsNodeCancel)
	native.Register(FS_GET_NODE_LIST, FsGetNodeList)
	native.Register(FS_GET_NODE_LIST_BY_ADDRS, FsGetNodeListByAddrs)
	native.Register(FS_STORE_FILE, FsStoreFile)
	native.Register(FS_FILE_RENEW, FsFileRenew)
	native.Register(FS_GET_FILE_INFO, FsGetFileInfo)
	native.Register(FS_GET_FILE_INFOS, FsGetFileInfos)
	native.Register(FS_GET_FILE_LIST, FsGetFileList)
	native.Register(FS_NODE_WITH_DRAW_PROFIT, FsNodeWithDrawProfit)
	native.Register(FS_FILE_PROVE, FsFileProve)
	native.Register(FS_GET_FILE_PROVE_DETAILS, FsGetFileProveDetails)
	native.Register(FS_DELETE_FILE, FsDeleteFile)
	native.Register(FS_DELETE_FILES, FsDeleteFiles)
	native.Register(FS_CHANGE_FILE_OWNER, FsChangeFileOwner)
	native.Register(FS_WHITE_LIST_OP, FsWhiteListOp)
	native.Register(FS_GET_WHITE_LIST, FsGetWhiteList)
	native.Register(FS_CHANGE_FILE_PRIVILEGE, FsChangeFilePrivilege)
	native.Register(FS_MANAGE_USER_SPACE, FsManageUserSpace)
	native.Register(FS_GET_USER_SPACE, FsGetUserSpace)
	native.Register(FS_GET_USER_SPACE_COST, FsGetUpdateCost)
	native.Register(FS_DELETE_USER_SPACE, FsDeleteUserSpace)
	native.Register(FS_GET_UNPROVE_PRIMARY_FILES, FsGetUnProvePrimaryFiles)
	native.Register(FS_GET_UNPROVE_CANDIDATE_FILES, FsGetUnProveCandidateFiles)

	native.Register(FS_CREATE_SECTOR, FsCreateSector)
	native.Register(FS_GET_SECTOR_INFO, FsGetSectorInfo)
	native.Register(FS_DELETE_SECTOR_INFO, FsDeleteSector)
	native.Register(FS_DELETE_FILE_IN_SECTOR, FsDeleteFileInSector)
	native.Register(FS_GET_SECTORS_FOR_NODE, FsGetSectorsForNode)
	native.Register(FS_SECTOR_PROVE, FsSectorProve)
	native.Register(FS_CHECK_NODE_SECTOR_PROVED_INTIME, FsCheckNodeSectorProvedInTime)

	native.Register(FS_GET_USER_UNSETTLED_FILES, FsGetUnSettledFiles)
	native.Register(FS_DELETE_UNSETTLED_FILES, FsDeleteUnsettledFiles)

	native.Register(FS_GET_POC_PROVELIST, FsGetPocProveList)
}

func FsInit(native *native.NativeService) ([]byte, error) {
	var fsSetting FsSetting
	infoSource := common.NewZeroCopySource(native.Input)
	if err := fsSetting.Deserialization(infoSource); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Init] FsSetting deserialize error!")
	}
	setFsSetting(native, fsSetting)
	return utils.BYTE_TRUE, nil
}

func FsGetSetting(native *native.NativeService) ([]byte, error) {
	fsSetting, err := getFsSetting(native)
	if err != nil || fsSetting == nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Init] GetFsSetting error!")
	}
	fs := new(bytes.Buffer)
	fsSetting.Serialize(fs)
	return EncRet(true, fs.Bytes()), nil
}

func FsGetUploadStorageFee(native *native.NativeService) ([]byte, error) {
	fsSetting, err := getFsSetting(native)
	if err != nil || fsSetting == nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Init] FsGetStorageFee error!")
	}
	source := common.NewZeroCopySource(native.Input)
	whiteListOpBytes, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsWhiteListOp DecodeBytes error!")
	}
	var uploadInfo UploadOption
	reader := bytes.NewReader(whiteListOpBytes)
	if err = uploadInfo.Deserialize(reader); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsWhiteListOp DecodeBytes error!")
	}
	log.Debugf("uploadInfo StorageType:%v", uploadInfo.StorageType)
	sf := calcUploadFee(&uploadInfo, fsSetting, native.Height)
	bf := new(bytes.Buffer)
	sf.Serialize(bf)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("FsGetUploadStorageFee default serialize !")
	}
	return EncRet(true, bf.Bytes()), nil
}
