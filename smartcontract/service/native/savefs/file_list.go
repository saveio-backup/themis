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
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type FileHash struct {
	Hash []byte
}

type FileList struct {
	FileNum uint64
	List    []FileHash
}

func (this *FileList) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.FileNum); err != nil {
		return fmt.Errorf("[FileList] [FileNum:%v] serialize from error:%v", this.FileNum, err)
	}

	for index := 0; uint64(index) < this.FileNum; index++ {
		if err := utils.WriteBytes(w, this.List[index].Hash); err != nil {
			return fmt.Errorf("[FileList] [FileList:%v] serialize from error:%v", this.List[index].Hash, err)
		}
	}
	return nil
}

func (this *FileList) Deserialize(r io.Reader) error {
	var err error
	if this.FileNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[FileList] [FileNum] deserialize from error:%v", err)
	}
	var tmpHash []byte
	for index := 0; uint64(index) < this.FileNum; index++ {
		if tmpHash, err = utils.ReadBytes(r); err != nil {
			return fmt.Errorf("[FileList] [FileList] deserialize from error:%v", err)
		}
		fileHash := FileHash{tmpHash}
		this.List = append(this.List, fileHash)
	}
	return nil
}

func (this *FileList) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.FileNum)
	for i := uint64(0); i < this.FileNum; i++ {
		utils.EncodeBytes(sink, this.List[i].Hash)
	}
}
func (this *FileList) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.FileNum, err = utils.DecodeVarUint(source); err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("fileList deserialization panic with fileNum %d", this.FileNum)
		}
	}()
	fileList := make([]FileHash, 0, this.FileNum)
	for i := uint64(0); i < this.FileNum; i++ {
		hash, err := utils.DecodeBytes(source)
		if err != nil {
			return err
		}
		fileList = append(fileList, FileHash{Hash: hash})
	}
	this.List = fileList
	return nil
}

func (this *FileList) Add(hash []byte) {
	flag := false
	for i := uint64(0); i < this.FileNum; i++ {
		if bytes.Equal(this.List[i].Hash, hash) {
			flag = true
			break
		}
	}
	if !flag {
		fileHash := FileHash{hash}
		this.List = append(this.List, fileHash)
		this.FileNum++
	}
}

// no check for duplicate file hash for performance
func (this *FileList) AddNoCheck(hash []byte) {
	fileHash := FileHash{hash}
	this.List = append(this.List, fileHash)
	this.FileNum++
}

func (this *FileList) Del(hash []byte) error {
	var i uint64
	if this.FileNum == 0 {
		return nil
	}
	for i = 0; i < this.FileNum; i++ {
		if bytes.Equal(this.List[i].Hash, hash) {
			this.List = append(this.List[:i], this.List[i+1:]...)
			this.FileNum -= 1
			break
		}
	}
	return nil
}

func (this *FileList) Has(hash []byte) bool {
	var i uint64
	if this.FileNum == 0 {
		return false
	}
	for i = 0; i < this.FileNum; i++ {
		if bytes.Equal(this.List[i].Hash, hash) {
			return true
		}
	}
	return false
}

func AddFileToList(native *native.NativeService, walletAddr common.Address, fileHash []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsFileListKey(contract, walletAddr)
	return addFileToList(native, fileListKey, walletAddr, fileHash)
}

func DelFileFromList(native *native.NativeService, walletAddr common.Address, fileHash []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsFileListKey(contract, walletAddr)
	return delFileFromList(native, fileListKey, walletAddr, fileHash)
}

func GetFsFileList(native *native.NativeService, walletAddr common.Address) (*FileList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsFileListKey(contract, walletAddr)
	return getFsFileList(native, fileListKey)
}

func AddFileToPrimaryList(native *native.NativeService, walletAddr common.Address, fileHash []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsFilePrimaryListKey(contract, walletAddr)
	return addFileToList(native, fileListKey, walletAddr, fileHash)
}

func DelFileFromPrimaryList(native *native.NativeService, walletAddr common.Address, fileHash []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsFilePrimaryListKey(contract, walletAddr)
	return delFileFromList(native, fileListKey, walletAddr, fileHash)
}

func GetFsFilePrimaryList(native *native.NativeService, walletAddr common.Address) (*FileList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsFilePrimaryListKey(contract, walletAddr)
	return getFsFileList(native, fileListKey)
}

func AddFileToCandidateList(native *native.NativeService, walletAddr common.Address, fileHash []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsFileCandidateListKey(contract, walletAddr)
	return addFileToList(native, fileListKey, walletAddr, fileHash)
}

func DelFileFromCandidateList(native *native.NativeService, walletAddr common.Address, fileHash []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsFileCandidateListKey(contract, walletAddr)
	return delFileFromList(native, fileListKey, walletAddr, fileHash)
}

func GetFsFileCandidateList(native *native.NativeService, walletAddr common.Address) (*FileList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsFileCandidateListKey(contract, walletAddr)
	return getFsFileList(native, fileListKey)

}

// UnSettled file list saves files which has at least one node submitted last prove, but not all proves are submitted
// this kind of files can be deleted when all nodes finish last prove or by user with deleteUnsettledFiles after one
// prove interval of file expire
func AddFileToUnSettleList(native *native.NativeService, walletAddr common.Address, fileHash []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsUnSettledListKey(contract, walletAddr)
	return addFileToList(native, fileListKey, walletAddr, fileHash)
}

func DelFileFromUnSettledList(native *native.NativeService, walletAddr common.Address, fileHash []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsUnSettledListKey(contract, walletAddr)
	return delFileFromList(native, fileListKey, walletAddr, fileHash)
}

func GetFsUnSettledList(native *native.NativeService, walletAddr common.Address) (*FileList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileListKey := GenFsUnSettledListKey(contract, walletAddr)
	return getFsFileList(native, fileListKey)
}

func addFileToList(native *native.NativeService, fileListKey []byte, walletAddr common.Address, fileHash []byte) error {
	var fileList *FileList
	fileList, err := getFsFileList(native, fileListKey)
	if fileList == nil {
		fileList = new(FileList)
	}
	fileList.Add(fileHash)
	fileListBf := new(bytes.Buffer)
	err = fileList.Serialize(fileListBf)
	if err != nil {
		return errors.NewErr("[FS Profit] FsDeleteFile fileList serialize error!")
	}
	utils.PutBytes(native, fileListKey, fileListBf.Bytes())
	return nil
}

func delFileFromList(native *native.NativeService, fileListKey []byte, walletAddr common.Address, fileHash []byte) error {
	var fileList *FileList
	fileList, err := getFsFileList(native, fileListKey)
	if err != nil {
		return errors.NewErr("[FS Profit] FsDeleteFile getFsFileList error!")
	}
	if fileList == nil || fileList.FileNum == 0 {
		return nil
	}
	fileList.Del(fileHash)
	fileListBf := new(bytes.Buffer)
	err = fileList.Serialize(fileListBf)
	if err != nil {
		return errors.NewErr("[FS Profit] FsDeleteFile fileList serialize error!")
	}
	utils.PutBytes(native, fileListKey, fileListBf.Bytes())
	return nil
}

func getFsFileList(native *native.NativeService, fileListKey []byte) (*FileList, error) {
	item, err := utils.GetStorageItem(native, fileListKey)
	if err != nil {
		return nil, errors.NewErr("[FS Profit] FsFileList GetStorageItem error!")
	}
	if item == nil {
		return &FileList{0, nil}, nil
	}
	var fsFileList FileList
	reader := bytes.NewReader(item.Value)
	err = fsFileList.Deserialize(reader)
	if err != nil {
		return nil, errors.NewErr("[FS Profit] FsFileList deserialize error!")
	}
	return &fsFileList, nil
}

func setFsFileList(native *native.NativeService, fileListKey []byte, fileList *FileList) error {
	fileListBf := new(bytes.Buffer)
	err := fileList.Serialize(fileListBf)
	if err != nil {
		return errors.NewErr("[FS Profit] FsSetFileList fileList serialize error!")
	}
	utils.PutBytes(native, fileListKey, fileListBf.Bytes())
	return nil
}
