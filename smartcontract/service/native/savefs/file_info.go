/*
IsPlotFile
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
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

const (
	PRIVATE   = 0
	PUBLIC    = 1
	WHITELIST = 2
)

const (
	FileStorageTypeUseSpace = 0
	FileStorageTypeCustom   = 1
)

type FileInfo struct {
	FileHash       []byte
	FileOwner      common.Address
	FileDesc       []byte
	Privilege      uint64
	FileBlockNum   uint64
	FileBlockSize  uint64
	ProveInterval  uint64
	ProveTimes     uint64
	ExpiredHeight  uint64
	CopyNum        uint64
	Deposit        uint64
	FileProveParam []byte
	ProveBlockNum  uint64
	BlockHeight    uint64 // store file info block height
	ValidFlag      bool
	StorageType    uint64
	RealFileSize   uint64
	PrimaryNodes   NodeList // Nodes store file
	CandidateNodes NodeList // Nodes backup file
	BlocksRoot     []byte
	ProveLevel     uint64      // prove level will decide the proveInterval when set
	SectorRefs     []SectorRef // store sectors that has reference to this file
	IsPlotFile     bool
	PlotInfo       *PlotInfo
}

func (this *FileInfo) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.FileHash); err != nil {
		return fmt.Errorf("[FileInfo] [FileHash:%v] serialize from error:%v", this.FileHash, err)
	}
	if err := utils.WriteAddress(w, this.FileOwner); err != nil {
		return fmt.Errorf("[FileInfo] [FileOwner:%v] serialize from error:%v", this.FileOwner, err)
	}
	if err := utils.WriteBytes(w, this.FileDesc); err != nil {
		return fmt.Errorf("[FileInfo] [FileDesc:%v] serialize from error:%v", this.FileDesc, err)
	}
	if err := utils.WriteVarUint(w, this.Privilege); err != nil {
		return fmt.Errorf("[FileInfo] [Privilege:%v] serialize from error:%v", this.Privilege, err)
	}
	if err := utils.WriteVarUint(w, this.FileBlockNum); err != nil {
		return fmt.Errorf("[FileInfo] [FileBlockNum:%v] serialize from error:%v", this.FileBlockNum, err)
	}
	if err := utils.WriteVarUint(w, this.FileBlockSize); err != nil {
		return fmt.Errorf("[FileInfo] [FileBlockSize:%v] serialize from error:%v", this.FileBlockSize, err)
	}
	if err := utils.WriteVarUint(w, this.ProveInterval); err != nil {
		return fmt.Errorf("[FileInfo] [ProveInterval:%v] serialize from error:%v", this.ProveInterval, err)
	}
	if err := utils.WriteVarUint(w, this.ProveTimes); err != nil {
		return fmt.Errorf("[FileInfo] [ProveTimes:%v] serialize from error:%v", this.ProveTimes, err)
	}
	if err := utils.WriteVarUint(w, this.ExpiredHeight); err != nil {
		return fmt.Errorf("[FileInfo] [ExpiredHeight:%v] serialize from error:%v", this.ExpiredHeight, err)
	}
	if err := utils.WriteVarUint(w, this.CopyNum); err != nil {
		return fmt.Errorf("[FileInfo] [CopyNum:%v] serialize from error:%v", this.CopyNum, err)
	}
	if err := utils.WriteVarUint(w, this.Deposit); err != nil {
		return fmt.Errorf("[FileInfo] [Deposit:%v] serialize from error:%v", this.Deposit, err)
	}
	if err := utils.WriteBytes(w, this.FileProveParam); err != nil {
		return fmt.Errorf("[FileInfo] [FileProveParam:%v] serialize from error:%v", this.FileProveParam, err)
	}
	if err := utils.WriteVarUint(w, this.ProveBlockNum); err != nil {
		return fmt.Errorf("[FileInfo] [ProveBlockNum:%v] serialize from error:%v", this.ProveBlockNum, err)
	}
	if err := utils.WriteVarUint(w, this.BlockHeight); err != nil {
		return fmt.Errorf("[FileInfo] [BlockHeight:%v] serialize from error:%v", this.BlockHeight, err)
	}
	if err := utils.WriteBool(w, this.ValidFlag); err != nil {
		return fmt.Errorf("[FileInfo] [ValidFlag:%v] serialize from error:%v", this.ValidFlag, err)
	}
	if err := utils.WriteVarUint(w, this.StorageType); err != nil {
		return fmt.Errorf("[FileInfo] [StorageType:%v] serialize from error:%v", this.StorageType, err)
	}
	if err := utils.WriteVarUint(w, this.RealFileSize); err != nil {
		return fmt.Errorf("[FileInfo] [RealFileSize:%v] serialize from error:%v", this.RealFileSize, err)
	}
	if err := this.PrimaryNodes.Serialize(w); err != nil {
		log.Errorf("serialize primary nodes err %s", err)
	}
	if err := this.CandidateNodes.Serialize(w); err != nil {
		log.Errorf("serialize candidate nodes err %s", err)
	}
	if err := utils.WriteBytes(w, this.BlocksRoot); err != nil {
		return fmt.Errorf("[FileInfo] [BlocksRoot:%v] serialize from error:%v", this.BlocksRoot, err)
	}
	if err := utils.WriteVarUint(w, this.ProveLevel); err != nil {
		return fmt.Errorf("[FileInfo] [ProveLevel:%v] serialize from error:%v", this.RealFileSize, err)
	}
	if err := utils.WriteVarUint(w, uint64(len(this.SectorRefs))); err != nil {
		return fmt.Errorf("[FileInfo] [SectorRefs len:%v] serialize from error:%v", this.RealFileSize, err)
	}
	for i := 0; i < len(this.SectorRefs); i++ {
		ref := this.SectorRefs[i]
		if err := utils.WriteAddress(w, ref.NodeAddr); err != nil {
			return fmt.Errorf("[FileInfo] [SectorRefs NodeAddr:%v] serialize from error:%v", this.RealFileSize, err)
		}
		if err := utils.WriteVarUint(w, ref.SectorID); err != nil {
			return fmt.Errorf("[FileInfo] [SectorRefs SectorID:%v] serialize from error:%v", this.RealFileSize, err)
		}
	}
	if err := utils.WriteBool(w, this.IsPlotFile); err != nil {
		return fmt.Errorf("[FileInfo] [IsPlotFile:%v] serialize from error:%v", this.IsPlotFile, err)
	}

	if this.IsPlotFile {
		if this.PlotInfo == nil {
			return fmt.Errorf("[FileInfo] PlotInfo is nil")
		}
		if err := this.PlotInfo.Serialize(w); err != nil {
			return fmt.Errorf("[FileInfo] [PlotInfo:%v] serialize from error:%v", this.PlotInfo, err)
		}
	}
	return nil
}

func (this *FileInfo) Deserialize(r io.Reader) error {
	var err error
	if this.FileHash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FileInfo] [FileHash] deserialize from error:%v", err)
	}
	if this.FileOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[FileInfo] [FileOwner] deserialize from error:%v", err)
	}
	if this.FileDesc, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FileInfo] [FileDesc] deserialize from error:%v", err)
	}
	if this.Privilege, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [Privilege] deserialize from error:%v", err)
	}
	if this.FileBlockNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [FileBlockNum] deserialize from error:%v", err)
	}
	if this.FileBlockSize, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [FileBlockSize] deserialize from error:%v", err)
	}
	if this.ProveInterval, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [ProveInterval] deserialize from error:%v", err)
	}
	if this.ProveTimes, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [ProveTimes] deserialize from error:%v", err)
	}
	if this.ExpiredHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [ExpiredHeight] deserialize from error:%v", err)
	}
	if this.CopyNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [CopyNum] deserialize from error:%v", err)
	}
	if this.Deposit, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [Deposit] deserialize from error:%v", err)
	}
	if this.FileProveParam, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[FileInfo] [FileProveParam] deserialize from error:%v", err)
	}
	if this.ProveBlockNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [ProveBlockNum] deserialize from error:%v", err)
	}
	if this.BlockHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [BlockHeight] deserialize from error:%v", err)
	}
	if this.ValidFlag, err = utils.ReadBool(r); err != nil {
		return fmt.Errorf("[FileInfo] [ValidFlag] deserialize from error:%v", err)
	}
	if this.StorageType, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [StorageType] deserialize from error:%v", err)
	}
	if this.RealFileSize, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [RealFileSize] deserialize from error:%v", err)
	}
	var primNodes NodeList
	if err = primNodes.Deserialize(r); err == nil {
		this.PrimaryNodes = primNodes
	} else {
		log.Errorf("Deserialize primary nodes err %s", err)
	}
	var candiNodes NodeList
	if err = candiNodes.Deserialize(r); err == nil {
		this.CandidateNodes = candiNodes
	} else {
		log.Errorf("Deserialize candidate nodes err %s", err)
	}
	if this.BlocksRoot, err = utils.ReadBytes(r); err != nil {
		log.Errorf("[FileInfo] [FileHash] deserialize from error:%v", err)
		// return fmt.Errorf("[FileInfo] [FileHash] deserialize from error:%v", err)
	}

	if this.ProveLevel, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [ProveLevel] deserialize from error:%v", err)
	}
	var refLen uint64
	var nodeAddr common.Address
	var sectorId uint64
	sectorRefs := make([]SectorRef, 0)

	if refLen, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfo] [SectorRefs len] deserialize from error:%v", err)
	}
	for i := uint64(0); i < refLen; i++ {
		if nodeAddr, err = utils.ReadAddress(r); err != nil {
			return fmt.Errorf("[FileInfo] [SectorRefs NodeAddr] deserialize from error:%v", err)
		}
		if sectorId, err = utils.ReadVarUint(r); err != nil {
			return fmt.Errorf("[FileInfo] [SectorRefs SectorID] deserialize from error:%v", err)
		}
		sectorRefs = append(sectorRefs, SectorRef{nodeAddr, sectorId})
	}
	this.SectorRefs = sectorRefs

	if this.IsPlotFile, err = utils.ReadBool(r); err != nil {
		return fmt.Errorf("[FileInfo] [IsPlotFile] deserialize from error:%v", err)
	}

	if this.IsPlotFile {
		plotInfo := new(PlotInfo)
		if err = plotInfo.Deserialize(r); err != nil {
			return fmt.Errorf("[FileInfo] [PlotInfo] deserialize from error:%v", err)
		}
		this.PlotInfo = plotInfo
	}
	return nil
}

func (this *FileInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeBytes(sink, this.FileHash)
	utils.EncodeAddress(sink, this.FileOwner)
	utils.EncodeBytes(sink, this.FileDesc)
	utils.EncodeVarUint(sink, this.Privilege)
	utils.EncodeVarUint(sink, this.FileBlockNum)
	utils.EncodeVarUint(sink, this.FileBlockSize)
	utils.EncodeVarUint(sink, this.ProveInterval)
	utils.EncodeVarUint(sink, this.ProveTimes)
	utils.EncodeVarUint(sink, this.ExpiredHeight)
	utils.EncodeVarUint(sink, this.CopyNum)
	utils.EncodeVarUint(sink, this.Deposit)
	utils.EncodeBytes(sink, this.FileProveParam)
	utils.EncodeVarUint(sink, this.ProveBlockNum)
	utils.EncodeVarUint(sink, this.BlockHeight)
	utils.EncodeBool(sink, this.ValidFlag)
	utils.EncodeVarUint(sink, this.StorageType)
	utils.EncodeVarUint(sink, this.RealFileSize)
	this.PrimaryNodes.Serialization(sink)
	this.CandidateNodes.Serialization(sink)
	utils.EncodeBytes(sink, this.BlocksRoot)
	utils.EncodeVarUint(sink, this.ProveLevel)
	utils.EncodeVarUint(sink, uint64(len(this.SectorRefs)))
	for i := 0; i < len(this.SectorRefs); i++ {
		ref := this.SectorRefs[i]
		utils.EncodeAddress(sink, ref.NodeAddr)
		utils.EncodeVarUint(sink, ref.SectorID)
	}
	utils.EncodeBool(sink, this.IsPlotFile)
	if this.IsPlotFile && this.PlotInfo != nil {
		this.PlotInfo.Serialization(sink)
	}
}

func (this *FileInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.FileOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.FileDesc, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.Privilege, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FileBlockNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FileBlockSize, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ProveInterval, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ProveTimes, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ExpiredHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.CopyNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Deposit, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FileProveParam, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	this.ProveBlockNum, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.BlockHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ValidFlag, err = utils.DecodeBool(source)
	if err != nil {
		return err
	}
	this.StorageType, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.RealFileSize, err = utils.DecodeVarUint(source)
	if err != nil {
		log.Errorf("decode real size err %s", err)
	}
	var primNodes NodeList
	if err = primNodes.Deserialization(source); err == nil {
		this.PrimaryNodes = primNodes
	} else {
		log.Errorf("deserialize primary nodes err %s", err)
	}
	var candiNodes NodeList
	if err = candiNodes.Deserialization(source); err == nil {
		this.CandidateNodes = candiNodes
	} else {
		log.Errorf("deserialize candidate nodes err %s", err)
	}
	this.BlocksRoot, err = utils.DecodeBytes(source)
	if err != nil {
		log.Errorf("decode blocks root failed %s", err)
	}
	this.ProveLevel, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	var refLen uint64
	var nodeAddr common.Address
	var sectorId uint64
	sectorRefs := make([]SectorRef, 0)
	refLen, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	for i := uint64(0); i < refLen; i++ {
		nodeAddr, err = utils.DecodeAddress(source)
		if err != nil {
			return err
		}
		sectorId, err = utils.DecodeVarUint(source)
		if err != nil {
			return err
		}
		sectorRefs = append(sectorRefs, SectorRef{nodeAddr, sectorId})
	}
	this.SectorRefs = sectorRefs

	this.IsPlotFile, err = utils.DecodeBool(source)
	if err != nil {
		return err
	}
	if this.IsPlotFile {
		plotInfo := new(PlotInfo)
		if err = plotInfo.Deserialization(source); err != nil {
			return err
		}
		this.PlotInfo = plotInfo
	}
	return nil
}

type FileInfoList struct {
	FileNum uint64
	List    []FileInfo
}

func (this *FileInfoList) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.FileNum); err != nil {
		return fmt.Errorf("[FileInfoList] [FileNum:%v] serialize from error:%v", this.FileNum, err)
	}

	for index := 0; uint64(index) < this.FileNum; index++ {
		if err := this.List[index].Serialize(w); err != nil {
			return fmt.Errorf("[FileInfoList] [List:%v] serialize from error:%v", this.List[index].FileHash, err)
		}
	}
	return nil
}

func (this *FileInfoList) Deserialize(r io.Reader) error {
	var err error
	if this.FileNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileInfoList] [FileNum] deserialize from error:%v", err)
	}
	var tmpInfo FileInfo
	for index := 0; uint64(index) < this.FileNum; index++ {
		if err := tmpInfo.Deserialize(r); err != nil {
			return fmt.Errorf("[FileInfoList] [List] deserialize from error:%v", err)
		}
		this.List = append(this.List, tmpInfo)
	}
	return nil
}

type PlotInfo struct {
	NumericID  uint64 // numeric ID for plot file
	StartNonce uint64 // start nonce in plot file
	Nonces     uint64 // number of nonce in plot file
}

func (this *PlotInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.NumericID); err != nil {
		return fmt.Errorf("[PlotInfo] [NumericID:%v] serialize from error:%v", this.NumericID, err)
	}
	if err := utils.WriteVarUint(w, this.StartNonce); err != nil {
		return fmt.Errorf("[PlotInfo] [StartNonce:%v] serialize from error:%v", this.StartNonce, err)
	}
	if err := utils.WriteVarUint(w, this.Nonces); err != nil {
		return fmt.Errorf("[PlotInfo] [Nonces:%v] serialize from error:%v", this.Nonces, err)
	}
	return nil
}
func (this *PlotInfo) Deserialize(r io.Reader) error {
	var err error
	if this.NumericID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[PlotInfo] [NumericID] deserialize from error:%v", err)
	}
	if this.StartNonce, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[PlotInfo] [StartNonce] deserialize from error:%v", err)
	}
	if this.Nonces, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[PlotInfo] [Nonces] deserialize from error:%v", err)
	}
	return nil
}

func (this *PlotInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.NumericID)
	utils.EncodeVarUint(sink, this.StartNonce)
	utils.EncodeVarUint(sink, this.Nonces)
}

func (this *PlotInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.NumericID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.StartNonce, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Nonces, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

func getFsFileInfo(native *native.NativeService, fileHash []byte) (*FileInfo, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileInfoKey := GenFsFileInfoKey(contract, fileHash)
	item, err := utils.GetStorageItem(native, fileInfoKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo GetStorageItem error!")
	}
	if item == nil {
		return nil, errors.NewErr("[FS Profit] FsFileInfo not found!")
	}

	var fsFileInfo FileInfo
	fsFileInfoSource := common.NewZeroCopySource(item.Value)
	err = fsFileInfo.Deserialization(fsFileInfoSource)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo deserialize error!")
	}
	return &fsFileInfo, nil
}

func setFsFileInfo(native *native.NativeService, fileInfo *FileInfo) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	bf := new(bytes.Buffer)
	if err := fileInfo.Serialize(bf); err != nil {
		return errors.NewErr("[FS Profit] FsFileInfo serialize error!")
	}
	fileInfoKey := GenFsFileInfoKey(contract, fileInfo.FileHash)
	utils.PutBytes(native, fileInfoKey, bf.Bytes())

	return nil
}

func deleteFsFileInfo(native *native.NativeService, fileHash []byte) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	fileInfoKey := GenFsFileInfoKey(contract, fileHash)
	utils.DelStorageItem(native, fileInfoKey)
}

func addSectorRefForFileInfo(native *native.NativeService, fileInfo *FileInfo, sectorInfo *SectorInfo) error {
	if isSectorRefByFileInfo(fileInfo, sectorInfo.NodeAddr, sectorInfo.SectorID) {
		return fmt.Errorf("addSectorRefForFileInfo file already has sector ref")
	}

	fileInfo.SectorRefs = append(fileInfo.SectorRefs, SectorRef{sectorInfo.NodeAddr, sectorInfo.SectorID})
	return setFsFileInfo(native, fileInfo)
}

func isSectorRefByFileInfo(fileInfo *FileInfo, nodeAddr common.Address, sectorID uint64) bool {
	for _, ref := range fileInfo.SectorRefs {
		if ref.NodeAddr.ToBase58() == nodeAddr.ToBase58() && ref.SectorID == sectorID {
			return true
		}
	}
	return false
}
