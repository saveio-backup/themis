package savefs

import (
	"bytes"
	"math"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/usdt"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func FsManageUserSpace(native *native.NativeService) ([]byte, error) {
	log.Debugf("FsManageUserSpace height %d\n", native.Height)

	var userSpaceParams UserSpaceParams
	source := common.NewZeroCopySource(native.Input)
	if err := userSpaceParams.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] userSpaceParams deserialize error!")
	}
	if !native.ContextRef.CheckWitness(userSpaceParams.WalletAddr) {
		return utils.BYTE_FALSE, errors.NewErr("FS UserSpace] FsManageUserSpace CheckWitness failed!")
	}
	newUserSpace, state, updatedFiles, err := getUserspaceChange(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if state.Value > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, state.From, state.To, state.Value)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] FsManageUserSpace AppCallTransfer, transfer error!")
		}
	}
	// updated files
	for _, fi := range updatedFiles {
		if err = setFsFileInfo(native, fi); err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] FsManageUserSpace setFsFileInfo error:" + err.Error())
		}
	}
	// update userspace
	if err = setUserSpace(native, newUserSpace, userSpaceParams.Owner); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] FsManageUserSpace setUserSpace  error!")
	}

	log.Debugf("owner :%s, size %v, block count: %v\n",
		userSpaceParams.Owner.ToBase58(), userSpaceParams.Size, userSpaceParams.BlockCount)
	SetUserSpaceEvent(native, userSpaceParams.WalletAddr, userSpaceParams.Size.Type, userSpaceParams.Size.Value,
		userSpaceParams.BlockCount.Type, userSpaceParams.BlockCount.Value)

	return utils.BYTE_TRUE, nil
}

func FsGetUpdateCost(native *native.NativeService) ([]byte, error) {
	var userSpaceParams UserSpaceParams
	source := common.NewZeroCopySource(native.Input)
	if err := userSpaceParams.Deserialization(source); err != nil {
		return EncRet(false, []byte("[FS UserSpace] userSpaceParams deserialize error!")), nil
	}
	_, state, _, err := getUserspaceChange(native)
	if err != nil {
		log.Errorf("get user space change err %s", err)
		return EncRet(false, []byte(err.Error())), nil
	}
	bf := new(bytes.Buffer)
	err = state.Serialize(bf)
	if err != nil {
		log.Errorf("serial user space  err %s", err)
		return EncRet(false, []byte("[FS UserSpace] FsGetUpdateCost Serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

func FsGetUserSpace(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	walletAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[FS UserSpace] FsGetUserSpace DecodeAddress error!")), nil
	}

	userspace, err := getUserSpace(native, walletAddr)
	if err != nil {
		return EncRet(false, []byte("[FS UserSpace] FsGetUserSpace GetUserSpace error!")), nil
	}

	bf := new(bytes.Buffer)
	err = userspace.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[FS Userspace] FsGetUserSpace userspace serialize error!")), nil
	}
	return EncRet(true, bf.Bytes()), nil
}

func getUserspaceChange(native *native.NativeService) (*UserSpace, *usdt.State, []*FileInfo, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	currentHeight := uint64(native.Height)

	var userSpaceParams UserSpaceParams
	source := common.NewZeroCopySource(native.Input)
	if err := userSpaceParams.Deserialization(source); err != nil {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] userSpaceParams deserialize error!")
	}
	log.Debugf("change user space wallet addr: %v", userSpaceParams.WalletAddr.ToBase58())
	// nothing happens
	if userSpaceParams.Size.Value == 0 && userSpaceParams.BlockCount.Value == 0 {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] nothing happen")
	}

	fileList, err := GetFsFileList(native, userSpaceParams.Owner)
	if err != nil {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] GetFsFileList error")
	}
	log.Debugf("get file list len %v", len(fileList.List))
	// precheck, want to revoke size or time
	wantToRevoke := (UserSpaceType(userSpaceParams.Size.Type) == UserSpaceRevoke ||
		UserSpaceType(userSpaceParams.BlockCount.Type) == UserSpaceRevoke)
	if wantToRevoke && fileList.FileNum > 0 {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] can't revoke, there exists files")
	}

	fsSetting, err := getFsSetting(native)
	if err != nil {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] getFsSetting error!")
	}
	userSpaceKey := GenFsUserSpaceKey(contract, userSpaceParams.Owner)
	item, err := utils.GetStorageItem(native, userSpaceKey)
	if err != nil {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] GetStorageItem error!")
	}
	var oldUserspace *UserSpace
	// check old userspace
	notFound := (item == nil || len(item.Value) == 0)
	if !notFound {
		oldUserspace = &UserSpace{}
		reader := bytes.NewReader(item.Value)
		if err = oldUserspace.Deserialize(reader); err != nil {
			return nil, nil, nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS UserSpace] Set deserialize error!")
		}
		if oldUserspace.ExpireHeight <= currentHeight {
			oldUserspace.Used = 0
			oldUserspace.Remain = 0
			oldUserspace.Balance = 0
			oldUserspace.ExpireHeight = currentHeight
			oldUserspace.UpdateHeight = currentHeight
		}
	}

	// first operate user space or operate a expired space
	if notFound || oldUserspace.ExpireHeight == currentHeight {
		if wantToRevoke {
			return nil, nil, nil, errors.NewErr("[FS UserSpace]  no user space to revoke")
		}
		if userSpaceParams.Size.Value == 0 || userSpaceParams.BlockCount.Value == 0 {
			return nil, nil, nil,
				errors.NewErr("[FS UserSpace]  size and block count should both bigger than 0 first time")
		}
		if userSpaceParams.BlockCount.Value < fsSetting.DefaultProvePeriod {
			return nil, nil, nil,
				errors.NewErr("[FS UserSpace]  block count too small at first purchase user space!")
		}
	}
	if oldUserspace != nil {
		log.Debugf("oldUserspace.used: %d, remain: %d, expired: %d, balance: %d",
			oldUserspace.Used, oldUserspace.Remain, oldUserspace.ExpireHeight, oldUserspace.Balance)
	} else {
		log.Debugf("oldUserspace not found")
	}
	if wantToRevoke && (userSpaceParams.WalletAddr.ToBase58() != userSpaceParams.Owner.ToBase58()) {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] can't revoke other user space")
	}
	var newUserSpace *UserSpace
	var updatedFiles []*FileInfo
	var transferIn, transferOut uint64
	if UserSpaceType(userSpaceParams.Size.Type) == UserSpaceNone &&
		UserSpaceType(userSpaceParams.BlockCount.Type) == UserSpaceNone {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] nothing happen")
	}
	if UserSpaceType(userSpaceParams.Size.Type) != UserSpaceRevoke &&
		UserSpaceType(userSpaceParams.BlockCount.Type) != UserSpaceRevoke {
		// both add  01,10,11
		us, amount, updated, err := fsAddUserSpace(native, oldUserspace, userSpaceParams.Size.Value,
			userSpaceParams.BlockCount.Value, currentHeight, fsSetting, fileList)
		if err != nil {
			return nil, nil, nil, err
		}
		newUserSpace = us
		transferIn = amount
		updatedFiles = updated
	} else if UserSpaceType(userSpaceParams.Size.Type) != UserSpaceAdd &&
		UserSpaceType(userSpaceParams.BlockCount.Type) != UserSpaceAdd {
		// both revoke 02,20,22
		us, amount, err := fsRevokeUserspace(oldUserspace, userSpaceParams.Size.Value,
			userSpaceParams.BlockCount.Value, currentHeight, fsSetting)
		if err != nil {
			return nil, nil, nil, err
		}
		newUserSpace = us
		transferOut = amount
	} else {
		// one add && one revoke
		// 12, 21
		addedSize, addedBlockCount, revokeSize, revokeBlockcount := uint64(0), uint64(0), uint64(0), uint64(0)
		if UserSpaceType(userSpaceParams.Size.Type) == UserSpaceAdd &&
			UserSpaceType(userSpaceParams.BlockCount.Type) == UserSpaceRevoke {
			addedSize, addedBlockCount = userSpaceParams.Size.Value, 0
			revokeSize, revokeBlockcount = 0, userSpaceParams.BlockCount.Value
		} else if UserSpaceType(userSpaceParams.Size.Type) == UserSpaceRevoke &&
			UserSpaceType(userSpaceParams.BlockCount.Type) == UserSpaceAdd {
			addedSize, addedBlockCount = 0, userSpaceParams.BlockCount.Value
			revokeSize, revokeBlockcount = userSpaceParams.Size.Value, 0
		} else {
			return nil, nil, nil, errors.NewErr("[FS UserSpace] FsManageUserSpace Unknown type !")
		}
		us, addedAmount, update, err := fsAddUserSpace(native, oldUserspace, addedSize,
			addedBlockCount, currentHeight, fsSetting, fileList)
		if err != nil {
			return nil, nil, nil, err
		}
		us2, revokedAmount, err := fsRevokeUserspace(us, revokeSize, revokeBlockcount, currentHeight, fsSetting)
		if err != nil {
			return nil, nil, nil, err
		}
		newUserSpace = us2
		transferIn = addedAmount
		transferOut = revokedAmount
		updatedFiles = update
	}
	newUserSpace.UpdateHeight = uint64(native.Height)
	log.Debugf("transfer in %d, out: %d, newuserspace.used: %d, remain: %d, expired: %d, balance: %d, height: %d",
		transferIn, transferOut,
		newUserSpace.Used, newUserSpace.Remain, newUserSpace.ExpireHeight,
		newUserSpace.Balance, newUserSpace.UpdateHeight)
	// transfer state
	state := &usdt.State{}
	if transferIn >= transferOut {
		state.From = userSpaceParams.WalletAddr
		state.To = contract
		state.Value = transferIn - transferOut
	} else {
		state.From = contract
		state.To = userSpaceParams.WalletAddr
		state.Value = transferOut - transferIn
	}
	return newUserSpace, state, updatedFiles, nil
}

func fsAddUserSpace(native *native.NativeService, oldUserspace *UserSpace,
	addSize, addBlockCount, currentHeight uint64, fsSetting *FsSetting, fileList *FileList) (
	*UserSpace, uint64, []*FileInfo, error) {
	// new slice for updated file infos
	updatedFiles := make([]*FileInfo, 0)
	newUserSpace := &UserSpace{}
	if oldUserspace == nil {
		oldUserspace = &UserSpace{
			Used:         0,
			Remain:       0,
			ExpireHeight: currentHeight,
			Balance:      0,
		}
	}
	log.Debugf("add user space: old.used:%d, remain:%d, expired:%d, balance:%d, updated:%d",
		oldUserspace.Used, oldUserspace.Remain, oldUserspace.ExpireHeight,
		oldUserspace.Balance, oldUserspace.UpdateHeight)
	// calculate added challenge times
	addedTimes := uint64(math.Ceil(float64(addBlockCount) / float64(fsSetting.DefaultProvePeriod)))

	// calculate remain chanllenge times
	remainTimes := uint64(0)
	if oldUserspace.ExpireHeight > currentHeight {
		remainTimes = uint64(math.Ceil(
			float64(oldUserspace.ExpireHeight-currentHeight) / float64(fsSetting.DefaultProvePeriod)))
	}
	// calculate renew fee
	spaceSize := oldUserspace.Remain + addSize
	total := calcFee(fsSetting, addedTimes+remainTimes, fsSetting.DefaultCopyNum, spaceSize,
		fsSetting.DefaultProvePeriod*(addedTimes+remainTimes))
	log.Debugf("addedTimes %d addblockCount %d  default %d, spaceSize %d, remainTimes %d, total %v",
		addedTimes, addBlockCount, fsSetting.DefaultProvePeriod,
		spaceSize, remainTimes, total)
	deposit := uint64(0)
	// if renew fee large than balance, user need to deposit
	if total.ValidationFee+total.TxnFee+total.SpaceFee > oldUserspace.Balance {
		deposit = (total.ValidationFee + total.TxnFee + total.SpaceFee) - oldUserspace.Balance
	}
	newUserSpace.Used = oldUserspace.Used
	newUserSpace.Remain = oldUserspace.Remain + addSize
	newExpiredHeight := oldUserspace.ExpireHeight + addBlockCount
	newUserSpace.ExpireHeight = newExpiredHeight
	newUserSpace.Balance = oldUserspace.Balance + deposit
	// calulate challenge times for extend space, because of userSpaceParams.Size  > 0
	log.Debugf("newExpiredHeight %d, balance %d, deposit %d",
		newUserSpace.ExpireHeight, newUserSpace.Balance, deposit)
	if addBlockCount == 0 || fileList.FileNum == 0 {
		return newUserSpace, deposit, updatedFiles, nil
	}
	// need update files
	// find all file and update challenge times;
	for _, storedFileHash := range fileList.List {
		fileInfo, err := getFsFileInfo(native, storedFileHash.Hash)
		if err != nil {
			return nil, 0, nil, errors.NewErr("[FS UserSpace] FsManageUserSpace getFsFileInfo error")
		}
		if fileInfo.StorageType != FileStorageTypeUseSpace {
			continue
		}
		if newExpiredHeight <= fileInfo.ExpiredHeight {
			// origin stored file info has exists
			continue
		}
		// renew file
		renewDuration := newExpiredHeight - fileInfo.ExpiredHeight
		renewTimes := uint64(math.Ceil(float64(renewDuration) / float64(fileInfo.ProveInterval)))
		fileInfo.ExpiredHeight = newExpiredHeight
		updatedFiles = append(updatedFiles, fileInfo)
		totalDeposit := calcFee(fsSetting, fileInfo.ProveTimes+renewTimes, fileInfo.CopyNum,
			fileInfo.FileBlockNum*fileInfo.FileBlockSize, (fileInfo.ProveTimes+renewTimes)*fileInfo.ProveInterval)
		oldDeposit := calcFee(fsSetting, fileInfo.ProveTimes, fileInfo.CopyNum,
			fileInfo.FileBlockNum*fileInfo.FileBlockSize, fileInfo.ProveTimes*fileInfo.ProveInterval)
		if totalDeposit.Sum() < oldDeposit.Sum() {
			return nil, 0, nil, errors.NewErr("[FS UserSpace] FsManageUserSpace renew file new amount failed")
		}
		renewAmount := totalDeposit.Sum() - oldDeposit.Sum()
		deposit += renewAmount
		log.Debugf("file %s origin expired height %d, new expired height %d, "+
			"prove interval %d, fileSize %d, renew %d, new deposit %d",
			storedFileHash.Hash, fileInfo.ExpiredHeight,
			newExpiredHeight, fileInfo.ProveInterval, fileInfo.FileBlockNum*fileInfo.FileBlockSize,
			renewAmount, deposit)
	}
	return newUserSpace, deposit, updatedFiles, nil
}

func fsRevokeUserspace(oldUserspace *UserSpace, revokeSize, revokeBlockCount, currentHeight uint64,
	fsSetting *FsSetting) (*UserSpace, uint64, error) {
	if oldUserspace.Remain < revokeSize {
		return nil, 0, errors.NewErr("[FS UserSpace] FsManageUserSpace no enough remain space to revoke!")
	}
	if oldUserspace.ExpireHeight-revokeBlockCount < currentHeight {
		return nil, 0, errors.NewErr("[FS UserSpace] FsManageUserSpace revoke too much block count!")
	}

	// calulate challenge times for extend space, because of userSpaceParams.Size  > 0
	remainTimes := uint64(math.Ceil(
		float64(oldUserspace.ExpireHeight-revokeBlockCount-currentHeight) / float64(fsSetting.DefaultProvePeriod)))
	remainBalance := calcFee(fsSetting, remainTimes, fsSetting.DefaultCopyNum,
		(oldUserspace.Remain - revokeSize), fsSetting.DefaultProvePeriod*remainTimes)
	log.Debugf("remain times %d, expired %d, current %d, remain balance  %d, total balance %d",
		remainTimes, oldUserspace.ExpireHeight, currentHeight, remainBalance.Sum(), oldUserspace.Balance)
	amount := uint64(0)
	if oldUserspace.Balance > remainBalance.Sum() {
		amount = oldUserspace.Balance - remainBalance.Sum()
	}
	newUserSpace := &UserSpace{}
	newUserSpace.Used = oldUserspace.Used
	newUserSpace.Remain = oldUserspace.Remain - revokeSize
	newUserSpace.Balance = oldUserspace.Balance - amount
	newUserSpace.ExpireHeight = oldUserspace.ExpireHeight - revokeBlockCount
	return newUserSpace, amount, nil
}
