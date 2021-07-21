package savefs

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

const (
	PROVE_LEVEL_HIGH = iota + 1
	PROVE_LEVEL_MEDIEUM
	PROVE_LEVEL_LOW
)

const (
	PROVE_PERIOD_HIGHT   = DEFAULT_PROVE_PERIOD
	PROVE_PERIOD_MEDIEUM = 2 * DEFAULT_PROVE_PERIOD
	PROVE_PERIOD_LOW     = 8 * DEFAULT_PROVE_PERIOD
)

type SectorInfo struct {
	NodeAddr         common.Address
	SectorID         uint64   // node defines the sector id
	Size             uint64   // declared sector size, should be no less than sum of all files
	Used             uint64   // used sector size
	ProveLevel       uint64   // files in same sector has same prove level
	FirstProveHeight uint64   // first prove height for sector
	NextProveHeight  uint64   // next prove height for sector
	TotalBlockNum    uint64   // total block num in the sector
	FileNum          uint64   // total file num in sector
	GroupNum         uint64   // sectorInfoGroup num
	IsPlots          bool     // is plots sector
	FileList         FileList // store the file in the order it is uploaded
}

func (this *SectorInfo) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.NodeAddr); err != nil {
		return fmt.Errorf("[SectorInfo] [NodeAddr:%v] serialize from error:%v", this.NodeAddr, err)
	}
	if err := utils.WriteVarUint(w, this.SectorID); err != nil {
		return fmt.Errorf("[SectorInfo] [SectorID:%v] serialize from error:%v", this.SectorID, err)
	}
	if err := utils.WriteVarUint(w, this.Size); err != nil {
		return fmt.Errorf("[SectorInfo] [Size:%v] serialize from error:%v", this.Size, err)
	}
	if err := utils.WriteVarUint(w, this.Used); err != nil {
		return fmt.Errorf("[SectorInfo] [Used:%v] serialize from error:%v", this.Used, err)
	}
	if err := utils.WriteVarUint(w, this.ProveLevel); err != nil {
		return fmt.Errorf("[SectorInfo] [ProveLevel:%v] serialize from error:%v", this.ProveLevel, err)
	}
	if err := utils.WriteVarUint(w, this.FirstProveHeight); err != nil {
		return fmt.Errorf("[SectorInfo] [FirstProveHeight:%v] serialize from error:%v", this.ProveLevel, err)
	}
	if err := utils.WriteVarUint(w, this.NextProveHeight); err != nil {
		return fmt.Errorf("[SectorInfo] [NextProveHeight:%v] serialize from error:%v", this.ProveLevel, err)
	}
	if err := utils.WriteVarUint(w, this.TotalBlockNum); err != nil {
		return fmt.Errorf("[SectorInfo] [TotalBlockNum:%v] serialize from error:%v", this.TotalBlockNum, err)
	}
	if err := utils.WriteVarUint(w, this.FileNum); err != nil {
		return fmt.Errorf("[SectorInfo] [FileNum:%v] serialize from error:%v", this.FileNum, err)
	}
	if err := utils.WriteVarUint(w, this.GroupNum); err != nil {
		return fmt.Errorf("[SectorInfo] [GroupNum:%v] serialize from error:%v", this.GroupNum, err)
	}
	if err := utils.WriteBool(w, this.IsPlots); err != nil {
		return fmt.Errorf("[SectorInfo] [IsPlots:%v] serialize from error:%v", this.FileList, err)
	}
	if err := this.FileList.Serialize(w); err != nil {
		return fmt.Errorf("[SectorInfo] [FileList:%v] serialize from error:%v", this.FileList, err)
	}

	return nil
}

func (this *SectorInfo) Deserialize(r io.Reader) error {
	var err error
	if this.NodeAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[SectorInfo] [NodeAddr] Deserialize from error:%v", err)
	}
	if this.SectorID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfo] [SectorID] Deserialize from error:%v", err)
	}
	if this.Size, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfo] [Size] Deserialize from error:%v", err)
	}
	if this.Used, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfo] [Used] Deserialize from error:%v", err)
	}
	if this.ProveLevel, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfo] [ProveLevel] Deserialize from error:%v", err)
	}
	if this.FirstProveHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfo] [FirstProveHeight] Deserialize from error:%v", err)
	}
	if this.NextProveHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfo] [NextProveHeight] Deserialize from error:%v", err)
	}
	if this.TotalBlockNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfo] [TotalBlockNum] Deserialize from error:%v", err)
	}
	if this.FileNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfo] [FileNum] Deserialize from error:%v", err)
	}
	if this.GroupNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfo] [GroupNum] Deserialize from error:%v", err)
	}
	if this.IsPlots, err = utils.ReadBool(r); err != nil {
		return fmt.Errorf("[SectorInfo] [IsPlots] Deserialize from error:%v", err)
	}
	if err = this.FileList.Deserialize(r); err != nil {
		return fmt.Errorf("[SectorInfo] [FileList] Deserialize from error:%v", err)
	}

	return nil
}

func (this *SectorInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.NodeAddr)
	utils.EncodeVarUint(sink, this.SectorID)
	utils.EncodeVarUint(sink, this.Size)
	utils.EncodeVarUint(sink, this.Used)
	utils.EncodeVarUint(sink, this.ProveLevel)
	utils.EncodeVarUint(sink, this.FirstProveHeight)
	utils.EncodeVarUint(sink, this.NextProveHeight)
	utils.EncodeVarUint(sink, this.TotalBlockNum)
	utils.EncodeVarUint(sink, this.FileNum)
	utils.EncodeVarUint(sink, this.GroupNum)
	utils.EncodeBool(sink, this.IsPlots)
	this.FileList.Serialization(sink)
}

func (this *SectorInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.NodeAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.SectorID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Size, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Used, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ProveLevel, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FirstProveHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.NextProveHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.TotalBlockNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FileNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.GroupNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.IsPlots, err = utils.DecodeBool(source)
	if err != nil {
		return err
	}
	err = this.FileList.Deserialization(source)
	if err != nil {
		return err
	}
	return nil
}

type SectorRef struct {
	NodeAddr common.Address
	SectorID uint64
}

func (this *SectorRef) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.NodeAddr); err != nil {
		return fmt.Errorf("[SectorRef] [NodeAddr:%v] serialize from error:%v", this.NodeAddr, err)
	}
	if err := utils.WriteAddress(w, this.NodeAddr); err != nil {
		return fmt.Errorf("[SectorRef] [SectorID:%v] serialize from error:%v", this.NodeAddr, err)
	}
	return nil
}

func (this *SectorRef) Deserialize(r io.Reader) error {
	var err error
	if this.NodeAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[SectorRef] [NodeAddr] Deserialize from error:%v", err)
	}
	if this.SectorID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorRef] [SectorID] Deserialize from error:%v", err)
	}
	return nil
}

func (this *SectorRef) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.NodeAddr)
	utils.EncodeVarUint(sink, this.SectorID)
}

func (this *SectorRef) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.NodeAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.SectorID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

type SectorFileRef struct {
	NodeAddr common.Address
	SectorID uint64
	FileHash []byte
}

func (this *SectorFileRef) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.NodeAddr); err != nil {
		return fmt.Errorf("[SectorFileRef] [NodeAddr:%v] serialize from error:%v", this.NodeAddr, err)
	}
	if err := utils.WriteVarUint(w, this.SectorID); err != nil {
		return fmt.Errorf("[SectorFileRef] [SectorID:%v] serialize from error:%v", this.NodeAddr, err)
	}
	if err := utils.WriteBytes(w, this.FileHash); err != nil {
		return fmt.Errorf("[SectorFileRef] [FileHash:%v] serialize from error:%v", this.NodeAddr, err)
	}
	return nil
}

func (this *SectorFileRef) Deserialize(r io.Reader) error {
	var err error
	if this.NodeAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[SectorFileRef] [NodeAddr] Deserialize from error:%v", err)
	}
	if this.SectorID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorFileRef] [SectorID] Deserialize from error:%v", err)
	}
	if this.FileHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[SectorFileRef] [FileHash] Deserialize from error:%v", err)
	}
	return nil
}

func (this *SectorFileRef) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.NodeAddr)
	utils.EncodeVarUint(sink, this.SectorID)
	utils.EncodeBytes(sink, this.FileHash)
}

func (this *SectorFileRef) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.NodeAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.SectorID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FileHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}

type SectorInfos struct {
	SectorCount uint64
	Sectors     []*SectorInfo
}

func (this *SectorInfos) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.SectorCount); err != nil {
		return fmt.Errorf("[SectorInfos] [SectorCount:%v] serialize from error:%v", this.SectorCount, err)
	}
	if uint64(len(this.Sectors)) != this.SectorCount {
		return fmt.Errorf("[SectorInfos] sectorCount and number of sectorInfos no match")
	}

	for _, sector := range this.Sectors {
		if err := sector.Serialize(w); err != nil {
			return fmt.Errorf("[SectorInfos] [Sector%v] serialize from error:%v", sector, err)
		}
	}
	return nil
}

func (this *SectorInfos) Deserialize(r io.Reader) error {
	var err error
	if this.SectorCount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorInfos] [SectorCount] Deserialize from error:%v", err)
	}

	sectors := make([]*SectorInfo, 0)
	for i := uint64(0); i < this.SectorCount; i++ {
		sector := new(SectorInfo)
		if err = sector.Deserialize(r); err != nil {
			return fmt.Errorf("[SectorInfos] [Sector%v] Deserialize  from error:%v", sector, err)
		}
		sectors = append(sectors, sector)
	}
	this.Sectors = sectors
	return nil
}

// caller should guarantee file with the fileHash exist
func addFileToSector(native *native.NativeService, sectorInfo *SectorInfo, fileInfo *FileInfo) error {
	// check first if enough space in sector for file
	if sectorInfo.Used+fileInfo.FileBlockNum*fileInfo.FileBlockSize > sectorInfo.Size {
		return errors.NewErr("addFileToSector error, not enough space in sector")
	}

	groupCreated, err := addSectorFileInfo(native, sectorInfo.NodeAddr, sectorInfo.SectorID, &SectorFileInfo{
		FileHash:   fileInfo.FileHash,
		BlockCount: fileInfo.FileBlockNum,
	})
	if err != nil {
		return errors.NewErr("addSectorFileInfo error!")
	}

	sectorInfo.FileNum++
	sectorInfo.Used += fileInfo.FileBlockNum * fileInfo.FileBlockSize
	sectorInfo.TotalBlockNum += fileInfo.FileBlockNum
	if groupCreated {
		sectorInfo.GroupNum++
	}
	return setSectorInfo(native, sectorInfo)
}

func deleteFileFromSector(native *native.NativeService, sectorInfo *SectorInfo, fileInfo *FileInfo) error {
	groupDeleted, err := deleteSectorFileInfo(native, sectorInfo.NodeAddr, sectorInfo.SectorID, fileInfo.FileHash)
	if err != nil {
		log.Errorf("deleteSectorFileInfo error %s", err)
		return errors.NewErr("deleteSectorFileInfo error!")
	}

	sectorInfo.FileNum--
	sectorInfo.TotalBlockNum -= fileInfo.FileBlockNum
	sectorInfo.Used -= fileInfo.FileBlockNum * fileInfo.FileBlockSize
	if groupDeleted {
		sectorInfo.GroupNum--
	}

	if getSectorFileNum(sectorInfo) == 0 {
		sectorInfo.NextProveHeight = 0
	}
	return setSectorInfo(native, sectorInfo)
}

// fileHash in the fileList should be removed after one prove interval after fileInfo is expired
func deleteExpiredFilesFromSector(native *native.NativeService, nodeAddr common.Address, sectorID uint64) ([]FileHash, error) {
	sectorInfo, err := getSectorInfoWithFileList(native, nodeAddr, sectorID)
	if err != nil {
		return nil, errors.NewErr("getSectorInfo error!")
	}

	fileHashes := make([]FileHash, 0)
	for _, fileHash := range sectorInfo.FileList.List {
		fileHashes = append(fileHashes, FileHash{Hash: fileHash.Hash})
	}

	deletedFiles := make([]FileHash, 0)
	for _, fileHash := range fileHashes {
		fileInfo, err := getFsFileInfo(native, fileHash.Hash)
		if err != nil {
			return nil, errors.NewErr("getFileInfo error!")
		}

		if isFileExpired(native, fileInfo) {
			err = deleteFileFromSector(native, sectorInfo, fileInfo)
			if err != nil {
				return nil, errors.NewErr("deleteFileFromSector error!")
			}
			deletedFiles = append(deletedFiles, FileHash{Hash: fileHash.Hash})
		}
	}

	return deletedFiles, nil
}

func isFileExpired(native *native.NativeService, fileInfo *FileInfo) bool {
	if fileInfo.ExpiredHeight+fileInfo.ProveInterval < uint64(native.Height) {
		return true
	}
	return false
}

func isFileInSector(native *native.NativeService, nodeAddr common.Address, sectorID uint64, fileHash []byte) (bool, error) {
	sectorInfo, err := getSectorInfoWithFileList(native, nodeAddr, sectorID)
	if err != nil {
		return false, errors.NewErr("getSectorInfo error!")
	}
	return sectorInfo.FileList.Has(fileHash), nil
}

func setSectorInfo(native *native.NativeService, sectorInfo *SectorInfo) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	// empty the file list, file list is saved in sector file info groups
	sectorInfo.FileList = FileList{}

	sectorInfoKey := GenFsSectorInfoKey(contract, sectorInfo.NodeAddr, sectorInfo.SectorID)

	sink := common.NewZeroCopySink(nil)
	sectorInfo.Serialization(sink)

	utils.PutBytes(native, sectorInfoKey, sink.Bytes())
	return nil
}

// get sector info without file list
func getSectorInfo(native *native.NativeService, nodeAddr common.Address, sectorID uint64) (*SectorInfo, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	sectorInfoKey := GenFsSectorInfoKey(contract, nodeAddr, sectorID)
	item, err := utils.GetStorageItem(native, sectorInfoKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[SectorInfo] SectorInfo GetStorageItem error!")
	}

	if item == nil {
		return nil, errors.NewErr("[SectorInfo] SectorInfo not found!")
	}

	log.Debugf("sectorInfo for node %s item.value : %v", nodeAddr.ToBase58(), item.Value)

	var sectorInfo SectorInfo

	sectorInfoSource := common.NewZeroCopySource(item.Value)
	err = sectorInfo.Deserialization(sectorInfoSource)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[SectorInfo] SectorInfo deserialize error!")
	}
	return &sectorInfo, nil
}

func deleteSectorInfo(native *native.NativeService, nodeAddr common.Address, sectorID uint64) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	_, err := getSectorInfo(native, nodeAddr, sectorID)
	if err != nil {
		return err
	}

	sectorInfoKey := GenFsSectorInfoKey(contract, nodeAddr, sectorID)
	utils.DelStorageItem(native, sectorInfoKey)
	return nil
}

// delete sector info and all sector info groups
func deleteSector(native *native.NativeService, nodeAddr common.Address, sectorID uint64) error {
	err := deleteSectorInfo(native, nodeAddr, sectorID)
	if err != nil {
		return errors.NewErr("deleteSectorInfo error!")
	}

	err = deleteAllSectorFileInfoGroup(native, nodeAddr, sectorID)
	if err != nil {
		return errors.NewErr("deleteAllSectorFileInfoGroup error!")
	}
	return nil
}

func getSectorInfoWithFileList(native *native.NativeService, nodeAddr common.Address, sectorID uint64) (*SectorInfo, error) {
	sectorInfo, err := getSectorInfo(native, nodeAddr, sectorID)
	if err != nil {
		return nil, errors.NewErr("getSectorInfo error!")
	}

	return fillSectorFileList(native, sectorInfo)
}

func getSectorsForNode(native *native.NativeService, nodeAddr common.Address) (*SectorInfos, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	sectorInfoPrefix := GenFsSectorInfoPrefix(contract, nodeAddr)

	sectorInfos := &SectorInfos{
		SectorCount: 0,
		Sectors:     make([]*SectorInfo, 0),
	}

	iter := native.CacheDB.NewIterator(sectorInfoPrefix[:])
	for has := iter.First(); has; has = iter.Next() {
		var sectorInfo SectorInfo
		source := common.NewZeroCopySource(iter.Value())
		if err := sectorInfo.Deserialization(source); err != nil {
			item, err := utils.GetStorageItem(native, iter.Key())
			if err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[SectorInfo] SectorInfo GetStorageItem error!")
			}

			if item == nil {
				return nil, errors.NewErr("[SectorInfo] SectorInfo not found!")
			}

			sectorInfoSource := common.NewZeroCopySource(item.Value)
			err = sectorInfo.Deserialization(sectorInfoSource)
			if err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[SectorInfo] SectorInfo deserialize error!")
			}
		}
		sectorInfos.Sectors = append(sectorInfos.Sectors, &sectorInfo)
		sectorInfos.SectorCount++
	}
	iter.Release()

	return sectorInfos, nil
}

func getSectorsWithFileListForNode(native *native.NativeService, nodeAddr common.Address) (*SectorInfos, error) {
	sectorInfos, err := getSectorsForNode(native, nodeAddr)
	if err != nil {
		return nil, errors.NewErr("getSectorsForNode error!")
	}

	for _, sectorInfo := range sectorInfos.Sectors {
		_, err := fillSectorFileList(native, sectorInfo)
		if err != nil {
			return nil, errors.NewErr("filleSectorFileList error!")
		}
	}

	return sectorInfos, nil
}

func getSectorTotalSizeForNode(native *native.NativeService, nodeAddr common.Address) (uint64, error) {
	sectorInfos, err := getSectorsForNode(native, nodeAddr)
	if err != nil {
		return 0, err
	}

	totalSize := (uint64)(0)

	for _, sectorInfo := range sectorInfos.Sectors {
		totalSize += sectorInfo.Size
	}
	return totalSize, nil
}

func getSectorFileNum(sectorInfo *SectorInfo) uint64 {
	return sectorInfo.FileNum
}
