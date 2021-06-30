package savefs

import (
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
	"strings"
)

const SECTOR_FILE_INFO_GROUP_MAX_LEN = 5000

// put sectorFileInfo in group to limit db operation
type SectorFileInfoGroup struct {
	FileNum     uint64
	GroupID     uint64
	MinFileHash []byte
	MaxFileHash []byte
	FileList    []*SectorFileInfo // store sorted sector file info
}

func (this *SectorFileInfoGroup) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.FileNum)
	utils.EncodeVarUint(sink, this.GroupID)
	utils.EncodeBytes(sink, this.MinFileHash)
	utils.EncodeBytes(sink, this.MaxFileHash)
	for i := uint64(0); i < this.FileNum; i++ {
		this.FileList[i].Serialization(sink)
	}
}

func (this *SectorFileInfoGroup) Deserialization(source *common.ZeroCopySource) error {
	var err error

	this.FileNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.GroupID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.MinFileHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.MaxFileHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}

	fileList := make([]*SectorFileInfo, 0)
	for i := uint64(0); i < this.FileNum; i++ {
		sectorFileInfo := new(SectorFileInfo)
		err = sectorFileInfo.Deserialization(source)
		if err != nil {
			return err
		}
		fileList = append(fileList, sectorFileInfo)
	}
	this.FileList = fileList
	return nil
}

type SectorFileInfoGroupRef struct {
	NodeAddr common.Address
	SectorID uint64
	GroupID  uint64
}

func (this *SectorFileInfoGroupRef) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.NodeAddr)
	utils.EncodeVarUint(sink, this.SectorID)
	utils.EncodeVarUint(sink, this.GroupID)
}

func (this *SectorFileInfoGroupRef) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.NodeAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.SectorID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.GroupID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

func addSectorFileInfo(native *native.NativeService, nodeAddr common.Address, sectorID uint64, sectorFileInfo *SectorFileInfo) (groupCreated bool, err error) {
	var groupInfo *SectorFileInfoGroup
	var groupNum uint64

	sectorInfo, err := getSectorInfo(native, nodeAddr, sectorID)
	if err != nil {
		return false, errors.NewErr("getSectorInfo error!")
	}

	groupNum = sectorInfo.GroupNum

	// valid group num begins from 1
	if groupNum == 0 {
		groupNum = 1
		groupCreated = true

		groupInfo = &SectorFileInfoGroup{
			FileNum:     0,
			GroupID:     groupNum,
			MinFileHash: nil,
			MaxFileHash: nil,
			FileList:    make([]*SectorFileInfo, 0),
		}
	} else {
		groupInfo, err = getSectorFileInfoGroup(native, nodeAddr, sectorID, groupNum)
		if err != nil {
			return false, errors.NewErr("getSectorFileInfoGroup error!")
		}

		if groupInfo.FileNum == SECTOR_FILE_INFO_GROUP_MAX_LEN {
			log.Debugf("reach max group len, allocat new group")

			groupNum++
			groupCreated = true

			groupInfo = &SectorFileInfoGroup{
				FileNum:     0,
				GroupID:     groupNum,
				MinFileHash: nil,
				MaxFileHash: nil,
				FileList:    make([]*SectorFileInfo, 0),
			}
		}
	}

	// add to group
	err = addSectorFileInfoToGroup(native, nodeAddr, sectorID, groupInfo, sectorFileInfo)
	if err != nil {
		return false, errors.NewErr("addSectorFileInfoToGroup error!")
	}

	return groupCreated, nil
}

// delete sector file info from group, group compact not performed
// do not delete sector group even no file in group,  groupDeleted always return false
func deleteSectorFileInfo(native *native.NativeService, nodeAddr common.Address, sectorID uint64, fileHash []byte) (groupDeleted bool, err error) {
	groupNum, err := getSectorFileInfoGroupNum(native, nodeAddr, sectorID)
	if err != nil {
		return false, errors.NewErr("getSectorFileInfoGroupNum error!")
	}

	for i := uint64(1); i <= groupNum; i++ {
		sectorFileInfoGroup, err := getSectorFileInfoGroup(native, nodeAddr, sectorID, i)
		if err != nil {
			return false, errors.NewErr("getSectorFileInfoGroup error!")
		}

		index, found := findFileInGroup(sectorFileInfoGroup, fileHash)
		if found {
			fileNum := sectorFileInfoGroup.FileNum

			fileList := sectorFileInfoGroup.FileList
			fileList = append(fileList[0:index], fileList[index+1:]...)
			sectorFileInfoGroup.FileList = fileList

			// update min or max file hash if needed
			if index == 0 {
				sectorFileInfoGroup.MinFileHash = fileHash
			}
			if index == fileNum-1 {
				sectorFileInfoGroup.MaxFileHash = fileHash
			}

			sectorFileInfoGroup.FileNum--

			err = setSectorFileInfoGroup(native, nodeAddr, sectorID, sectorFileInfoGroup)
			if err != nil {
				return false, errors.NewErr("setSectorFileInfoGroup error!")
			}

			return false, nil
		}
	}

	return false, errors.NewErr("file not found in sector")
}

func addSectorFileInfoToGroup(native *native.NativeService, nodeAddr common.Address, sectorID uint64,
	sectorFileInfoGroup *SectorFileInfoGroup, sectorFileInfo *SectorFileInfo) error {
	if sectorFileInfoGroup.FileNum == SECTOR_FILE_INFO_GROUP_MAX_LEN {
		return errors.NewErr("sectorInfoGroup is full")
	}

	fileList := sectorFileInfoGroup.FileList
	fileHash := sectorFileInfo.FileHash
	fileHashStr := string(fileHash)

	found := false

	// insert into the fileList and update min or max if needed
	for i := uint64(0); i < sectorFileInfoGroup.FileNum; i++ {
		if strings.Compare(string(fileList[i].FileHash), fileHashStr) > 0 {
			rear := append([]*SectorFileInfo{}, fileList[i:]...)
			fileList = append(fileList[:i], sectorFileInfo)
			fileList = append(fileList, rear...)

			found = true
			if i == 0 {
				sectorFileInfoGroup.MinFileHash = fileHash
			}
			break
		}
	}

	if !found {
		fileList = append(fileList, sectorFileInfo)
		sectorFileInfoGroup.MaxFileHash = fileHash

		// when it is first added set min and max
		if len(fileList) == 1 {
			sectorFileInfoGroup.MinFileHash = fileHash
		}
	}

	sectorFileInfoGroup.FileList = fileList
	sectorFileInfoGroup.FileNum++

	err := setSectorFileInfoGroup(native, nodeAddr, sectorID, sectorFileInfoGroup)
	if err != nil {
		return errors.NewErr("setSectorFileInfoGroup error!")
	}
	return nil
}

func setSectorFileInfoGroup(native *native.NativeService, nodeAddr common.Address, sectorID uint64, sectorInfoGroup *SectorFileInfoGroup) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	sectorFileInfoGroupKey := GenFsSectorFileInfoGroupKey(contract, nodeAddr, sectorID, sectorInfoGroup.GroupID)

	sink := common.NewZeroCopySink(nil)
	sectorInfoGroup.Serialization(sink)

	utils.PutBytes(native, sectorFileInfoGroupKey, sink.Bytes())
	return nil
}

func getSectorFileInfoGroup(native *native.NativeService, nodeAddr common.Address, sectorID uint64, groupID uint64) (*SectorFileInfoGroup, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	sectorFileInfoGroupKey := GenFsSectorFileInfoGroupKey(contract, nodeAddr, sectorID, groupID)

	item, err := utils.GetStorageItem(native, sectorFileInfoGroupKey)
	if err != nil {
		return nil, errors.NewErr("[SectorFileInfoGroup] getSectorFileInfoGroup GetStorageItem error!")
	}

	if item == nil {
		return nil, errors.NewErr("[SectorFileInfoGroup] SectorFileInfoGroup not found!")
	}

	var sectorFileInfoGroup SectorFileInfoGroup

	sectorFileInfoGroupSource := common.NewZeroCopySource(item.Value)
	err = sectorFileInfoGroup.Deserialization(sectorFileInfoGroupSource)
	if err != nil {
		return nil, errors.NewErr("[SectorFileInfoGroup] SectorFileInfoGroup deserialize error!")
	}
	return &sectorFileInfoGroup, nil
}

func deleteSectorFileInfoGroup(native *native.NativeService, nodeAddr common.Address, sectorID uint64, groupID uint64) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	sectorFileInfoGroupKey := GenFsSectorFileInfoGroupKey(contract, nodeAddr, sectorID, groupID)
	utils.DelStorageItem(native, sectorFileInfoGroupKey)
	return nil
}

func deleteAllSectorFileInfoGroup(native *native.NativeService, nodeAddr common.Address, sectorID uint64) error {
	groupNum, err := getSectorFileInfoGroupNum(native, nodeAddr, sectorID)
	if err != nil {
		return err
	}
	for i := uint64(1); i <= groupNum; i++ {
		err := deleteSectorFileInfoGroup(native, nodeAddr, sectorID, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// get ordered sector files info for sector for check sector prove
func getOrderedSectorFileInfosForSector(native *native.NativeService, nodeAddr common.Address, sectorID uint64) ([]*SectorFileInfo, error) {
	return mergeSortSectorFileInfo(native, nodeAddr, sectorID)
}

func getSectorFileInfoGroupNum(native *native.NativeService, nodeAddr common.Address, sectorID uint64) (uint64, error) {
	sectorInfo, err := getSectorInfo(native, nodeAddr, sectorID)
	if err != nil {
		return 0, errors.NewErr("getSectorInfo error!")
	}

	return sectorInfo.GroupNum, nil
}

func compareFileHash(fileHash1 []byte, fileHash2 []byte) int {
	return strings.Compare(string(fileHash1), string(fileHash2))
}

func findFileInGroup(group *SectorFileInfoGroup, fileHash []byte) (index uint64, found bool) {
	if group.FileNum == 0 {
		return 0, false
	}

	// compare with min and max file hash to know if in group scope
	if compareFileHash(group.MinFileHash, fileHash) > 0 || compareFileHash(group.MaxFileHash, fileHash) < 0 {
		return 0, false
	}

	start := uint64(0)
	end := uint64(group.FileNum - 1)
	index = 0

	for {
		// not found
		if start > end {
			break
		}

		index = (start + end) / 2

		result := compareFileHash(group.FileList[index].FileHash, fileHash)
		if result == 0 {
			return index, true
		} else if result > 0 {
			end = index - 1
		} else {
			start = index + 1
		}
	}

	return 0, false
}
