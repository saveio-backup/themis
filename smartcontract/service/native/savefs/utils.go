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
	"fmt"
	"io"

	"github.com/saveio/themis/vm/wasmvm/util"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/usdt"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type FileStoreType int

const (
	FileStoreTypeNormal FileStoreType = iota
	FileStoreTypeProfessional
)
const (
	FS_INIT                            = "FsInit"
	FS_GETSETTING                      = "FsGetSetting"
	FS_GETSTORAGEFEE                   = "FsGetStorageFee"
	FS_NODE_REGISTER                   = "FsNodeRegister"
	FS_NODE_QUERY                      = "FsNodeQuery"
	FS_NODE_UPDATE                     = "FsNodeUpdate"
	FS_NODE_CANCEL                     = "FsNodeCancel"
	FS_GET_NODE_LIST                   = "FsGetNodeList"
	FS_GET_NODE_LIST_BY_ADDRS          = "FsGetNodeListByAddrs"
	FS_STORE_FILE                      = "FsStoreFile"
	FS_GET_FILE_INFO                   = "FsGetFileInfo"
	FS_GET_FILE_INFOS                  = "FsGetFileInfos"
	FS_GET_FILE_LIST                   = "FsGetFileList"
	FS_NODE_WITH_DRAW_PROFIT           = "FsNodeWithDrawProfit"
	FS_FILE_PROVE                      = "FsFileProve"
	FS_GET_FILE_PROVE_DETAILS          = "FsGetFileProveDetails"
	FS_DELETE_FILE                     = "FsDeleteFile"
	FS_DELETE_FILES                    = "FsDeleteFiles"
	FS_CHANGE_FILE_OWNER               = "FsChangeFileOwner"
	FS_WHITE_LIST_OP                   = "FsWhiteListOp"
	FS_GET_WHITE_LIST                  = "FsGetWhiteList"
	FS_FILE_RENEW                      = "FsFileRenew"
	FS_CHANGE_FILE_PRIVILEGE           = "FsChangeFilePrivilege"
	FS_MANAGE_USER_SPACE               = "FsManageUserSpace"
	FS_GET_USER_SPACE                  = "FsGetUserSpace"
	FS_GET_USER_SPACE_COST             = "FsGetUpdateCost"
	FS_DELETE_USER_SPACE               = "FsDeleteUserSpace"
	FS_GET_UNPROVE_PRIMARY_FILES       = "FsGetUnProvePrimaryFiles"
	FS_GET_UNPROVE_CANDIDATE_FILES     = "FsGetUnProveCandidateFiles"
	FS_CREATE_SECTOR                   = "FsCreateSector"
	FS_GET_SECTOR_INFO                 = "FsGetSectorInfo"
	FS_DELETE_SECTOR_INFO              = "FsDeleteSectorInfo"
	FS_DELETE_FILE_IN_SECTOR           = "FsDeleteFileInSector"
	FS_GET_SECTORS_FOR_NODE            = "FsGetSectorsForNode"
	FS_SECTOR_PROVE                    = "FsSectorProve"
	FS_CHECK_NODE_SECTOR_PROVED_INTIME = "FsCheckNodeSectorProvedInTime"
	FS_GET_USER_UNSETTLED_FILES        = "FsGetUserUnsettledFiles"
	FS_DELETE_UNSETTLED_FILES          = "FsDeleteUnsettledFiles"
	FS_GET_POC_PROVELIST               = "FsGetPocProveList"
)

const (
	SAVEFS_SETTING                    = "savefssetting"
	SAVEFS_NODE_INFO                  = "savefsnodeInfo"
	SAVEFS_NODE_SET                   = "savefsnodeset"
	SAVEFS_FILE_INFO                  = "savefsfileinfo"
	SAVEFS_FILE_WHITE_LIST            = "savefsfilewhitelist"
	SAVEFS_FILE_LIST                  = "savefsfilelist"
	SAVEFS_FILE_PROVE                 = "savefsfileprove"
	SAVEFS_FILE_READ_PLEDGE           = "savefsfilereadpledge"
	SAVEFS_USER_SPACE                 = "savefsuserspace"
	SAVEFS_PRIMARY_FILE_LIST          = "savefsprimaryfilelist"
	SAVEFS_CANDIDATE_FILE_LIST        = "savefscandidatefilelist"
	SAVEFS_UNSETTLED_FILE_LIST        = "savefsunsettledfilelist"
	SAVEFS_SECTOR_INFO                = "savefssectorinfo"
	SAVEFS_SECTOR_FILE_INFO_GROUP     = "savefssectorfileinfogroup"
	SAVEFS_SECTOR_FILE_INFO_GROUP_NUM = "savefsnumofsectorfileinfogroup"
	SAVEFS_SECTOR_PUNISHMENT_HEIGHT   = "savefssectorpunishmentheight"
	SAVEFS_MINER_PROVE_KEY            = "savefsminerpocprove"
	SAVEFS_MINER_PROVE_LIST_KEY       = "savefsminerpocprovelist"
)
const (
	FS_GAS_PRICE           = 1
	GAS_PER_GB_PER_Block   = 1
	GAS_PER_KB_FOR_READ    = 1
	GAS_FOR_CHALLENGE      = 200000
	MAX_PROVE_BLOCKS       = 32
	MIN_VOLUME             = 1000 * 1000
	DEFAULT_PROVE_PERIOD   = 3600 * 24 / DEFAULT_BLOCK_INTERVAL
	DeFAULT_PROVE_LEVEL    = PROVE_LEVEL_HIGH
	DEFAULT_COPY_NUM       = 2
	DEFAULT_BLOCK_INTERVAL = 5
	MIN_SECTOR_SIZE        = 1000 * 1000
)

type UploadOption struct {
	FileDesc        []byte
	FileSize        uint64
	ProveInterval   uint64
	ProveLevel      uint64
	ExpiredHeight   uint64
	Privilege       uint64
	CopyNum         uint64
	Encrypt         bool
	EncryptPassword []byte
	RegisterDNS     bool
	BindDNS         bool
	DnsURL          []byte
	WhiteList       WhiteList
	Share           bool
	StorageType     uint64
}

func (this *UploadOption) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.FileDesc); err != nil {
		return fmt.Errorf("[UploadOption] [FileDesc:%v] serialize from error:%v", this.FileDesc, err)
	}
	if err := utils.WriteVarUint(w, this.FileSize); err != nil {
		return fmt.Errorf("[UploadOption] [FileSize:%v] serialize from error:%v", this.FileSize, err)
	}
	if err := utils.WriteVarUint(w, this.ProveInterval); err != nil {
		return fmt.Errorf("[UploadOption] [ProveInterval:%v] serialize from error:%v", this.ProveInterval, err)
	}
	if err := utils.WriteVarUint(w, this.ProveLevel); err != nil {
		return fmt.Errorf("[UploadOption] [ProveLevel:%v] serialize from error:%v", this.ProveLevel, err)
	}
	if err := utils.WriteVarUint(w, this.ExpiredHeight); err != nil {
		return fmt.Errorf("[UploadOption] [ExpiredHeight:%v] serialize from error:%v", this.ExpiredHeight, err)
	}
	if err := utils.WriteVarUint(w, this.Privilege); err != nil {
		return fmt.Errorf("[UploadOption] [Privilege:%v] serialize from error:%v", this.Privilege, err)
	}
	if err := utils.WriteVarUint(w, this.CopyNum); err != nil {
		return fmt.Errorf("[UploadOption] [CopyNum:%v] serialize from error:%v", this.CopyNum, err)
	}
	if err := utils.WriteBool(w, this.Encrypt); err != nil {
		return fmt.Errorf("[UploadOption] [Encrypt:%v] serialize from error:%v", this.Encrypt, err)
	}
	if err := utils.WriteBytes(w, this.EncryptPassword); err != nil {
		return fmt.Errorf("[UploadOption] [EncryptPassword:%v] serialize from error:%v", this.EncryptPassword, err)
	}
	if err := utils.WriteBool(w, this.RegisterDNS); err != nil {
		return fmt.Errorf("[UploadOption] [RegisterDNS:%v] serialize from error:%v", this.RegisterDNS, err)
	}
	if err := utils.WriteBool(w, this.BindDNS); err != nil {
		return fmt.Errorf("[UploadOption] [BindDNS:%v] serialize from error:%v", this.BindDNS, err)
	}
	if err := utils.WriteBytes(w, this.DnsURL); err != nil {
		return fmt.Errorf("[UploadOption] [DnsURL:%v] serialize from error:%v", this.DnsURL, err)
	}
	if err := this.WhiteList.Serialize(w); err != nil {
		return fmt.Errorf("[UploadOption] [WhiteList:%v] serialize from error:%v", this.WhiteList, err)
	}
	if err := utils.WriteBool(w, this.Share); err != nil {
		return fmt.Errorf("[UploadOption] [Share:%v] serialize from error:%v", this.Share, err)
	}
	if err := utils.WriteVarUint(w, this.StorageType); err != nil {
		return fmt.Errorf("[UploadOption] [StorageType:%v] serialize from error:%v", this.StorageType, err)
	}
	return nil
}

func (this *UploadOption) Deserialize(r io.Reader) error {
	var err error
	this.FileDesc, err = utils.ReadBytes(r)
	if err != nil {
		log.Errorf("file desc err %v", err)
		return err
	}
	this.FileSize, err = utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	this.ProveInterval, err = utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	this.ProveLevel, err = utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	this.ExpiredHeight, err = utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	this.Privilege, err = utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	this.CopyNum, err = utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	this.Encrypt, err = utils.ReadBool(r)
	if err != nil {
		return err
	}
	this.EncryptPassword, err = utils.ReadBytes(r)
	if err != nil {
		return err
	}
	this.RegisterDNS, err = utils.ReadBool(r)
	if err != nil {
		return err
	}
	this.BindDNS, err = utils.ReadBool(r)
	if err != nil {
		return err
	}
	this.DnsURL, err = utils.ReadBytes(r)
	if err != nil {
		return err
	}
	var whitelist WhiteList
	err = whitelist.Deserialize(r)
	if err != nil {
		return err
	}
	this.WhiteList = whitelist
	this.Share, err = utils.ReadBool(r)
	if err != nil {
		return err
	}
	this.StorageType, err = utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	return nil
}

func GenFsSettingKey(contract common.Address) []byte {
	return append(contract[:], SAVEFS_SETTING...)
}

func GenFsNodeInfoKey(contract common.Address, walletAddr common.Address) []byte {
	key := append(contract[:], SAVEFS_NODE_INFO...)
	return append(key, walletAddr[:]...)
}

func GenFsNodeSetKey(contract common.Address) []byte {
	return append(contract[:], SAVEFS_NODE_SET...)
}

func GenFsFileInfoKey(contract common.Address, fileHash []byte) []byte {
	key := append(contract[:], SAVEFS_FILE_INFO...)
	return append(key, fileHash[:]...)
}

func GenFsFileListKey(contract common.Address, walletAddr common.Address) []byte {
	key := append(contract[:], SAVEFS_FILE_LIST...)
	return append(key, walletAddr[:]...)
}

func GenFsProveDetailsKey(contract common.Address, fileHash []byte) []byte {
	key := append(contract[:], SAVEFS_FILE_PROVE...)
	return append(key, fileHash[:]...)
}

func GenFsFileReadPledgeKey(contract common.Address, userAddr []byte, fileHash []byte) []byte {
	key := append(contract[:], SAVEFS_FILE_READ_PLEDGE...)
	key = append(key[:], userAddr...)
	return append(key, fileHash[:]...)
}

func GenFsWhiteListKey(contract common.Address, fileHash []byte) []byte {
	key := append(contract[:], SAVEFS_FILE_WHITE_LIST...)
	return append(key, fileHash[:]...)
}

func GenFsUserSpaceKey(contract common.Address, walletAddr common.Address) []byte {
	key := append(contract[:], SAVEFS_USER_SPACE...)
	return append(key, walletAddr[:]...)
}

func GenFsFilePrimaryListKey(contract common.Address, walletAddr common.Address) []byte {
	key := append(contract[:], SAVEFS_PRIMARY_FILE_LIST...)
	return append(key, walletAddr[:]...)
}

func GenFsFileCandidateListKey(contract common.Address, walletAddr common.Address) []byte {
	key := append(contract[:], SAVEFS_CANDIDATE_FILE_LIST...)
	return append(key, walletAddr[:]...)
}

func GenFsUnSettledListKey(contract common.Address, walletAddr common.Address) []byte {
	key := append(contract[:], SAVEFS_UNSETTLED_FILE_LIST...)
	return append(key, walletAddr[:]...)
}

func GenFsSectorInfoPrefix(contract common.Address, nodeAddr common.Address) []byte {
	key := append(contract[:], SAVEFS_SECTOR_INFO...)
	return append(key, nodeAddr[:]...)
}

func GenFsSectorInfoKey(contract common.Address, nodeAddr common.Address, sectorID uint64) []byte {
	prefix := GenFsSectorInfoPrefix(contract, nodeAddr)
	return append(prefix, util.Int64ToBytes(sectorID)...)
}

func GenFsSectorFileInfoGroupPrefix(contract common.Address, nodeAddr common.Address, sectorID uint64) []byte {
	key := append(contract[:], SAVEFS_SECTOR_FILE_INFO_GROUP...)
	key = append(key[:], nodeAddr[:]...)
	key = append(key[:], util.Int64ToBytes(sectorID)...)
	return key
}

func GenFsSectorFileInfoGroupKey(contract common.Address, nodeAddr common.Address, sectorID uint64, groupID uint64) []byte {
	prefix := GenFsSectorFileInfoGroupPrefix(contract, nodeAddr, sectorID)
	return append(prefix[:], util.Int64ToBytes(groupID)...)
}

func GenFsSectorFileInfoGroupNumKey(contract common.Address, nodeAddr common.Address, sectorID uint64) []byte {
	key := append(contract[:], SAVEFS_SECTOR_FILE_INFO_GROUP_NUM...)
	key = append(key[:], nodeAddr[:]...)
	key = append(key[:], util.Int64ToBytes(sectorID)...)
	return key
}

func GenFsNodeSectorPunishmentKey(contract common.Address, nodeAddr common.Address, sectorID uint64) []byte {
	key := append(contract[:], SAVEFS_SECTOR_PUNISHMENT_HEIGHT...)
	key = append(key[:], nodeAddr[:]...)
	key = append(key[:], util.Int64ToBytes(sectorID)...)
	return key
}

func GenPocProveKey(contract common.Address, miner common.Address, height uint64) []byte {
	key := append(contract[:], SAVEFS_MINER_PROVE_KEY...)
	key = append(key[:], miner[:]...)
	key = append(key[:], util.Int64ToBytes(height)...)
	return key
}

func GenPocProveListKey(contract common.Address, height uint64) []byte {
	key := append(contract[:], SAVEFS_MINER_PROVE_LIST_KEY...)
	key = append(key[:], util.Int64ToBytes(height)...)
	return key
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
