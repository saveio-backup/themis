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

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/savefs/pdp"
	"github.com/saveio/themis/smartcontract/service/native/usdt"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func FsNodeRegister(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var fsNodeInfo FsNodeInfo
	infoSource := common.NewZeroCopySource(native.Input)
	if err := fsNodeInfo.Deserialization(infoSource); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeInfo deserialize error!")
	}

	if !native.ContextRef.CheckWitness(fsNodeInfo.WalletAddr) {
		return utils.BYTE_FALSE, errors.NewErr("FS Govern] CheckWitness failed!")
	}

	fsNodeInfoKey := GenFsNodeInfoKey(contract, fsNodeInfo.WalletAddr)
	item, err := utils.GetStorageItem(native, fsNodeInfoKey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] GetStorageItem error!")
	}
	if item != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] Node have registered!")
	}

	fsSetting, err := getFsSetting(native)
	if err != nil || fsSetting == nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] GetFsSetting error!")
	}

	if fsNodeInfo.Volume < fsSetting.MinVolume {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] Volume < MinVolume!")
	}

	pledge := fsSetting.FsGasPrice * fsSetting.GasPerGBPerBlock * fsNodeInfo.Volume
	err = appCallTransfer(native, utils.UsdtContractAddress, fsNodeInfo.WalletAddr, contract, pledge)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] appCallTransfer, transfer error!")
	}

	fsNodeInfo.Pledge = pledge
	fsNodeInfo.Profit = 0
	fsNodeInfo.RestVol = fsNodeInfo.Volume

	info := new(bytes.Buffer)
	err = fsNodeInfo.Serialize(info)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeInfo serialize error!")
	}
	utils.PutBytes(native, fsNodeInfoKey, info.Bytes())

	err = nodeListOperate(native, fsNodeInfo.WalletAddr, true)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] NodeListOperate add error!")
	}
	RegisterNodeEvent(native, fsNodeInfo.WalletAddr, fsNodeInfo.NodeAddr, fsNodeInfo.Volume, fsNodeInfo.ServiceTime)
	return utils.BYTE_TRUE, nil
}

func FsNodeQuery(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)

	walletAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[FS Govern] DecodeAddress error!")), nil
	}

	fsNodeInfo, err := getFsNodeInfo(native, walletAddr)
	if err != nil {
		return EncRet(false, []byte("[FS Govern] FsNodeQuery getFsNodeInfo error!")), nil
	}

	info := new(bytes.Buffer)
	fsNodeInfo.Serialize(info)
	return EncRet(true, info.Bytes()), nil
}

func FsNodeUpdate(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var newFsNodeInfo FsNodeInfo
	newInfoSource := common.NewZeroCopySource(native.Input)
	if err := newFsNodeInfo.Deserialization(newInfoSource); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeInfo deserialize error!")
	}

	fsSetting, err := getFsSetting(native)
	if err != nil || fsSetting == nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] GetFsSetting error!")
	}

	if newFsNodeInfo.Volume < fsSetting.MinVolume {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] Volume < MinVolume!")
	}

	if !native.ContextRef.CheckWitness(newFsNodeInfo.WalletAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] CheckWitness failed!")
	}

	oldFsNodeInfo, err := getFsNodeInfo(native, newFsNodeInfo.WalletAddr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeUpdate getFsNodeInfo error!")
	}

	newPledge := fsSetting.FsGasPrice * fsSetting.GasPerGBPerBlock * newFsNodeInfo.Volume
	if newFsNodeInfo.WalletAddr != oldFsNodeInfo.WalletAddr {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeInfo walletAddr changed!")
	}

	var state usdt.State
	if newPledge < oldFsNodeInfo.Pledge {
		state = usdt.State{From: contract, To: newFsNodeInfo.WalletAddr, Value: oldFsNodeInfo.Pledge - newPledge}
	} else if newPledge > oldFsNodeInfo.Pledge {
		state = usdt.State{From: newFsNodeInfo.WalletAddr, To: contract, Value: newPledge - oldFsNodeInfo.Pledge}
	}
	if newPledge != oldFsNodeInfo.Pledge {
		err = appCallTransfer(native, utils.UsdtContractAddress, state.From, state.To, state.Value)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] appCallTransfer, transfer error!")
		}
	}

	newFsNodeInfo.Pledge = newPledge
	newFsNodeInfo.Profit = oldFsNodeInfo.Profit
	newFsNodeInfo.RestVol = oldFsNodeInfo.RestVol + newFsNodeInfo.Volume - oldFsNodeInfo.Volume

	if err = setFsNodeInfo(native, &newFsNodeInfo); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] setFsNodeInfo error:" + err.Error())
	}
	return utils.BYTE_TRUE, nil
}

func FsNodeCancel(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	source := common.NewZeroCopySource(native.Input)
	addr, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeCancel DecodeAddress error!")
	}

	if !native.ContextRef.CheckWitness(addr) {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] CheckWitness failed!")
	}

	fsNodeInfo, err := getFsNodeInfo(native, addr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeCancel getFsNodeInfo error!")
	}

	if fsNodeInfo.Pledge > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract, fsNodeInfo.WalletAddr, fsNodeInfo.Pledge+fsNodeInfo.Profit)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeCancel appCallTransfer,  transfer error!")
		}
	}

	fsNodeInfoKey := GenFsNodeInfoKey(contract, addr)
	utils.DelStorageItem(native, fsNodeInfoKey)

	err = nodeListOperate(native, fsNodeInfo.WalletAddr, false)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeCancel NodeListOperate delete error!")
	}

	sectorInfos, err := getSectorsForNode(native, addr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeCancel getSectorsForNode error!")
	}

	for _, sectorInfo := range sectorInfos.Sectors {
		err = deleteSector(native, addr, sectorInfo.SectorID)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeCancel deleteSectorInfo error!")
		}
	}

	UnRegisterNodeEvent(native, fsNodeInfo.WalletAddr)
	return utils.BYTE_TRUE, nil
}

func FsNodeWithDrawProfit(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	source := common.NewZeroCopySource(native.Input)
	addr, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeWithDrawProfit DecodeAddress error!")
	}

	if !native.ContextRef.CheckWitness(addr) {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] CheckWitness failed!")
	}

	fsNodeInfo, err := getFsNodeInfo(native, addr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeWithDrawProfit getFsNodeInfo error!")
	}

	if fsNodeInfo.Profit > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract, fsNodeInfo.WalletAddr, fsNodeInfo.Profit)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeCancel appCallTransfer,  transfer error!")
		}
		fsNodeInfo.Profit = 0
	} else {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeWithDrawProfit profit <= 0 error! ")
	}

	if err = setFsNodeInfo(native, fsNodeInfo); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsNodeWithDrawProfit setFsNodeInfo error:" + err.Error())
	}

	return utils.BYTE_TRUE, nil
}

func FsFileProve(native *native.NativeService) ([]byte, error) {
	var fileProve FileProve
	source := common.NewZeroCopySource(native.Input)
	if err := fileProve.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FileProve deserialize error!")
	}
	if !native.ContextRef.CheckWitness(fileProve.NodeWallet) {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] CheckWitness failed!")
	}

	fsSetting, err := getFsSetting(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FileProve GetFsFileInfo error!")
	}

	fileInfo, err := getFsFileInfo(native, fileProve.FileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FileProve GetFsFileInfo error!")
	}

	// check if the node can prove the file
	canProve := false
	for _, primaryNode := range fileInfo.PrimaryNodes.AddrList {
		if primaryNode.ToBase58() == fileProve.NodeWallet.ToBase58() {
			canProve = true
			break
		}
	}
	if !canProve {
		// no in primary node list, check candidate list
		for _, candidateNode := range fileInfo.CandidateNodes.AddrList {
			if candidateNode.ToBase58() == fileProve.NodeWallet.ToBase58() {
				canProve = true
				break
			}
		}
	}
	if !canProve {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FileProve No in prove node list error!")
	}

	nodeInfo, err := getFsNodeInfo(native, fileProve.NodeWallet)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve GetFsNodeInfo error!")
	}

	proveDetails, err := getProveDetails(native, fileInfo.FileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve GetProveDetails error!")
	}

	if fileProve.SectorID != 0 && uint64(native.Height) < fileInfo.ExpiredHeight {
		for i := 0; uint64(i) < proveDetails.ProveDetailNum; i++ {
			if proveDetails.ProveDetails[i].WalletAddr == fileProve.NodeWallet {
				log.Errorf("[FS Govern] Should prove by sector!")
				return utils.BYTE_FALSE, errors.NewErr("[FS Govern] Should prove by sector!")
			}
		}
	}

	ret, err := checkProve(native, &fileProve, fileInfo)
	if err != nil {
		log.Debugf("check prove error %v for file %s", err, string(fileProve.FileHash))
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] CheckProve error!")
	}
	if !ret {
		log.Debugf("check prove ret %v for file %s", ret, string(fileProve.FileHash))
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] ProveData Verify failed!")
	}

	log.Debugf("check prove success for file %s", string(fileProve.FileHash))
	found := false
	settleFlag := false
	haveProveTimes := uint64(0)

	var proveDetail *ProveDetail

	fileExpiredHeight := fileInfo.ExpiredHeight
	for i := 0; uint64(i) < proveDetails.ProveDetailNum; i++ {
		if proveDetails.ProveDetails[i].WalletAddr == fileProve.NodeWallet {
			proveDetail = &proveDetails.ProveDetails[i]

			haveProveTimes = proveDetail.ProveTimes
			firstProveHeight := proveDetail.BlockHeight
			if haveProveTimes == fileInfo.ProveTimes || uint64(native.Height) > fileExpiredHeight {
				proveDetail.Finished = true
				settleFlag = true
			}
			if haveProveTimes > fileInfo.ProveTimes {
				return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve Prove times reached limit!")
			}
			if !checkProveExpire(native, haveProveTimes, fileInfo.ProveInterval, firstProveHeight, fileExpiredHeight) {
				return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve Prove out of date!")
			}
			proveDetail.ProveTimes++
			found = true
			break
		}
	}

	//The first file prove only indicate node has store the file.
	if !found {
		if proveDetails.ProveDetailNum == fileInfo.CopyNum+1 {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve already have enough nodes!")
		}

		nodeInfo.RestVol -= fileInfo.FileBlockNum * fileInfo.FileBlockSize
		if err := setFsNodeInfo(native, nodeInfo); err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve setFsNodeInfo error:" + err.Error())
		}

		// prove detail record the height for first file prove
		proveDetail = &ProveDetail{nodeInfo.NodeAddr, nodeInfo.WalletAddr, 1, uint64(native.Height), false}
		proveDetails.ProveDetails = append(proveDetails.ProveDetails, *proveDetail)
		proveDetails.ProveDetailNum++
	}

	if err = setProveDetails(native, fileInfo.FileHash, proveDetails); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] ProveDetails setProveDetails error!")
	}

	if !found {
		// first prove, add file to sector
		sectorInfo, err := getSectorInfo(native, fileProve.NodeWallet, fileProve.SectorID)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve Sector not found")
		}

		err = addFileToSector(native, sectorInfo, fileInfo)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve addFileToSector error:" + err.Error())
		}

		err = addSectorRefForFileInfo(native, fileInfo, sectorInfo)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve addSectorRefForFileInfo err:" + err.Error())
		}

		// update next challenge height for sector prove if it is the first file added to the sector
		if sectorInfo.NextProveHeight == 0 {
			sectorInfo.NextProveHeight = fileProve.BlockHeight + fileInfo.ProveInterval
		}

		err = setSectorInfo(native, sectorInfo)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve secSectorInfo err:" + err.Error())
		}
	}

	//transfer profit
	if settleFlag {
		if fileProve.SectorID != 0 {
			sectorInfo, err := getSectorInfo(native, fileProve.NodeWallet, fileProve.SectorID)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewErr("[FS Govern] GetSectorInfo error:" + err.Error())
			}

			err = deleteFileFromSector(native, sectorInfo, fileInfo)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewErr("[FS Govern] DeleteFileFromSector error:" + err.Error())
			}
		}

		err := settleForFile(native, fileInfo, nodeInfo, proveDetail, proveDetails, fsSetting)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve SettleForFile error:" + err.Error())
		}
	}

	FilePDPSuccessEvent(native, fileInfo.FileHash, nodeInfo.WalletAddr)
	return utils.BYTE_TRUE, nil
}

func punishBrokenNode(native *native.NativeService, bakAddr common.Address, brokenAddr common.Address, amount uint64) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	brokenNodeInfo, err := getFsNodeInfo(native, brokenAddr)
	if err != nil {
		return fmt.Errorf("PunishBrokenNode GetFsNodeInfo error: %s", err.Error())
	}
	brokenNodeInfo.Pledge -= amount

	if err = setFsNodeInfo(native, brokenNodeInfo); err != nil {
		return fmt.Errorf("PunishBrokenNode SetFsNodeInfo error: %s", err.Error())
	}

	err = appCallTransfer(native, utils.UsdtContractAddress, contract, bakAddr, amount)
	return err
}

func checkProveExpire(native *native.NativeService, haveProvedTimes, ProveInterval, fileBlockHeight, fileExpiredHeight uint64) bool {
	currBlockHeight := uint64(native.Height)
	// no periodic file prove when have sector prove, just check for last prove
	if currBlockHeight > fileExpiredHeight {
		return true
	}
	return false
}

func nodeListOperate(native *native.NativeService, walletAddr common.Address, isAdd bool) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	nodeSetKey := GenFsNodeSetKey(contract)
	nodeSet, err := utils.GetStorageItem(native, nodeSetKey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] GetStorageItem nodeSetKey error!")
	}

	var nodeList NodeList
	if nodeSet != nil {
		reader := bytes.NewReader(nodeSet.Value)
		if err = nodeList.Deserialize(reader); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] Set deserialize error!")
		}
	}

	if isAdd {
		nodeList.Add(walletAddr)
	} else {
		nodeList.Del(walletAddr)
	}
	bf := new(bytes.Buffer)
	err = nodeList.Serialize(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] Put node to set error!")
	}
	utils.PutBytes(native, nodeSetKey, bf.Bytes())
	return nil
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

func getFsNodeInfo(native *native.NativeService, walletAddr common.Address) (*FsNodeInfo, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	fsNodeInfoKey := GenFsNodeInfoKey(contract, walletAddr)
	item, err := utils.GetStorageItem(native, fsNodeInfoKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] FsNodeInfo GetStorageItem error!")
	}
	if item == nil {
		return nil, errors.NewErr("[FS Govern] FsNodeInfo not found!")
	}
	var fsNodeInfo FsNodeInfo
	fsNodeInfoSource := common.NewZeroCopySource(item.Value)
	err = fsNodeInfo.Deserialization(fsNodeInfoSource)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] FsNodeInfo deserialize error!")
	}
	return &fsNodeInfo, nil
}

func setFsNodeInfo(native *native.NativeService, fsNodeInfo *FsNodeInfo) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	fsNodeInfoKey := GenFsNodeInfoKey(contract, fsNodeInfo.WalletAddr)
	info := new(bytes.Buffer)
	if err := fsNodeInfo.Serialize(info); err != nil {
		return errors.NewErr("[FS Govern] FsNodeInfo serialize error!")
	}
	utils.PutBytes(native, fsNodeInfoKey, info.Bytes())
	return nil
}

func getFsNodeList(native *native.NativeService) (*NodeList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	nodeSetKey := GenFsNodeSetKey(contract)
	nodeSet, err := utils.GetStorageItem(native, nodeSetKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] GetStorageItem nodeSetKey error!")
	}
	if nodeSet == nil {
		return nil, errors.NewErr("[FS Govern] FsGetNodeList No nodeSet found!")
	}

	var nodeList NodeList
	reader := bytes.NewReader(nodeSet.Value)
	if err = nodeList.Deserialize(reader); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] Set deserialize error!")
	}
	return &nodeList, nil
}

func getProveDetails(native *native.NativeService, fileHash []byte) (*FsProveDetails, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	proveDetailsKey := GenFsProveDetailsKey(contract, fileHash)
	item, err := utils.GetStorageItem(native, proveDetailsKey)
	if err != nil || item == nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] FsFileProve GetStorageItem error!")
	}
	var proveDetails FsProveDetails
	reader := bytes.NewReader(item.Value)
	if err = proveDetails.Deserialize(reader); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Govern] FsFileProve ProveDetails deserialize error!")
	}

	for i := uint64(0); i < proveDetails.ProveDetailNum; i++ {
		nodeInfo, err := getFsNodeInfo(native, proveDetails.ProveDetails[i].WalletAddr)
		if err != nil {
			return nil, errors.NewErr("[FS Govern] GetProveDetails GetFsNodeInfo error!")
		}
		proveDetails.ProveDetails[i].NodeAddr = nodeInfo.NodeAddr
	}
	return &proveDetails, nil
}

func setProveDetails(native *native.NativeService, fileHash []byte, proveDetails *FsProveDetails) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	proveDetailsKey := GenFsProveDetailsKey(contract, fileHash)
	proveDetailsBuff := new(bytes.Buffer)
	if err := proveDetails.Serialize(proveDetailsBuff); err != nil {
		return errors.NewErr("[FS Govern] ProveDetails serialize error!")
	}
	utils.PutBytes(native, proveDetailsKey, proveDetailsBuff.Bytes())
	return nil
}

func checkProve(native *native.NativeService, fileProve *FileProve, fileInfo *FileInfo) (bool, error) {
	pp, err := getProveParam(fileInfo.FileProveParam)
	if err != nil {
		return false, errors.NewErr("[FS Govern] ProveParam deserialize error!")
	}

	header, err := native.Store.GetHeaderByHeight(uint32(fileProve.BlockHeight))
	if err != nil {
		return false, err
	}

	currBlockHeight := uint64(native.Height)
	if header == nil {
		log.Errorf("header is nil of blockheight %d, current height:%d", fileProve.BlockHeight, currBlockHeight)
		return false, errors.NewErr("[FS Govern] block header is nil!")
	}

	if fileProve.BlockHeight > currBlockHeight+fileInfo.ProveInterval || fileProve.BlockHeight+fileInfo.ProveInterval < currBlockHeight {
		log.Errorf("invalid prove blockheight %d, current height:%d, prove interval:%d", fileProve.BlockHeight, currBlockHeight, fileInfo.ProveInterval)
		return false, errors.NewErr("[FS Govern] invalid prove blockheight!")
	}

	blockHash := header.Hash()
	challenge := GenChallenge(fileProve.NodeWallet, blockHash, uint32(fileInfo.FileBlockNum), uint32(fileInfo.ProveBlockNum))

	var pd ProveData
	pdReader := bytes.NewReader(fileProve.ProveData)
	err = pd.Deserialize(pdReader)
	if err != nil {
		return false, errors.NewErr("[FS Govern] ProveData deserialize error!")
	}

	p := pdp.NewPdp(0)
	err = p.VerifyProofWithMerklePathForFile(0, pd.Proofs, pp.FileID, pd.Tags, challenge, pd.MerklePath, pp.RootHash)
	if err != nil {
		return false, errors.NewErr("[FS Govern] ProveData Verify failed!")
	}
	return true, nil
}

func calculateProfitForSettle(fileInfo *FileInfo, proveDetail *ProveDetail, fsSetting *FsSetting) uint64 {
	// first prove just indicate the whole file has been uploaded and dont calc for profit
	// copyNum pass 0 to calculate total fee for one node
	total := calcFee(fsSetting, proveDetail.ProveTimes-1, 0, fileInfo.FileBlockNum*fileInfo.FileBlockSize, fileInfo.ExpiredHeight-fileInfo.BlockHeight)
	log.Debugf("prove times: %d, block num: %d, block size: %d, expire height : %d, block height : %d, valid fee: %d, storage fee : %d\n",
		proveDetail.ProveTimes, fileInfo.FileBlockNum, fileInfo.FileBlockSize, fileInfo.ExpiredHeight, fileInfo.BlockHeight, total.ValidationFee, total.SpaceFee)

	return total.ValidationFee + total.SpaceFee
}
