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
	"fmt"
	"math/rand"
	"time"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func FsGetNodeList(native *native.NativeService) ([]byte, error) {
	nodeList, err := getFsNodeList(native)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetNodeList getFsNodeList error!")), nil
	}
	nodeAddrList := nodeList.GetList()
	if nodeAddrList == nil {
		return nil, fmt.Errorf("[FS Govern] NodeList GetList is nil")
	}
	if 0 == len(nodeAddrList) {
		return EncRet(true, nil), nil
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	r.Shuffle(len(nodeAddrList), func(i, j int) {
		nodeAddrList[i], nodeAddrList[j] = nodeAddrList[j], nodeAddrList[i]
	})

	bf := new(bytes.Buffer)
	var fsNodesInfo FsNodesInfo
	for _, addr := range nodeAddrList {
		fsNodeInfo, err := getFsNodeInfo(native, addr)
		if err != nil {
			fmt.Errorf("[FS Profit] FsGetNodeList getFsNodeInfo(%v) error", addr)
			continue
		}
		fsNodesInfo.NodeInfo = append(fsNodesInfo.NodeInfo, *fsNodeInfo)
		fsNodesInfo.NodeNum++
	}
	err = fsNodesInfo.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetNodeList FsNodeInfos serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

func FsGetNodeListByAddrs(native *native.NativeService) ([]byte, error) {
	var nodeList NodeList
	source := common.NewZeroCopySource(native.Input)
	nodeListBytes, err := utils.DecodeBytes(source)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetNodeListByAddrs DecodeBytes error!")), nil
	}
	reader := bytes.NewReader(nodeListBytes)
	if err = nodeList.Deserialize(reader); err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetNodeListByAddrs Deserialize error!")), nil
	}
	log.Infof("searchNode %v %v", nodeList.AddrNum, nodeList.AddrList)
	if nodeList.AddrNum == 0 {
		return EncRet(true, nil), nil
	}
	bf := new(bytes.Buffer)
	var fsNodesInfo FsNodesInfo
	for _, addr := range nodeList.AddrList {
		fsNodeInfo, err := getFsNodeInfo(native, addr)
		if err != nil {
			log.Errorf("[FS Profit] FsGetNodeList getFsNodeInfo(%v) error", addr)
			continue
		}
		// if fsNodeInfo.WalletAddr !=
		fsNodesInfo.NodeInfo = append(fsNodesInfo.NodeInfo, *fsNodeInfo)
		fsNodesInfo.NodeNum++
	}
	err = fsNodesInfo.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetNodeList FsNodeInfos serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

func FsStoreFile(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	source := common.NewZeroCopySource(native.Input)
	fileInfoBytes, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsStoreFile DecodeBytes error!")
	}
	var fileInfo FileInfo
	reader := bytes.NewReader(fileInfoBytes)
	if err = fileInfo.Deserialize(reader); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsStoreFile DecodeBytes error!")
	}

	if !native.ContextRef.CheckWitness(fileInfo.FileOwner) {
		return utils.BYTE_FALSE, errors.NewErr("FS Profit] CheckWitness failed!")
	}

	fileInfoKey := GenFsFileInfoKey(contract, fileInfo.FileHash)
	item, err := utils.GetStorageItem(native, fileInfoKey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] GetStorageItem error!")
	}
	if item != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] File have stored!")
	}

	if fileInfo.ExpiredHeight <= uint64(native.Height) {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] Wrong expire height!")
	}

	switch fileInfo.ProveLevel {
	case PROVE_LEVEL_HIGH:
	case PROVE_LEVEL_MEDIEUM:
	case PROVE_LEVEL_LOW:
	default:
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] invalid prove level!")
	}
	// get prove interval according to prove level
	fileInfo.ProveInterval = GetProveIntervalByProveLevel(fileInfo.ProveLevel)
	fsSetting, err := getFsSettingWithProveLevel(native, fileInfo.ProveLevel)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsStoreFile getFsSettingWithProveLevel error!")
	}

	fileInfo.ValidFlag = true
	uploadOpt := &UploadOption{
		ExpiredHeight: fileInfo.ExpiredHeight,
		ProveInterval: fileInfo.ProveInterval,
		CopyNum:       fileInfo.CopyNum,
		FileSize:      fileInfo.FileBlockSize * fileInfo.FileBlockNum,
	}
	uploadFee := calcDepositFee(uploadOpt, fsSetting, native.Height)
	log.Debugf("deposit fee %d %d", uploadFee.ValidationFee, uploadFee.SpaceFee)
	fileInfo.Deposit = uploadFee.Sum()
	fileInfo.ProveTimes = calcProveTimesByUploadInfo(uploadOpt, native.Height)

	log.Debugf("rate:%d, blkNum:%d, blkSize: %d, gaskbblk: %d, gasC: %d, times:%d gasPrice:%d, copyNum:%d, deposit :%d\n", fileInfo.ProveInterval, fileInfo.FileBlockNum,
		fileInfo.FileBlockSize, fsSetting.GasPerKBForRead, fsSetting.GasForChallenge, fileInfo.ProveTimes, fsSetting.FsGasPrice, fileInfo.CopyNum, fileInfo.Deposit)

	if fileInfo.StorageType == FileStorageTypeUseSpace {
		userspace, err := getUserSpace(native, fileInfo.FileOwner)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("FS Profit] GetUserSpace error!")
		}

		if userspace.Balance < fileInfo.Deposit {
			return utils.BYTE_FALSE, errors.NewErr("FS Profit] Userspace insufficient balance!")
		}

		if userspace.Remain < fileInfo.FileBlockNum*fileInfo.FileBlockSize {
			return utils.BYTE_FALSE, errors.NewErr("FS Profit] Userspace insufficient remain storage!")
		}

		if userspace.ExpireHeight != fileInfo.ExpiredHeight {
			return utils.BYTE_FALSE, errors.NewErr("FS Profit] Userspace wrong expiredHeight!")
		}

		log.Debugf("userspace store file: %v, fileinfo: %v", userspace, fileInfo)
		userspace.Balance -= fileInfo.Deposit
		userspace.Remain -= fileInfo.FileBlockNum * fileInfo.FileBlockSize
		userspace.Used += fileInfo.FileBlockNum * fileInfo.FileBlockSize

		if err = setUserSpace(native, userspace, fileInfo.FileOwner); err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Profit] SetUserSpace error!")
		}
	} else {
		log.Debugf("use transfer\n")
		err = appCallTransfer(native, utils.UsdtContractAddress, fileInfo.FileOwner, contract, fileInfo.Deposit)
		if err != nil {
			log.Errorf("transfer error from %s, to %s, amount %v, err %v", fileInfo.FileOwner, contract, fileInfo.Deposit, err)
			return utils.BYTE_FALSE, errors.NewErr("[FS Profit] AppCallTransfer, transfer error!")
		}
		fileInfo.StorageType = FileStorageTypeCustom
	}

	fileInfo.ProveBlockNum = fsSetting.MaxProveBlockNum
	fileInfo.BlockHeight = uint64(native.Height)

	if err = setFsFileInfo(native, &fileInfo); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsStoreFile setFsFileInfo error!")
	}

	if err = AddFileToList(native, fileInfo.FileOwner, fileInfo.FileHash); err != nil {
		return utils.BYTE_FALSE, err
	}

	for _, primaryWalletAddr := range fileInfo.PrimaryNodes.AddrList {
		if err = AddFileToPrimaryList(native, primaryWalletAddr, fileInfo.FileHash); err != nil {
			return utils.BYTE_FALSE, err
		}
	}

	for _, candidateWalletAddr := range fileInfo.CandidateNodes.AddrList {
		if err = AddFileToCandidateList(native, candidateWalletAddr, fileInfo.FileHash); err != nil {
			return utils.BYTE_FALSE, err
		}
	}

	var proveDetails FsProveDetails
	proveDetails.CopyNum = fileInfo.CopyNum
	proveDetails.ProveDetailNum = 0

	if err = setProveDetails(native, fileInfo.FileHash, &proveDetails); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] ProveDetails setProveDetails error:" + err.Error())
	}

	StoreFileEvent(native, fileInfo.FileHash, fileInfo.RealFileSize, fileInfo.FileOwner, fileInfo.Deposit)

	return utils.BYTE_TRUE, nil
}

func FsFileRenew(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var fileReNew FileReNew
	source := common.NewZeroCopySource(native.Input)
	if err := fileReNew.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsFileRenew deserialize error!")
	}

	if !native.ContextRef.CheckWitness(fileReNew.FromAddr) {
		return utils.BYTE_FALSE, errors.NewErr("FS Profit] FsFileRenew CheckWitness failed!")
	}

	fsSetting, err := getFsSetting(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsFileRenew getFsSetting error!")
	}

	fileInfo, err := getFsFileInfo(native, fileReNew.FileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] getFsFileInfo  error!")
	}

	if fileInfo.StorageType != FileStorageTypeCustom {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] wrong storage type!")
	}

	if uint64(native.Height) > fileInfo.ExpiredHeight {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsFileRenew File is not exist!")
	}

	totalRenew := calcFee(fsSetting, fileReNew.ReNewTimes, fileInfo.CopyNum, fileInfo.FileBlockNum*fileInfo.FileBlockSize, fileReNew.ReNewTimes*fileInfo.ProveInterval)
	reNewFee := totalRenew.ValidationFee + totalRenew.SpaceFee

	err = appCallTransfer(native, utils.UsdtContractAddress, fileReNew.FromAddr, contract, reNewFee)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] AppCallTransfer, transfer error!")
	}

	fileInfo.ProveTimes += fileReNew.ReNewTimes
	fileInfo.Deposit += reNewFee
	fileInfo.ExpiredHeight += fileInfo.ProveInterval * fileReNew.ReNewTimes

	if err = setFsFileInfo(native, fileInfo); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsFileRenew setFsFileInfo error:" + err.Error())
	}

	return utils.BYTE_TRUE, nil
}

func FsGetFileList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	walletAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileList DecodeAddress error!")), nil
	}
	fileList, err := GetFsFileList(native, walletAddr)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileList GetFsFileList error!")), nil
	}
	bf := new(bytes.Buffer)
	err = fileList.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileList FileList Serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

func FsGetFileInfo(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	fileHash, err := utils.DecodeBytes(source)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileInfo DecodeBytes error!")), nil
	}

	fileInfo, err := getFsFileInfo(native, fileHash)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileInfo getFsFileInfo error!")), nil
	}
	bf := new(bytes.Buffer)
	err = fileInfo.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileInfo FileInfo serialize error!")), nil
	}

	return EncRet(true, bf.Bytes()), nil
}

func FsGetFileInfos(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	fileListBytes, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsDeleteFiles DecodeBytes error!")
	}
	var fileList FileList
	reader := bytes.NewReader(fileListBytes)
	if err = fileList.Deserialize(reader); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsDeleteFiles DecodeBytes error!")
	}
	var fileInfos FileInfoList
	for _, fileHash := range fileList.List {
		item, err := getFsFileInfo(native, fileHash.Hash)
		if err != nil {
			return EncRet(false, []byte("[FS Profit] FsGetFileInfo GetStorageItem error!")), nil
		}
		if item == nil {
			return EncRet(false, []byte("[FS Profit] FsGetFileInfo not found!")), nil
		}
		fileInfos.FileNum++
		fileInfos.List = append(fileInfos.List, *item)
	}
	bf := new(bytes.Buffer)
	err = fileInfos.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileList FileList Serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

func FsWhiteListOp(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	whiteListOpBytes, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsWhiteListOp DecodeBytes error!")
	}

	var whiteListOp WhiteListOp
	reader := bytes.NewReader(whiteListOpBytes)
	if err = whiteListOp.Deserialize(reader); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsWhiteListOp DecodeBytes error!")
	}
	if whiteListOp.Op == ADD {
		err = AddRulesToList(native, whiteListOp.FileHash, whiteListOp.List.List)
	} else if whiteListOp.Op == DEL {
		err = DelRulesFromList(native, whiteListOp.FileHash, whiteListOp.List.List)
	} else if whiteListOp.Op == ADD_COV {
		err = CovRulesToList(native, whiteListOp.FileHash, whiteListOp.List.List)
	} else if whiteListOp.Op == DEL_ALL {
		err = CleRulesFromList(native, whiteListOp.FileHash)
	} else {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsWhiteListOp Op error!")
	}
	if err != nil {
		return utils.BYTE_FALSE, nil
	}
	return utils.BYTE_TRUE, nil
}

func FsGetWhiteList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	fileHash, err := utils.DecodeBytes(source)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetWhiteList DecodeBytes error!")), nil
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	whiteListKey := GenFsWhiteListKey(contract, fileHash)
	item, err := utils.GetStorageItem(native, whiteListKey)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetWhiteList GetStorageItem error!")), nil
	}
	if item == nil {
		return EncRet(false, []byte("[FS Profit] FsGetWhiteList not found!")), nil
	}
	return EncRet(true, item.Value), nil
}

func FsGetFileProveDetails(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	fileHash, err := utils.DecodeBytes(source)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileProveDetails DecodeBytes error!")), nil
	}

	proveDetails, err := getProveDetailsWithNodeAddr(native, fileHash)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileProveDetails GetProveDetails error!")), nil
	}

	bf := new(bytes.Buffer)
	err = proveDetails.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetFileProveDetails ProveDetails Serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

func FsDeleteFile(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	fileHash, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsDeleteFile DecodeBytes error!")
	}

	fileInfo, err := getFsFileInfo(native, fileHash)
	if err != nil {
		log.Debugf("FsDeleteFile: getFsFileInfo error : %s for  %s", err.Error(), string(fileHash))
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsDeleteFile getFsFileInfo error:" + err.Error())
	}

	log.Debugf("FsDeleteFile: Deposit %d for %s", fileInfo.Deposit, string(fileHash))
	err = deleteFiles(native, []*FileInfo{fileInfo})
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	DeleteFileEvent(native, fileHash, fileInfo.FileOwner)
	return utils.BYTE_TRUE, nil
}

func FsDeleteFiles(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)

	fileListBytes, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsDeleteFiles DecodeBytes error!")
	}

	var fileList FileList
	reader := bytes.NewReader(fileListBytes)
	if err = fileList.Deserialize(reader); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsDeleteFiles DecodeBytes error!")
	}

	fileInfos := make([]*FileInfo, 0, fileList.FileNum)
	fileInfoMap := make(map[string]struct{})

	for _, fileHash := range fileList.List {
		fileHashStr := (string)(fileHash.Hash)
		// skip duplicate
		if _, exist := fileInfoMap[fileHashStr]; exist {
			continue
		} else {
			fileInfo, err := getFsFileInfo(native, fileHash.Hash)
			if err != nil {
				continue
			}
			fileInfos = append(fileInfos, fileInfo)
			fileInfoMap[fileHashStr] = struct{}{}
		}
	}

	if fileList.FileNum > 0 && len(fileInfos) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsDeleteFile getFsFileInfo error:")
	}
	err = deleteFiles(native, fileInfos)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	hashes := make([]string, 0, len(fileInfos))
	for _, fi := range fileInfos {
		hashes = append(hashes, string(fi.FileHash))
	}
	DeleteFilesEvent(native, hashes, fileInfos[0].FileOwner)
	return utils.BYTE_TRUE, nil
}

func FsGetUnProvePrimaryFiles(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	walletAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetUnProvePrimaryFiles DecodeAddress error!")), nil
	}
	fileList, err := GetFsFilePrimaryList(native, walletAddr)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetUnProvePrimaryFiles GetFsFileList error!")), nil
	}
	var unProveList FileList
	for _, hash := range fileList.List {
		proveDetail, err := getProveDetails(native, hash.Hash)
		if err != nil || len(proveDetail.ProveDetails) == 0 {
			continue
		}
		prove := false
		for _, detail := range proveDetail.ProveDetails {
			log.Debugf("details %v, %v, %v", detail.WalletAddr.ToBase58(), walletAddr.ToBase58(), detail.ProveTimes)
			if detail.WalletAddr.ToBase58() == walletAddr.ToBase58() && detail.ProveTimes > 0 {
				prove = true
				break
			}
		}
		if prove {
			continue
		}
		unProveList.Add(hash.Hash)
	}
	bf := new(bytes.Buffer)
	err = unProveList.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetUnProvePrimaryFiles FileList Serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

func FsGetUnProveCandidateFiles(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	walletAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetUnProveCandidateFiles DecodeAddress error!")), nil
	}
	fileList, err := GetFsFileCandidateList(native, walletAddr)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetUnProveCandidateFiles GetFsFileList error!")), nil
	}
	var unProveList FileList
	for _, hash := range fileList.List {
		proveDetail, err := getProveDetails(native, hash.Hash)
		if err != nil || len(proveDetail.ProveDetails) == 0 {
			continue
		}
		prove := false
		for _, detail := range proveDetail.ProveDetails {
			if detail.WalletAddr.ToBase58() == walletAddr.ToBase58() && detail.ProveTimes > 0 {
				prove = true
				break
			}
		}
		if prove {
			continue
		}
		unProveList.Add(hash.Hash)
	}
	bf := new(bytes.Buffer)
	err = unProveList.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetUnProveCandidateFiles FileList Serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

func FsGetUnSettledFiles(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	walletAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetUnSettledFiles DecodeAddress error!")), nil
	}
	unsettledList, err := GetFsUnSettledList(native, walletAddr)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetUnSettledFiles GetFsFileList error!")), nil
	}
	bf := new(bytes.Buffer)
	err = unsettledList.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsGetUnSettledFiles FileList Serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

// delete unsettled files and give backup remaining deposit
func FsDeleteUnsettledFiles(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	source := common.NewZeroCopySource(native.Input)
	walletAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsDeleteUnSettledFiles DecodeAddress error!")), nil
	}

	unsettledList, err := GetFsUnSettledList(native, walletAddr)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsDeleteUnSettledFiles GetFsFileList error!")), nil
	}

	sType := []int{FileStorageTypeCustom, FileStorageTypeUseSpace}
	deletedFiles, amount, err := deleteExpiredFilesFromList(native, unsettledList, walletAddr, sType)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsDeleteUnSettledFiles deleteExpiredFilesFromList error!")), nil
	}

	log.Debugf("FsDeleteUnsettledFiles for %s, deletedFiles count %d, amount %d",
		walletAddr, len(deletedFiles), amount)

	// removed deleted files from unsettle list
	for _, fileHash := range deletedFiles {
		unsettledList.Del(fileHash)
	}

	err = setFsFileList(native, GenFsUnSettledListKey(contract, walletAddr), unsettledList)
	if err != nil {
		return EncRet(false, []byte("[FS Profit] FsDeleteUnSettledFiles setFsFileList error!")), nil
	}
	return utils.BYTE_TRUE, nil
}

func deleteExpiredFilesFromList(native *native.NativeService, fileList *FileList, walletAddr common.Address,
	storageTypes []int) ([][]byte, uint64, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fsSetting, err := getFsSetting(native)
	if err != nil {
		return nil, 0, errors.NewErr("[FS Profit] deleteExpiredFiles getFsSetting error!")
	}

	deletedFiles := make([][]byte, 0)
	height := uint64(native.Height)
	amount := uint64(0)
	for _, file := range fileList.List {
		fileInfo, err := getFsFileInfo(native, file.Hash)
		if err != nil {
			return nil, 0, errors.NewErr("[FS Profit] deleteExpiredFiles getFsFileInfo error!")
		}

		match := false
		// check if type as expected
		for _, sType := range storageTypes {
			if int(fileInfo.StorageType) == sType {
				match = true
				break
			}
		}

		if !match {
			continue
		}

		// check if file is not proved after expired for a prove interval
		if fileInfo.ExpiredHeight+fsSetting.DefaultProvePeriod > height {
			continue
		}

		amount += fileInfo.Deposit
		cleanupForDeleteFile(native, fileInfo, true, true)
		deletedFiles = append(deletedFiles, file.Hash)
	}

	if amount > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract, walletAddr, amount)
		if err != nil {
			return nil, 0, errors.NewErr("[FS Profit] FsDeleteUnSettledFiles AppCallTransfer error error!")
		}
	}

	return deletedFiles, amount, nil
}

func deleteFiles(native *native.NativeService, fileInfos []*FileInfo) error {
	if len(fileInfos) == 0 {
		return nil
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	refundAmount := uint64(0)
	fileOwner := fileInfos[0].FileOwner

	if !native.ContextRef.CheckWitness(fileOwner) {
		return errors.NewErr("[FS Profit] FsDeleteFile CheckWitness failed!")
	}

	fsSetting, err := getFsSetting(native)
	if err != nil {
		return errors.NewErr("[FS Profit] FsDeleteFile getFsSetting error!")
	}

	for _, fileInfo := range fileInfos {
		if fileInfo == nil {
			return errors.NewErr("[FS Profit] FsDeleteFile fileInfo is nil")
		}
		if fileInfo.FileOwner.ToBase58() != fileOwner.ToBase58() {
			return errors.NewErr("[FS Profit] FsDeleteFile file owner are different!")
		}
	}

	for _, fileInfo := range fileInfos {
		fileHash := fileInfo.FileHash

		for _, sectorRef := range fileInfo.SectorRefs {
			sectorInfo, err := getSectorInfo(native, sectorRef.NodeAddr, sectorRef.SectorID)
			if err != nil {
				return errors.NewErr("[FS Profit] FsDeleteFile getSectorInfo error!")
			}

			err = deleteFileFromSector(native, sectorInfo, fileInfo)
			if err != nil {
				return errors.NewErr("[FS Profit] FsDeleteFile deleteFileFromSector error!")
			}
		}

		if fileInfo.Deposit == 0 {
			cleanupForDeleteFile(native, fileInfo, true, true)
			continue
		}

		// fileInfo.deposit > 0
		fileProveDetails, err := getProveDetails(native, fileHash)
		if err != nil {
			return errors.NewErr("[FS Profit] FsDeleteFile GetProveDetails failed!")
		}

		restProfit := fileInfo.Deposit
		fileSize := fileInfo.FileBlockNum * fileInfo.FileBlockSize
		singleProveProfit := calcSingleValidFeeForFile(fsSetting, fileSize)

		for i := 0; uint64(i) < fileProveDetails.ProveDetailNum; i++ {
			fsNodeInfo, err := getFsNodeInfo(native, fileProveDetails.ProveDetails[i].WalletAddr)
			if err != nil {
				return errors.NewErr("[FS Profit] FsDeleteFile GetFsNodeInfo error!")
			}

			validProfit := (fileProveDetails.ProveDetails[i].ProveTimes - 1) * singleProveProfit
			storageProfit := calcStorageFeeForOneNode(fsSetting, fileSize, uint64(native.Height)-fileInfo.BlockHeight)

			totalProfit := validProfit + storageProfit

			if totalProfit >= restProfit {
				return errors.NewErr("[FS Profit] FsDeleteFile invalid profit")
			}

			fsNodeInfo.Profit += totalProfit

			if err = setFsNodeInfo(native, fsNodeInfo); err != nil {
				return errors.NewErr("[FS Profit] FsDeleteFile setFsNodeInfo error:" + err.Error())
			}

			restProfit -= totalProfit
		}
		if fileInfo.StorageType == FileStorageTypeCustom {
			//give back remaining profit
			refundAmount += restProfit
		} else if fileInfo.StorageType == FileStorageTypeUseSpace {
			userspace, err := getUserSpace(native, fileInfo.FileOwner)
			if err != nil {
				return errors.NewErr("[FS Profit] FsDeleteFile GetUserSpace error!")
			}
			if userspace.Used >= fileInfo.FileBlockNum*fileInfo.FileBlockSize {
				userspace.Balance += restProfit
				userspace.Remain += fileInfo.FileBlockNum * fileInfo.FileBlockSize
				userspace.Used -= fileInfo.FileBlockNum * fileInfo.FileBlockSize
			} else {
				log.Errorf("used is less than size %d %d", userspace.Used, fileInfo.FileBlockNum, fileInfo.FileBlockSize)
				return errors.NewErr("[FS Profit] FsDeleteFile userspace value error!")
			}
			if err = setUserSpace(native, userspace, fileInfo.FileOwner); err != nil {
				return errors.NewErr("[FS Profit] FsDeleteFile SetUserSpace error!")
			}
		}
		cleanupForDeleteFile(native, fileInfo, true, true)
	}

	if refundAmount == 0 {
		return nil
	}
	err = appCallTransfer(native, utils.UsdtContractAddress, contract, fileOwner, refundAmount)
	if err != nil {
		return errors.NewErr("[FS Profit] AppCallTransfer, transfer error!")
	}

	return nil
}

func cleanupForDeleteFile(native *native.NativeService, fileInfo *FileInfo, rmInfo bool, rmList bool) {
	fileHash := fileInfo.FileHash

	if rmInfo {
		deleteFsFileInfo(native, fileHash)
		deleteProveDetails(native, fileHash)
	}

	if rmList {
		DelFileFromList(native, fileInfo.FileOwner, fileInfo.FileHash)
		for _, primaryWalletAddr := range fileInfo.PrimaryNodes.AddrList {
			DelFileFromPrimaryList(native, primaryWalletAddr, fileInfo.FileHash)
		}
		for _, candidateWalletAddr := range fileInfo.CandidateNodes.AddrList {
			DelFileFromCandidateList(native, candidateWalletAddr, fileInfo.FileHash)
		}
	}
}

func FsChangeFileOwner(native *native.NativeService) ([]byte, error) {
	var ownerChange OwnerChange
	source := common.NewZeroCopySource(native.Input)
	err := ownerChange.Deserialization(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsChangeFileOwner OwnerChange Deserialization error!")
	}

	if !native.ContextRef.CheckWitness(ownerChange.CurOwner) {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}

	fileInfo, err := getFsFileInfo(native, ownerChange.FileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsChangeFileOwner GetFsFileInfo error!")
	}
	if fileInfo.FileOwner != ownerChange.CurOwner {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsChangeFileOwner Caller is not file's owner!")
	}
	fileInfo.FileOwner = ownerChange.NewOwner

	if err = setFsFileInfo(native, fileInfo); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsChangeFileOwner setFsFileInfo error:" + err.Error())
	}

	if err = DelFileFromList(native, ownerChange.CurOwner, ownerChange.FileHash); err != nil {
		return utils.BYTE_FALSE, err
	}
	if err = AddFileToList(native, ownerChange.NewOwner, ownerChange.FileHash); err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func FsChangeFilePrivilege(native *native.NativeService) ([]byte, error) {
	var priChange PriChange
	source := common.NewZeroCopySource(native.Input)
	err := priChange.Deserialization(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsChangeFilePrivilege PriChange Deserialization error!")
	}

	fileInfo, err := getFsFileInfo(native, priChange.FileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsChangeFilePrivilege GetFsFileInfo error!")
	}

	if !native.ContextRef.CheckWitness(fileInfo.FileOwner) {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	fileInfo.Privilege = priChange.Privilege

	if err = setFsFileInfo(native, fileInfo); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Profit] FsChangeFilePrivilege setFsFileInfo error:" + err.Error())
	}
	return utils.BYTE_TRUE, nil
}
