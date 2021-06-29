package savefs

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type StorageFee struct {
	TxnFee        uint64
	SpaceFee      uint64
	ValidationFee uint64
}

func (f *StorageFee) Sum() uint64 {
	return f.TxnFee + f.SpaceFee + f.ValidationFee
}

func (this *StorageFee) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.TxnFee); err != nil {
		return fmt.Errorf("[StorageFee] [TxnFee:%v] serialize from error:%v", this.TxnFee, err)
	}
	if err := utils.WriteVarUint(w, this.SpaceFee); err != nil {
		return fmt.Errorf("[StorageFee] [SpaceFee:%v] serialize from error:%v", this.SpaceFee, err)
	}
	if err := utils.WriteVarUint(w, this.ValidationFee); err != nil {
		return fmt.Errorf("[StorageFee] [ValidationFee:%v] serialize from error:%v", this.ValidationFee, err)
	}
	return nil
}

func (this *StorageFee) Deserialize(r io.Reader) error {
	var err error
	if this.TxnFee, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[StorageFee] [TxnFee] Deserialize from error:%v", err)
	}
	if this.SpaceFee, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[StorageFee] [SpaceFee] Deserialize from error:%v", err)
	}
	if this.ValidationFee, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[StorageFee] [ValidationFee] Deserialize from error:%v", err)
	}
	return nil
}

func calcUploadFee(uploadInfo *UploadOption, setting *FsSetting, currentHeight uint32) *StorageFee {
	fee := uint64(0)
	txGas := uint64(10000000)
	if uploadInfo.WhiteList.Num > 0 {
		fee = txGas * 4
	} else {
		fee = txGas * 3
	}
	sf := &StorageFee{
		TxnFee: fee,
	}
	//transacton fee only
	log.Debugf("FileStoreType(uploadInfo.StorageType) : %v", FileStoreType(uploadInfo.StorageType))
	if FileStoreType(uploadInfo.StorageType) == FileStoreTypeNormal {
		return sf
	}
	depositFee := calcDepositFee(uploadInfo, setting, currentHeight)
	sf.ValidationFee = depositFee.ValidationFee
	sf.SpaceFee = depositFee.SpaceFee
	return sf
}

func calcDepositFee(uploadInfo *UploadOption, setting *FsSetting, currentHeight uint32) *StorageFee {
	// fileSize unit is kb
	fileSize := uint64(uploadInfo.FileSize)
	if fileSize <= 0 {
		fileSize = 1
	}
	proveTime := calcProveTimesByUploadInfo(uploadInfo, currentHeight)
	fee := calcFee(setting, proveTime, uploadInfo.CopyNum, fileSize, uploadInfo.ExpiredHeight-uint64(currentHeight))

	return fee
}

// calculate the userspace deposit like the whole space is used by a single file
func calcDepositFeeForUserSpace(userspace *UserSpace, setting *FsSetting, currentHeight uint32) *StorageFee {
	uploadOpt := &UploadOption{
		FileSize:      userspace.Used + userspace.Remain,
		ProveInterval: setting.DefaultProvePeriod,
		ExpiredHeight: userspace.ExpireHeight,
		CopyNum:       setting.DefaultCopyNum,
	}

	return calcDepositFee(uploadOpt, setting, currentHeight)
}

func calcProveTimesByUploadInfo(uploadInfo *UploadOption, beginHeight uint32) uint64 {
	return (uploadInfo.ExpiredHeight-uint64(beginHeight))/uploadInfo.ProveInterval + 1
}

func calcProveTimesByFileInfo(fileInfo *FileInfo, beginHeight uint32) uint64 {
	uploadOpt := &UploadOption{
		ProveInterval: fileInfo.ProveInterval,
		ExpiredHeight: fileInfo.ExpiredHeight,
	}
	return calcProveTimesByUploadInfo(uploadOpt, beginHeight)
}

func calcFee(setting *FsSetting, proveTime, copyNum, fileSize, duration uint64) *StorageFee {
	validFee := calcValidFee(setting, proveTime, copyNum, fileSize)
	storageFee := calcStorageFee(setting, copyNum, fileSize, duration)

	log.Debugf("proveTime :%d, validFee :%d, storageFee: %d, duration: %d ", proveTime, validFee, storageFee, duration)

	return &StorageFee{
		ValidationFee: validFee,
		SpaceFee:      storageFee,
	}
}

func calcSingleValidFeeForFile(setting *FsSetting, fileSize uint64) uint64 {
	return uint64(float64(setting.GasForChallenge*fileSize) / float64(1024000))
}

func calcValidFeeForOneNode(setting *FsSetting, proveTime, fileSize uint64) uint64 {
	return proveTime * calcSingleValidFeeForFile(setting, fileSize)
}

func calcValidFee(setting *FsSetting, proveTime, copyNum, fileSize uint64) uint64 {
	return (copyNum + 1) * calcValidFeeForOneNode(setting, proveTime, fileSize)
}

func calcStorageFeeForOneNode(setting *FsSetting, fileSize, duration uint64) uint64 {
	return setting.GasPerGBPerBlock * fileSize * duration / uint64(1024000)
}

func calcStorageFee(setting *FsSetting, copyNum, fileSize, duration uint64) uint64 {
	return (copyNum + 1) * calcStorageFeeForOneNode(setting, fileSize, duration)
}

func calculateProfitForSettle(fileInfo *FileInfo, proveDetail *ProveDetail, fsSetting *FsSetting) uint64 {
	// first prove just indicate the whole file has been uploaded and dont calc for profit
	// copyNum pass 0 to calculate total fee for one node
	total := calcFee(fsSetting, proveDetail.ProveTimes-1, 0, fileInfo.FileBlockNum*fileInfo.FileBlockSize, fileInfo.ExpiredHeight-fileInfo.BlockHeight)
	log.Debugf("prove times: %d, block num: %d, block size: %d, expire height : %d, block height : %d, valid fee: %d, storage fee : %d\n",
		proveDetail.ProveTimes, fileInfo.FileBlockNum, fileInfo.FileBlockSize, fileInfo.ExpiredHeight, fileInfo.BlockHeight, total.ValidationFee, total.SpaceFee)

	return total.Sum()
}

func calculateNodePledge(fsNodeInfo *FsNodeInfo, fsSetting *FsSetting) uint64 {
	return fsSetting.FsGasPrice * fsSetting.GasPerGBPerBlock * fsNodeInfo.Volume
}
