package savefs

import (
	"bytes"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func setLastPunishmentHeightForNode(native *native.NativeService, nodeAddr common.Address, sectorID uint64, height uint64) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	key := GenFsNodeSectorPunishmentKey(contract, nodeAddr, sectorID)
	bf := new(bytes.Buffer)
	if err := utils.WriteVarUint(bf, height); err != nil {
		return errors.NewErr("write punishment height error")
	}
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}

func getLastPunishmentHeightForNode(native *native.NativeService, nodeAddr common.Address, sectorID uint64) (uint64, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	key := GenFsNodeSectorPunishmentKey(contract, nodeAddr, sectorID)
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return 0, errors.NewErr("GetPunishmentHeight GetStorageItem error!")
	}
	if item == nil {
		return 0, nil
	}

	height, err := utils.ReadVarUint(bytes.NewReader(item.Value))
	if err != nil {
		return 0, errors.NewErr("GetPunishmentHeight read height error!")
	}
	return height, nil
}

// when sector prove not ok or expired punish sector
func punishForSector(native *native.NativeService, sectorInfo *SectorInfo,
	nodeInfo *FsNodeInfo, fsSetting *FsSetting, times uint64) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	amount := times * calPunishmentForOneSectorProve(fsSetting, sectorInfo)

	log.Debugf("punish for sector, times %d, amount %d, node pledge %d", times, amount, nodeInfo.Pledge)

	if nodeInfo.Pledge >= amount {
		nodeInfo.Pledge -= amount
	} else {
		nodeInfo.Pledge = 0
		amount = nodeInfo.Pledge
	}

	if amount > 0 {
		err := appCallTransfer(native, utils.UsdtContractAddress, nodeInfo.WalletAddr, contract, amount)
		if err != nil {
			return errors.NewErr("[SectorProve] appCallTransfer, transfer error!")
		}

		if err := setFsNodeInfo(native, nodeInfo); err != nil {
			return errors.NewErr("[SectorProve] punishForSector setNodeInfo error!")
		}
	}

	err := setLastPunishmentHeightForNode(native, sectorInfo.NodeAddr, sectorInfo.SectorID, uint64(native.Height))
	if err != nil {
		return errors.NewErr("[CheckSectorProved] set lastPunishmentHeight for sector error!")
	}
	return nil
}

// when sector prove failed, we consider all files prove failed, so take sector.used to calculate profit
// and take it as punishment
func calPunishmentForOneSectorProve(fsSetting *FsSetting, sectorInfo *SectorInfo) uint64 {
	punishFactor := uint64(2)
	return punishFactor * calcSingleValidFeeForFile(fsSetting, sectorInfo.Used)
}

// calculate missing sector prove times
func calMissingSectorProveTimes(sectorInfo *SectorInfo, fsSetting *FsSetting,
	lastPunishHeight uint64, currHeight uint64) uint64 {
	interval := fsSetting.DefaultProvePeriod
	nextProveHeight := sectorInfo.NextProveHeight

	// prove not expired
	if nextProveHeight+interval >= currHeight {
		return 0
	}

	totalTimes := currHeight - nextProveHeight/interval

	var punishedTimes uint64
	if lastPunishHeight != 0 {
		if lastPunishHeight > nextProveHeight+interval {
			punishedTimes = (lastPunishHeight - nextProveHeight) / interval
		} else {
			// there is at least one success sector prove after last punish
			punishedTimes = 0
		}
	}

	// should not happen
	if totalTimes < punishedTimes {
		return 0
	}

	return totalTimes - punishedTimes
}
