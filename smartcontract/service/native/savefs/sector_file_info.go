package savefs

import (
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type SectorFileInfo struct {
	FileHash   []byte
	BlockCount uint64
}

func (this *SectorFileInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.FileHash)
	utils.EncodeVarUint(sink, this.BlockCount)
}

func (this *SectorFileInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error

	this.FileHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	if len(this.FileHash) == 0 {
		return errors.NewErr("length of fileHash is 0")
	}
	this.BlockCount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

// add fillSectorFileInfo func
func fillSectorFileList(native *native.NativeService, sectorInfo *SectorInfo) (*SectorInfo, error) {
	// do nothing when no file in sector
	if sectorInfo.FileNum == 0 {
		return sectorInfo, nil
	}

	sectorFileInfos, err := getOrderedSectorFileInfosForSector(native, sectorInfo.NodeAddr, sectorInfo.SectorID)
	if err != nil {
		return nil, err
	}

	for _, sectorFileInfo := range sectorFileInfos {
		sectorInfo.FileList.AddNoCheck(sectorFileInfo.FileHash)
	}

	return sectorInfo, nil
}
