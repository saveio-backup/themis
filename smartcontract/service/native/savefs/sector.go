package savefs

import (
	"bytes"
	"fmt"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/savefs/pdp"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func FsCreateSector(native *native.NativeService) ([]byte, error) {
	var sectorInfo SectorInfo
	source := common.NewZeroCopySource(native.Input)
	if err := sectorInfo.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] SectorInfo deserialize error!")
	}

	if !native.ContextRef.CheckWitness(sectorInfo.NodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] authentication failed!")
	}

	nodeInfo, err := getFsNodeInfo(native, sectorInfo.NodeAddr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] NodeInfo not found!")
	}

	if sectorInfo.SectorID == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector]  sector id is 0!")
	}

	if sectorInfo.Size < MIN_SECTOR_SIZE {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] sector size is smaller than min sector size!")
	}

	totalSize, err := getSectorTotalSizeForNode(native, sectorInfo.NodeAddr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] get total size for sectors error!")
	}

	if sectorInfo.Size+totalSize > nodeInfo.Volume {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] total size for sectors larger than node volume!")
	}

	switch sectorInfo.ProveLevel {
	case PROVE_LEVEL_HIGH:
	case PROVE_LEVEL_MEDIEUM:
	case PROVE_LEVEL_LOW:
	default:
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] invalid prove level!")
	}

	_, err = getSectorInfo(native, sectorInfo.NodeAddr, sectorInfo.SectorID)
	if err == nil {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] sector already exist")
	}

	info := &SectorInfo{
		NodeAddr:   sectorInfo.NodeAddr,
		SectorID:   sectorInfo.SectorID,
		ProveLevel: sectorInfo.ProveLevel,
		Size:       sectorInfo.Size,
		IsPlots:    sectorInfo.IsPlots,
	}

	err = setSectorInfo(native, info)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] setSectorInfo error!")
	}

	CreateSectorEvent(native, info.NodeAddr, info.SectorID, info.ProveLevel, info.Size, info.IsPlots)
	return utils.BYTE_TRUE, nil
}

func FsGetSectorInfo(native *native.NativeService) ([]byte, error) {
	var sectorRef SectorRef
	source := common.NewZeroCopySource(native.Input)
	if err := sectorRef.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[SectorInfo] SectorRef deserialize error!")
	}

	sectorInfo, err := getSectorInfoWithFileList(native, sectorRef.NodeAddr, sectorRef.SectorID)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[GetSectorInfo] GetSectorInfo error!")
	}

	sink := common.NewZeroCopySink(nil)
	sectorInfo.Serialization(sink)

	return EncRet(true, sink.Bytes()), nil
}

func FsDeleteSector(native *native.NativeService) ([]byte, error) {
	var sectorRef SectorRef
	source := common.NewZeroCopySource(native.Input)
	if err := sectorRef.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteSector] SectorRef deserialize error!")
	}

	if !native.ContextRef.CheckWitness(sectorRef.NodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteSector] authentication failed!")
	}

	sectorInfo, err := getSectorInfo(native, sectorRef.NodeAddr, sectorRef.SectorID)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteSector] sector not found!")
	}

	if getSectorFileNum(sectorInfo) != 0 {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteSector] Cannot delete a sector with file")
	}

	err = deleteSector(native, sectorRef.NodeAddr, sectorRef.SectorID)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteSector] deleteSectorInfo error!")
	}

	DeleteSectorEvent(native, sectorRef.NodeAddr, sectorRef.SectorID)
	return utils.BYTE_TRUE, nil
}

// delete a file in a sector by fs node, need this function?
func FsDeleteFileInSector(native *native.NativeService) ([]byte, error) {
	var sectorFileRef SectorFileRef
	source := common.NewZeroCopySource(native.Input)
	if err := sectorFileRef.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteFileInSector] SectorRef deserialize error!")
	}

	if !native.ContextRef.CheckWitness(sectorFileRef.NodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteFileInSector] authentication failed!")
	}

	fileInfo, err := getFsFileInfo(native, sectorFileRef.FileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteFileInSector] fileInfo not found!")
	}

	sectorInfo, err := getSectorInfo(native, sectorFileRef.NodeAddr, sectorFileRef.SectorID)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteFileInSector] sector not found!")
	}

	fileInSector, err := isFileInSector(native, sectorFileRef.NodeAddr, sectorFileRef.SectorID, sectorFileRef.FileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteFileInSector] check if file in sector error!")
	}

	if !fileInSector {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteFileInSector] file not in sector!")
	}

	err = deleteFileFromSector(native, sectorInfo, fileInfo)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DeleteFileInSector] delete file from sector failed!")
	}
	return utils.BYTE_TRUE, nil
}

func FsGetSectorsForNode(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	nodeAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[GetAllSectors] NodeAddr deserialize error!")
	}

	sectorInfos, err := getSectorsWithFileListForNode(native, nodeAddr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[GetAllSectors] getSectorsForNode error!")
	}

	buf := new(bytes.Buffer)
	if err := sectorInfos.Serialize(buf); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[GetAllSectors] sectorInfos Serialize error!")
	}
	return EncRet(true, buf.Bytes()), nil
}

func FsSectorProve(native *native.NativeService) ([]byte, error) {
	var sectorProve SectorProve
	source := common.NewZeroCopySource(native.Input)
	if err := sectorProve.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[SectorProve] SectorProve deserialize error!")
	}
	if !native.ContextRef.CheckWitness(sectorProve.NodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[SectorProve] CheckWitness failed!")
	}

	nodeInfo, err := getFsNodeInfo(native, sectorProve.NodeAddr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CreateSector] NodeInfo not found!")
	}

	sectorInfo, err := getSectorInfo(native, sectorProve.NodeAddr, sectorProve.SectorID)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[SectorProve] Sector not exist!")
	}

	fsSetting, err := getFsSettingWithProveLevel(native, sectorInfo.ProveLevel)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[SectorProve] getFsSettingWithProveLevel error!")
	}

	if uint64(native.Height) < sectorInfo.NextProveHeight {
		log.Error("[SectorProve] current height %d is smaller than nextProveHeight %d!", native.Height, sectorInfo.NextProveHeight)
		return utils.BYTE_FALSE, errors.NewErr("[SectorProve] current height is smaller than nextProveHeight!")
	}

	if sectorProve.ChallengeHeight != sectorInfo.NextProveHeight {
		log.Errorf("[SectorProve] challengeHeight %d in sectorProve is not the nextProveHeight %d",
			sectorProve.ChallengeHeight, sectorInfo.NextProveHeight)
		return utils.BYTE_FALSE, errors.NewErr("[SectorProve] challengeHeight in sectorProve is not the nextProveHeight")
	}

	ret, err := checkSectorProve(native, &sectorProve, sectorInfo)
	if err != nil {
		log.Errorf("checkSectorProve error %s", err)
		return utils.BYTE_FALSE, errors.NewErr("[SectorProve] CheckSectorProve error!")
	}

	if !ret {
		log.Errorf("checkSectorProve not success")
		err = punishForSector(native, sectorInfo, nodeInfo, fsSetting, 1)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[SectorProve] PunishForSector error!")
		}
		// NOTE: if not return BYTE_TRUE and no error, the db operations will not be committed to ledger
		return utils.BYTE_TRUE, nil
	}

	log.Debugf("checkSectorProve success for sector %d", sectorInfo.SectorID)

	// add profit for the node
	err = profitSplitForSector(native, sectorInfo, nodeInfo, fsSetting)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[SectorProve] updateProfitForSector error!")
	}

	if sectorInfo.FirstProveHeight == 0 {
		sectorInfo.FirstProveHeight = uint64(native.Height)
	}

	sectorInfo.NextProveHeight = uint64(native.Height) + fsSetting.DefaultProvePeriod
	err = setSectorInfo(native, sectorInfo)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[SectorProve] updateNextProveHeight error!")
	}

	if !sectorInfo.IsPlots {
		return utils.BYTE_TRUE, nil
	}

	// TODO: verify this sector is real a sector with plots
	// TODO: verify the size is plot file size
	pocProve := getPocProve(native, sectorInfo.NodeAddr, native.Height)
	if pocProve == nil {
		pocProve = &PocProve{}
	}
	pocProve.Height = native.Height
	pocProve.Miner = sectorInfo.NodeAddr
	pocProve.PlotSize += sectorInfo.Used
	if err := putPocProve(native, pocProve); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[SectorProve] putPocProve error! %s", err)
	}
	return utils.BYTE_TRUE, nil
}

// check if node has submitted sector prove in time, if not sector will be punished
func FsCheckNodeSectorProvedInTime(native *native.NativeService) ([]byte, error) {
	var sectorRef SectorRef
	source := common.NewZeroCopySource(native.Input)
	if err := sectorRef.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[SectorInfo] SectorRef deserialize error!")
	}

	nodeAddr := sectorRef.NodeAddr
	sectorID := sectorRef.SectorID

	nodeInfo, err := getFsNodeInfo(native, nodeAddr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved] NodeInfo not found!")
	}

	if nodeInfo.ServiceTime < uint64(native.Time) {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved] NodeInfo service time expired!")
	}

	if sectorRef.SectorID == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved]  sector id is 0!")
	}

	sectorInfo, err := getSectorInfo(native, nodeAddr, sectorID)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved] sector not exist!")
	}

	// sector with no files no need to check
	if sectorInfo.FileNum == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved] sector has no file!")
	}

	fsSetting, err := getFsSettingWithProveLevel(native, sectorInfo.ProveLevel)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved] getFsSetting error!")
	}

	height := uint64(native.Height)
	if sectorInfo.NextProveHeight+fsSetting.DefaultProvePeriod < height {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved] sector prove not expire!")
	}

	// check if sector has been punished already, should not punish again
	// punish node for sector, may give reward to contract caller
	lastHeight, err := getLastPunishmentHeightForNode(native, nodeAddr, sectorID)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved] getLastPunishmentHeight error!")
	}

	log.Debugf("last punishment height for node %s sector %d is %d", nodeAddr, sectorID, lastHeight)

	// punish should take times of missing sector prove into account
	times := calMissingSectorProveTimes(sectorInfo, fsSetting, lastHeight, height)
	if times == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved] node has been punished")
	}

	if err = punishForSector(native, sectorInfo, nodeInfo, fsSetting, times); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[CheckSectorProved] punish for sector error!")
	}

	return utils.BYTE_TRUE, nil
}

func checkSectorProve(native *native.NativeService, sectorProve *SectorProve, sectorInfo *SectorInfo) (bool, error) {
	var sectorProveData SectorProveData
	reader := bytes.NewReader(sectorProve.ProveData)
	err := sectorProveData.Deserialize(reader)
	if err != nil {
		log.Errorf("[SectorProve] SectorProveData deserialize error %s", err)
		return false, errors.NewErr("[SectorProve] SectorProveData deserialize error!")
	}

	header, err := native.Store.GetHeaderByHeight(uint32(sectorProve.ChallengeHeight))
	if err != nil {
		return false, err
	}

	currBlockHeight := uint64(native.Height)
	if header == nil {
		log.Errorf("header is nil of blockheight %d, current height:%d", sectorProve.ChallengeHeight, currBlockHeight)
		return false, errors.NewErr("[SectorProve] block header is nil!")
	}

	blockHash := header.Hash()

	challenges := GenChallenge(sectorProve.NodeAddr, blockHash, uint32(sectorInfo.TotalBlockNum), SECTOR_PROVE_BLOCK_NUM)

	if uint64(len(challenges)) != sectorProveData.BlockNum {
		return false, errors.NewErr("[SectorProve] length of challenges not same with the block num in sectorProve")
	}

	verifier := pdp.NewPdp(0)

	fileIDs, tags, updatedChal, path, rootHashes, err := prepareForPdpVerification(native, sectorInfo, challenges, &sectorProveData)
	if err != nil {
		return false, errors.NewErr("[SectorProve] prepareForPdpVerification error")
	}
	err = verifier.VerifyProofWithMerklePath(0, sectorProveData.Proofs, fileIDs, tags, updatedChal, path, rootHashes)
	if err != nil {
		log.Errorf("[SectorProve] VerifyProofWithMerklePath error %s", err)
		return false, errors.NewErr("[SectorProve] VerifyProofWithMerklePath error")
	}
	return true, nil
}

func prepareForPdpVerification(native *native.NativeService, sectorInfo *SectorInfo, challenges []pdp.Challenge,
	proveData *SectorProveData) ([]pdp.FileID, []pdp.Tag, []pdp.Challenge, []*pdp.MerklePath, [][]byte, error) {
	err := checkSectorProveData(sectorInfo, proveData)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.NewErr("[prepareForPdpVerification] checkSectorProveData error")
	}

	fileNum := sectorInfo.FileNum

	var sectorFileInfos []*SectorFileInfo
	sectorFileInfos, err = getOrderedSectorFileInfosForSector(native, sectorInfo.NodeAddr, sectorInfo.SectorID)
	if err != nil {
		log.Errorf("getOrderedSectorFileInfosForSector error %s", err)
		return nil, nil, nil, nil, nil, errors.NewErr("getOrderedSectorFileInfos error!")
	}

	if uint64(len(sectorFileInfos)) != fileNum {
		log.Errorf("len not match : %d %d", len(sectorFileInfos), fileNum)
		return nil, nil, nil, nil, nil, errors.NewErr("fileNum not match sectorFileInfo num!")
	}

	fileIDs := make([]pdp.FileID, 0)
	tags := make([]pdp.Tag, 0)
	updatedChal := make([]pdp.Challenge, 0)
	path := make([]*pdp.MerklePath, 0)
	rootHashes := make([][]byte, 0)

	var offset uint64
	var curIndex = 0

	challengeLen := len(challenges)
	for i := uint64(0); i < fileNum; i++ {
		sectorFileInfo := sectorFileInfos[i]

		fileHash := sectorFileInfo.FileHash

		blockCount := sectorFileInfo.BlockCount

		start := uint32(offset)
		end := uint32(offset + blockCount - 1)

		for i := curIndex; i < len(challenges); i++ {
			chal := challenges[curIndex]
			if chal.Index >= start && chal.Index <= end {
				fileInfo, err := getFsFileInfo(native, fileHash)
				if err != nil {
					return nil, nil, nil, nil, nil, errors.NewErr("[prepareForPdpVerification] getFsFileInfo error")
				}

				proveParam, err := getProveParam(fileInfo.FileProveParam)
				if err != nil {
					return nil, nil, nil, nil, nil, errors.NewErr("[prepareForPdpVerification] getProveParam error")
				}

				fileIDs = append(fileIDs, proveParam.FileID)
				tags = append(tags, proveData.Tags[curIndex])
				path = append(path, proveData.MerklePath[curIndex])
				rootHashes = append(rootHashes, proveParam.RootHash)
				updatedChal = append(updatedChal, pdp.Challenge{
					Index: chal.Index - start, // adjust the index to be index in the file for merkle path calculation for pdp
					Rand:  chal.Rand,
				})

				curIndex++
				// reach end of indexes
				if curIndex >= challengeLen {
					return fileIDs, tags, updatedChal, path, rootHashes, nil
				}
				// continue to check if there are other blocks challenged in same file
				continue
			}
			// check next file
			break
		}
		offset += blockCount
	}
	return fileIDs, tags, updatedChal, path, rootHashes, nil
}

func checkSectorProveData(sectorInfo *SectorInfo, proveData *SectorProveData) error {
	if proveData.ProveFileNum > getSectorFileNum(sectorInfo) {
		return errors.NewErr("[checkSectorProveData] proveFileNum larger than file num in sector")
	}
	if proveData.ProveFileNum > proveData.BlockNum {
		return errors.NewErr("[checkSectorProveData] proveFileNum larger than challenged block num in sector")
	}
	if proveData.BlockNum > sectorInfo.TotalBlockNum {
		return errors.NewErr("[checkSectorProveData] challenged block num  larger than total block num in sector")
	}
	return nil
}

func profitSplitForSector(native *native.NativeService, sectorInfo *SectorInfo, nodeInfo *FsNodeInfo, fsSetting *FsSetting) error {
	for _, file := range sectorInfo.FileList.List {
		fileHash := file.Hash

		fileInfo, err := getFsFileInfo(native, fileHash)
		if err != nil {
			return errors.NewErr("[SectorProve] fileInfo not found!")
		}

		proveDetails, err := getProveDetails(native, fileHash)
		if err != nil {
			return errors.NewErr("[SectorProve] GetProveDetails error!")
		}

		settleFlag := false
		fileExpiredHeight := fileInfo.ExpiredHeight
		found := false
		var proveDetail *ProveDetail

		for i := uint64(0); i < proveDetails.ProveDetailNum; i++ {
			if proveDetails.ProveDetails[i].WalletAddr == sectorInfo.NodeAddr {
				found = true
				proveDetail = &proveDetails.ProveDetails[i]
				haveProveTimes := proveDetail.ProveTimes
				if haveProveTimes == fileInfo.ProveTimes || uint64(native.Height) > fileExpiredHeight {
					proveDetail.Finished = true
					settleFlag = true
				}

				proveDetail.ProveTimes++
				break
			}
		}

		if !found {
			return errors.NewErr("[SectorProve] ProveDetail not found")
		}

		if err = setProveDetails(native, fileInfo.FileHash, proveDetails); err != nil {
			return errors.NewErr("[SectorProve] ProveDetails setProveDetails error!")
		}

		if settleFlag {
			err = settleForFile(native, fileInfo, nodeInfo, proveDetail, proveDetails, fsSetting)
			if err != nil {
				return errors.NewErr("[SectorProve] settle for file error!")
			}
			err = deleteFileFromSector(native, sectorInfo, fileInfo)
			if err != nil {
				return errors.NewErr("[SectorProve] delete file from sector error!")
			}
			// emit delete file event so that node can delete file from its sector
			DeleteFileEvent(native, fileHash, fileInfo.FileOwner)
		}
	}

	return nil
}
