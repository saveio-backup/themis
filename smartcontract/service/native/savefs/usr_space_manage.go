package savefs

import (
	"bytes"
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
		return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] FsManageUserSpace CheckWitness failed!")
	}
	newUserSpace, state, updatedFiles, err := getUserspaceChange(native, &userSpaceParams)
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

func NewFsManageUserSpace(native *native.NativeService) ([]byte, error) {
	log.Debugf("NewFsManageUserSpace height %d\n", native.Height)

	var userSpaceParams UserSpaceParams
	source := common.NewZeroCopySource(native.Input)
	if err := userSpaceParams.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] userSpaceParams deserialize error!")
	}
	if !native.ContextRef.CheckWitness(userSpaceParams.WalletAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] NewFsManageUserSpace CheckWitness failed!")
	}
	newUserSpace, state, err := newGetUserspaceChange(native, &userSpaceParams)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if state.Value > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, state.From, state.To, state.Value)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] NewFsManageUserSpace AppCallTransfer, transfer error!")
		}
	}
	// update userspace
	if err = setUserSpace(native, newUserSpace, userSpaceParams.Owner); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] NewFsManageUserSpace setUserSpace  error!")
	}

	log.Debugf("owner :%s, size %v, block count: %v\n",
		userSpaceParams.Owner.ToBase58(), userSpaceParams.Size, userSpaceParams.BlockCount)
	SetUserSpaceEvent(native, userSpaceParams.WalletAddr, userSpaceParams.Size.Type, userSpaceParams.Size.Value,
		userSpaceParams.BlockCount.Type, userSpaceParams.BlockCount.Value)

	return utils.BYTE_TRUE, nil
}
func FsCashUserSpace(native *native.NativeService) ([]byte, error) {
	log.Debug("FsCashUserSpace")
	source := common.NewZeroCopySource(native.Input)
	walletAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[FS UserSpace] FsCashUserSpace DecodeAddress error!")), nil
	}

	oldUserSpace,state, err := cashUserSpace(native, walletAddr)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if state.Value > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, state.From, state.To, state.Value)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] NewFsManageUserSpace AppCallTransfer, transfer error!")
		}
	}

	newUserSpace := &UserSpace{
		Used:         0,
		Remain:       0,
		ExpireHeight: uint64(native.Height) ,
		Balance:      0,
	}
	// update userspace
	if err = setUserSpace(native, newUserSpace, walletAddr); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] FsCashUserSpace setUserSpace  error!")
	}
	SetUserSpaceEvent(native, walletAddr, uint64(UserSpaceCash), oldUserSpace.Remain,
		uint64(UserSpaceCash), oldUserSpace.ExpireHeight - uint64(native.Height))

	return utils.BYTE_TRUE, nil
}
func FsGetUpdateCost(native *native.NativeService) ([]byte, error) {
	_, state, err := newGetUserspaceChange(native, nil)
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

func FsDeleteUserSpace(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	source := common.NewZeroCopySource(native.Input)
	walletAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[FS UserSpace] FsDeleteUserSpace DecodeAddress error!")), nil
	}

	if !native.ContextRef.CheckWitness(walletAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] FsDeleteUserSpace CheckWitness failed!")
	}

	userspace, err := getUserSpace(native, walletAddr)
	if err != nil {
		return EncRet(false, []byte("[FS UserSpace] FsGetUserSpace GetUserSpace error!")), nil
	}

	// allow to delete user space if no file in userspace or when user space has expired for at least one prove interval
	if userspace.Used == 0 && userspace.Balance > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract, walletAddr, userspace.Balance)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] FsDeleteUserSpace AppCallTransfer, transfer error!")
		}
	} else if userspace.ExpireHeight < uint64(native.Height) {
		err = deleteExpiredUserSpace(native, userspace, walletAddr)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] FsDeleteUserSpace deleteExpiredUserSpace error!")
		}
	} else {
		return utils.BYTE_FALSE, errors.NewErr("[FS UserSpace] FsDeleteUserSpace user space not expired!")
	}

	return utils.BYTE_TRUE, nil
}

func deleteExpiredUserSpace(native *native.NativeService, userspace *UserSpace, walletAddr common.Address) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fsSetting, err := getFsSetting(native)
	if err != nil {
		return errors.NewErr("[Fs UserSpace] deleteExpiredUserSpace getFsSetting error!")
	}

	// dont allow to delete userspace for one prove interval after expire in order to let fs server has enough time
	// to submit last file prove
	if fsSetting.DefaultProvePeriod+userspace.ExpireHeight > uint64(native.Height) {
		return errors.NewErr("[Fs UserSpace] cannot delete userspace when expired less than prove interval !")
	}

	// file list include files that no last pdp is submitted
	fileList, err := GetFsFileList(native, walletAddr)
	if err != nil {
		return errors.NewErr("[Fs UserSpace] deleteExpiredUserSpace GetFsFileList error!")
	}

	sType := []int{FileStorageTypeUseSpace}
	deletedFiles, amount, err := deleteExpiredFilesFromList(native, fileList, walletAddr, sType)
	if err != nil {
		return errors.NewErr("[Fs UserSpace] deleteExpiredUserSpace deleteExpiredFilesFromList error!")
	}
	log.Debugf("deleteExpiredUserSpace for %s from fileList, deletedFiles count %d, amount %d",
		walletAddr, len(deletedFiles), amount)

	unsettledList, err := GetFsUnSettledList(native, walletAddr)
	if err != nil {
		return errors.NewErr("[Fs UserSpace] deleteExpiredUserSpace GetUserUnSettledFileList error!")
	}

	deletedFiles, amount, err = deleteExpiredFilesFromList(native, unsettledList, walletAddr, sType)
	if err != nil {
		return errors.NewErr("[Fs UserSpace] deleteExpiredUserSpace deleteExpiredFilesFromList error!")
	}

	log.Debugf("deleteExpiredUserSpace for %s from unSettleList, deletedFiles count %d, amount %d",
		walletAddr, len(deletedFiles), amount)

	for _, fileHash := range deletedFiles {
		unsettledList.Del(fileHash)
	}

	err = setFsFileList(native, GenFsUnSettledListKey(contract, walletAddr), unsettledList)
	if err != nil {
		return errors.NewErr("[Fs UserSpace] deleteExpiredUserSpace setFsFileList error!")
	}

	return nil
}

func getUserspaceChange(native *native.NativeService, userSpaceParams *UserSpaceParams) (*UserSpace, *usdt.State, []*FileInfo, error) {
	currentHeight := uint64(native.Height)

	if userSpaceParams == nil {
		var params UserSpaceParams
		source := common.NewZeroCopySource(native.Input)
		if err := params.Deserialization(source); err != nil {
			return nil, nil, nil, errors.NewErr("[FS UserSpace] userSpaceParams deserialize error!")
		}
		userSpaceParams = &params
	}

	fsSetting, err := getFsSetting(native)
	if err != nil {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] getFsSetting error!")
	}

	log.Debugf("change user space wallet addr: %v", userSpaceParams.WalletAddr.ToBase58())

	if err := checkUserSpaceParams(native, userSpaceParams); err != nil {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] checkUserSpaceParams error!")
	}

	oldUserspace, err := getOldUserSpace(native, userSpaceParams.Owner)
	if err != nil {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] getOldUserSpace error!")
	}
	if oldUserspace != nil && oldUserspace.ExpireHeight <= currentHeight {
		processExpiredUserSpace(oldUserspace, currentHeight)
	}

	// first operate user space or operate a expired space
	if oldUserspace == nil || oldUserspace.ExpireHeight == currentHeight {
		if err = checkForFirstUserSpaceOperation(fsSetting, userSpaceParams); err != nil {
			return nil, nil, nil, errors.NewErr("[FS UserSpace] checkForFirstUserSpaceOperation error!")
		}
	}

	return processForUserSpaceOperations(native, userSpaceParams, oldUserspace, fsSetting)
}
func newGetUserspaceChange(native *native.NativeService, userSpaceParams *UserSpaceParams) (*UserSpace, *usdt.State, error) {
	currentHeight := uint64(native.Height)

	if userSpaceParams == nil {
		var params UserSpaceParams
		source := common.NewZeroCopySource(native.Input)
		if err := params.Deserialization(source); err != nil {
			return nil, nil, errors.NewErr("[FS UserSpace] userSpaceParams deserialize error!")
		}
		userSpaceParams = &params
	}

	fsSetting, err := getFsSetting(native)
	if err != nil {
		return nil, nil, errors.NewErr("[FS UserSpace] getFsSetting error!")
	}

	log.Debugf("change user space wallet addr: %v", userSpaceParams.WalletAddr.ToBase58())

	if err := checkUserSpaceParams(native, userSpaceParams); err != nil {
		return nil, nil, errors.NewErr("[FS UserSpace] checkUserSpaceParams error!")
	}

	oldUserspace, err := getOldUserSpace(native, userSpaceParams.Owner)


	if err != nil {
		return nil, nil, errors.NewErr("[FS UserSpace] getOldUserSpace error!")
	}
	// 原来空间已经过期
	if oldUserspace != nil && oldUserspace.ExpireHeight <= currentHeight {
		//更新旧空间为当前
		processExpiredUserSpace(oldUserspace, currentHeight)
	}

	// first operate user space or operate a expired space
	if oldUserspace == nil || oldUserspace.ExpireHeight == currentHeight {
		// 检查创建参数
		if err = checkForFirstUserSpaceOperation(fsSetting, userSpaceParams); err != nil {
			return nil, nil, errors.NewErr("[FS UserSpace] checkForFirstUserSpaceOperation error!")
		}
	}
	//增加空间
	return newProcessForUserSpaceOperations(native, userSpaceParams, oldUserspace, fsSetting)
}
func cashUserSpace(native *native.NativeService, address common.Address) (*UserSpace, *usdt.State, error) {

	oldUserspace,err := getOldUserSpace(native, address)
	if err != nil {
		return nil,nil, errors.NewErr("[FS UserSpace] getOldUserSpace error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	fsSetting, err := getFsSetting(native)
	if err != nil {
		return nil, nil, errors.NewErr("[FS UserSpace] getFsSetting error!")
	}
	currentHeight := uint64(native.Height)
	//获取到余额
	fee := newCalcFee(CashSpace,oldUserspace,fsSetting, fsSetting.DefaultCopyNum, 0, 0,currentHeight)

	// transfer state
	state := &usdt.State{}
		state.From = contract
		state.To = address
		state.Value = fee.Sum()

	return oldUserspace, state, nil
}

func checkUserSpaceParams(native *native.NativeService, userSpaceParams *UserSpaceParams) error {
	if userSpaceParams.Size.Value == 0 && userSpaceParams.BlockCount.Value == 0 {
		return errors.NewErr("[FS UserSpace] nothing happen")
	}

	ops, err := getUserSpaceOperationsFromParams(userSpaceParams)
	if err != nil {
		return err
	}

	if ops == UserSpaceOps_None_None {
		return errors.NewErr("[FS UserSpace] nothing happen")
	}

	// check for revoke space
	if isRevokeUserSpace(userSpaceParams) {
		if err = checkForUserSpaceRevoke(native, userSpaceParams); err != nil {
			return errors.NewErr("[FS UserSpace] checkForUserSpaceRevoke error")
		}
	}
	return nil
}
func checkForFirstUserSpaceOperation(fsSetting *FsSetting, userSpaceParams *UserSpaceParams) error {
	ops, err := getUserSpaceOperationsFromParams(userSpaceParams)
	if err != nil {
		return err
	}

	if ops != UserspaceOps_Add_Add {
		return errors.NewErr("[FS UserSpace] nothing happen")
	}

	if userSpaceParams.Size.Value == 0 || userSpaceParams.BlockCount.Value == 0 {
		return errors.NewErr("[FS UserSpace]  size and block count should both bigger than 0 first time")
	}
	if userSpaceParams.BlockCount.Value < fsSetting.DefaultProvePeriod {
		return errors.NewErr("[FS UserSpace]  block count too small at first purchase user space!")
	}
	return nil
}

func checkForUserSpaceRevoke(native *native.NativeService, userSpaceParams *UserSpaceParams) error {
	fileList, err := GetFsFileList(native, userSpaceParams.Owner)
	if err != nil {
		return errors.NewErr("[FS UserSpace] GetFsFileList error")
	}

	if fileList.FileNum > 0 {
		return errors.NewErr("[FS UserSpace] can't revoke, there exists files")
	}

	if userSpaceParams.WalletAddr.ToBase58() != userSpaceParams.Owner.ToBase58() {
		return errors.NewErr("[FS UserSpace] can't revoke other user space")
	}
	return nil
}

func processExpiredUserSpace(userSpace *UserSpace, currentHeight uint64) {
	// we consider expired user space is no longer available to user
	userSpace.Used = 0
	userSpace.Remain = 0
	// TODO :how should handle remaining balance?
	//userSpace.Balance = 0
	userSpace.ExpireHeight = currentHeight
	userSpace.UpdateHeight = currentHeight
}

func processForUserSpaceOperations(native *native.NativeService, userSpaceParams *UserSpaceParams,
	oldUserspace *UserSpace, fsSetting *FsSetting) (*UserSpace, *usdt.State, []*FileInfo, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var newUserSpace *UserSpace
	var updatedFiles []*FileInfo
	var transferIn, transferOut uint64
	var err error

	currentHeight := uint64(native.Height)

	if oldUserspace != nil {
		log.Debugf("oldUserspace.used: %d, remain: %d, expired: %d, balance: %d",
			oldUserspace.Used, oldUserspace.Remain, oldUserspace.ExpireHeight, oldUserspace.Balance)
	} else {
		log.Debugf("oldUserspace not found")
	}

	fileList, err := GetFsFileList(native, userSpaceParams.Owner)
	if err != nil {
		return nil, nil, nil, errors.NewErr("[FS UserSpace] GetFsFileList error")
	}

	userSpaceOps, _ := getUserSpaceOperationsFromParams(userSpaceParams)
	switch userSpaceOps {
	// at least one add, no revoke
	case UserspaceOps_Add_Add, UserspaceOps_Add_None, UserspaceOps_None_Add:
		newUserSpace, transferIn, updatedFiles, err = fsAddUserSpace(native, oldUserspace, userSpaceParams.Size.Value,
			userSpaceParams.BlockCount.Value, currentHeight, fsSetting, fileList)
		if err != nil {
			return nil, nil, nil, err
		}
		// at least one revoke no add
	case UserspaceOps_Revoke_Revoke, UserspaceOps_None_Revoke, UserspaceOps_Revoke_None:
		newUserSpace, transferOut, err = fsRevokeUserspace(oldUserspace, userSpaceParams.Size.Value,
			userSpaceParams.BlockCount.Value, currentHeight, fsSetting)
		if err != nil {
			return nil, nil, nil, err
		}
	case UserspaceOps_Add_Revoke, UserspaceOps_Revoke_Add:
		newUserSpace, transferIn, transferOut, updatedFiles, err = processForUserSpaceOneAddOneRevoke(native,
			userSpaceParams, oldUserspace, fsSetting, fileList, userSpaceOps)
	default:
		return nil, nil, nil, errors.NewErr("invalid userspace operation")
	}

	if newUserSpace == nil {
		return nil, nil, nil, errors.NewErr("new user space is nil")
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
func newProcessForUserSpaceOperations(native *native.NativeService, userSpaceParams *UserSpaceParams,
	oldUserspace *UserSpace, fsSetting *FsSetting) (*UserSpace, *usdt.State, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var newUserSpace *UserSpace
	var transferIn, transferOut uint64
	var err error

	currentHeight := uint64(native.Height)

	if oldUserspace != nil {
		log.Debugf("oldUserspace.used: %d, remain: %d, expired: %d, balance: %d",
			oldUserspace.Used, oldUserspace.Remain, oldUserspace.ExpireHeight, oldUserspace.Balance)
	} else {
		log.Debugf("oldUserspace not found")
	}

	userSpaceOps, _ := getUserSpaceOperationsFromParams(userSpaceParams)
	switch userSpaceOps {
	// at least one add, no revoke
	case UserspaceOps_Add_Add, UserspaceOps_Add_None, UserspaceOps_None_Add:
		newUserSpace, transferIn, err = newFsAddUserSpace(oldUserspace, userSpaceParams.Size.Value,
			userSpaceParams.BlockCount.Value, currentHeight, fsSetting)
		if err != nil {
			return nil, nil, err
		}
		// at least one revoke no add
	case UserspaceOps_Revoke_Revoke, UserspaceOps_None_Revoke, UserspaceOps_Revoke_None:
		return nil, nil, errors.NewErr("userspace revoke revoke function not implemented")
	case UserspaceOps_Add_Revoke, UserspaceOps_Revoke_Add:
		return nil, nil, errors.NewErr("userspace add revoke function not implemented")
	default:
		return nil, nil, errors.NewErr("invalid userspace operation")
	}

	if newUserSpace == nil {
		return nil, nil, errors.NewErr("new user space is nil")
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
	return newUserSpace, state, nil
}

func processForUserSpaceOneAddOneRevoke(native *native.NativeService, userSpaceParams *UserSpaceParams,
	oldUserspace *UserSpace, fsSetting *FsSetting, fileList *FileList, ops uint64) (*UserSpace, uint64, uint64, []*FileInfo, error) {
	currentHeight := uint64(native.Height)
	var addedSize, addedBlockCount, revokedSize, revokedBlockCount uint64

	switch ops {
	case UserspaceOps_Add_Revoke:
		addedSize = userSpaceParams.Size.Value
		revokedBlockCount = userSpaceParams.BlockCount.Value
	case UserspaceOps_Revoke_Add:
		revokedSize = userSpaceParams.Size.Value
		addedBlockCount = userSpaceParams.BlockCount.Value
	}

	us, addedAmount, update, err := fsAddUserSpace(native, oldUserspace, addedSize,
		addedBlockCount, currentHeight, fsSetting, fileList)
	if err != nil {
		return nil, 0, 0, nil, err
	}

	us2, revokedAmount, err := fsRevokeUserspace(us, revokedSize, revokedBlockCount, currentHeight, fsSetting)
	if err != nil {
		return nil, 0, 0, nil, err
	}

	return us2, addedAmount, revokedAmount, update, nil
}

func fsAddUserSpace(native *native.NativeService, oldUserspace *UserSpace,
	addSize, addBlockCount, currentHeight uint64, fsSetting *FsSetting, fileList *FileList) (
	*UserSpace, uint64, []*FileInfo, error) {

	// create user space
	if oldUserspace == nil {
		newUserSpace := &UserSpace{
			Used:         0,
			Remain:       addSize,
			ExpireHeight: currentHeight + addBlockCount,
			Balance:      0,
		}

		fee := calcDepositFeeForUserSpace(newUserSpace, fsSetting, uint32(currentHeight))
		newUserSpace.Balance = fee.Sum()

		return newUserSpace, fee.Sum(), nil, nil
	} else {
		log.Debugf("add user space: old.used:%d, remain:%d, expired:%d, balance:%d, updated:%d",
			oldUserspace.Used, oldUserspace.Remain, oldUserspace.ExpireHeight,
			oldUserspace.Balance, oldUserspace.UpdateHeight)

		updatedFiles := make([]*FileInfo, 0)

		newExpiredHeight := oldUserspace.ExpireHeight + addBlockCount
		newRemain := oldUserspace.Remain + addSize

		//  fee from now to expire height
		fee1 := calcDepositFeeForUserSpace(oldUserspace, fsSetting, uint32(currentHeight))

		newUserSpace := &UserSpace{
			Used:         oldUserspace.Used,
			Remain:       newRemain,
			ExpireHeight: newExpiredHeight,
			Balance:      oldUserspace.Balance,
		}
		// fee from now to new expire height with added size
		fee2 := calcDepositFeeForUserSpace(newUserSpace, fsSetting, uint32(currentHeight))

		if fee2.Sum() <= fee1.Sum() {
			log.Errorf("invalid fee for addUserSpace: fee1 %d, fee2 %d", fee1.Sum(), fee2.Sum())
			return nil, 0, nil, errors.NewErr("[FS Userspace] invalid fee for addUserSpace")
		}

		deposit := fee2.Sum() - fee1.Sum()

		// we calculate the userspace fee by assuming a file with the same size of user space from the beginning of
		// user space creation, but as block height grow, the deposit for a file with same size will decrease
		// so there might be enough remaining balance for the new added size/block count.

		// set used as 0 to do calculation for remaining size
		oldUserspace.Used = 0
		feeForRemaining := calcDepositFeeForUserSpace(oldUserspace, fsSetting, uint32(currentHeight))

		if oldUserspace.Balance <= feeForRemaining.Sum() {
			log.Errorf("invalid remaining fee for addUserSpace: fee %d", feeForRemaining.Sum())
			return nil, 0, nil, errors.NewErr("[FS Userspace] invalid remaining fee for addUserSpace")
		}

		availableInBalance := oldUserspace.Balance - feeForRemaining.Sum()
		if availableInBalance > deposit {
			// enough in balance for added size/block count, no need deposit
			deposit = 0
		} else {
			deposit = deposit - availableInBalance
			newUserSpace.Balance += deposit
		}

		// find all file and update challenge times and deposit when add block count
		if addBlockCount != 0 {
			var err error
			updatedFiles, err = updateFilesForRenew(native, fileList, fsSetting, newExpiredHeight)
			if err != nil {
				return nil, 0, nil, errors.NewErr("[FS UserSpace] updateFilesForRenew error")
			}
		}

		return newUserSpace, deposit, updatedFiles, nil
	}
}
func newFsAddUserSpace(oldUserspace *UserSpace,
	addSize, addBlockCount, currentHeight uint64, fsSetting *FsSetting) (
	*UserSpace, uint64, error) {
	// create user space
	if oldUserspace == nil {
		newUserSpace,_:= newCalcDepositFeeForUserSpace(nil,addSize,addBlockCount, fsSetting, uint32(currentHeight))
		return newUserSpace, newUserSpace.Balance, nil
	} else {
		log.Debugf("add user space: old.used:%d, remain:%d, expired:%d, balance:%d, updated:%d",
			oldUserspace.Used, oldUserspace.Remain, oldUserspace.ExpireHeight,
			oldUserspace.Balance, oldUserspace.UpdateHeight)
		log.Debugf("addSize:%d addBlockCount:%d",addSize,addBlockCount)
		// fee from now to new expire height with added size
		newUserSpace,deposit:= newCalcDepositFeeForUserSpace(oldUserspace,addSize,addBlockCount, fsSetting, uint32(currentHeight))
		log.Debugf("new user space: new.used:%d, remain:%d, expired:%d, balance:%d, updated:%d",
			newUserSpace.Used, newUserSpace.Remain, newUserSpace.ExpireHeight,
			newUserSpace.Balance, newUserSpace.UpdateHeight)
		return newUserSpace, deposit, nil
	}
}

func updateFilesForRenew(native *native.NativeService, fileList *FileList,
	fsSetting *FsSetting, newExpireHeight uint64) ([]*FileInfo, error) {
	updatedFiles := make([]*FileInfo, 0)

	for _, fileHash := range fileList.List {
		fileInfo, err := getFsFileInfo(native, fileHash.Hash)
		if err != nil {
			return nil, errors.NewErr("[FS UserSpace] FsManageUserSpace getFsFileInfo error")
		}
		if fileInfo.StorageType != FileStorageTypeUseSpace {
			continue
		}
		if newExpireHeight <= fileInfo.ExpiredHeight {
			// origin stored file info has exists
			continue
		}

		if err = updateFileInfoForRenew(fsSetting, newExpireHeight, fileInfo); err != nil {
			return nil, errors.NewErr("[FS UserSpace] updateFileInfoForRenew error")
		}
		updatedFiles = append(updatedFiles, fileInfo)

		log.Debugf("file %s origin expired height %d, new expired height %d, "+
			"prove interval %d, fileSize %d,new deposit %d",
			fileHash.Hash, fileInfo.ExpiredHeight, newExpireHeight,
			fileInfo.ProveInterval, fileInfo.FileBlockNum*fileInfo.FileBlockSize, fileInfo.Deposit)
	}
	return updatedFiles, nil
}

func updateFileInfoForRenew(fsSetting *FsSetting, newExpireHeight uint64, fileInfo *FileInfo) error {
	fileInfo.ExpiredHeight = newExpireHeight

	uploadOpt := &UploadOption{
		ExpiredHeight: fileInfo.ExpiredHeight,
		ProveInterval: fileInfo.ProveInterval,
		CopyNum:       fileInfo.CopyNum,
		FileSize:      fileInfo.FileBlockSize * fileInfo.FileBlockNum,
	}

	beginHeight := uint32(fileInfo.BlockHeight)

	// use block height for storeFile and new expire height for new deposit calc
	newDeposit := calcDepositFee(uploadOpt, fsSetting, beginHeight)

	if newDeposit.Sum() <= fileInfo.Deposit {
		log.Errorf("updateFileInfoForRenew, new deposit %d, orig deposit %d", newDeposit.Sum(), fileInfo.Deposit)
		return errors.NewErr("[FS UserSpace] new deposit is not larger than old value !")
	}

	fileInfo.Deposit = newDeposit.Sum()
	fileInfo.ProveTimes = calcProveTimesByUploadInfo(uploadOpt, beginHeight)
	return nil
}

func fsRevokeUserspace(oldUserspace *UserSpace, revokeSize, revokeBlockCount, currentHeight uint64,
	fsSetting *FsSetting) (*UserSpace, uint64, error) {
	if oldUserspace.Remain < revokeSize {
		return nil, 0, errors.NewErr("[FS UserSpace] FsManageUserSpace no enough remain space to revoke!")
	}
	if oldUserspace.ExpireHeight-revokeBlockCount < currentHeight {
		return nil, 0, errors.NewErr("[FS UserSpace] FsManageUserSpace revoke too much block count!")
	}

	newUserSpace := &UserSpace{
		Used:         oldUserspace.Used,
		Remain:       oldUserspace.Remain - revokeSize,
		ExpireHeight: oldUserspace.ExpireHeight - revokeBlockCount,
		Balance:      oldUserspace.Balance,
	}

	// fee from now to expire height
	fee1 := calcDepositFeeForUserSpace(oldUserspace, fsSetting, uint32(currentHeight))

	// fee from now to new expired height with revoked size
	fee2 := calcDepositFeeForUserSpace(newUserSpace, fsSetting, uint32(currentHeight))

	if fee1.Sum() <= fee2.Sum() {
		log.Errorf("invalid fee for revokeUserSpace: fee1 %d, fee2 %d", fee1.Sum(), fee2.Sum())
		return nil, 0, errors.NewErr("[FS Userspace] invalid fee for revokeUserSpace")
	}

	amount := fee1.Sum() - fee2.Sum()

	if newUserSpace.Balance < amount {
		log.Errorf("invalid balance for revokeUserSpace: balance %d, amount %d", newUserSpace.Balance, amount)
		return nil, 0, errors.NewErr("[FS Userspace] balance is smaller than revoked amount")
	}

	newUserSpace.Balance -= amount

	return newUserSpace, amount, nil
}
