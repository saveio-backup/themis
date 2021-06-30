package savefs

import (
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/event"
	"github.com/saveio/themis/smartcontract/service/native"
)

const (
	EVENT_FS_STORE_FILE = iota + 1
	EVENT_FS_DELETE_FILE
	EVENT_FS_DELETE_FILES
	EVENT_FS_SET_USER_SPACE
	EVENT_FS_REG_NODE
	EVENT_FS_UN_REG_NODE
	EVENT_FS_PROVE_FILE
	EVENT_FS_FILE_PDP_SUCCESS
	EVENT_FS_CREATE_SECTOR
	EVENT_FS_DELETE_SECTOR
)

func StoreFileEvent(native *native.NativeService, fileHash []byte, fileSize uint64, walletAddr common.Address, cost uint64) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_STORE_FILE,
		"blockHeight": native.Height,
		"eventName":   "uploadFile",
		"fileHash":    string(fileHash),
		"fileSize":    fileSize,
		"walletAddr":  walletAddr.ToBase58(),
		"cost":        cost,
	}
	newEvent(native, EVENT_FS_STORE_FILE, []common.Address{walletAddr}, event)
}

func DeleteFileEvent(native *native.NativeService, fileHash []byte, walletAddr common.Address) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_DELETE_FILE,
		"blockHeight": native.Height,
		"eventName":   "deleteFile",
		"fileHash":    string(fileHash),
		"walletAddr":  walletAddr.ToBase58(),
	}
	newEvent(native, EVENT_FS_DELETE_FILE, []common.Address{walletAddr}, event)
}

func DeleteFilesEvent(native *native.NativeService, fileHashes []string, walletAddr common.Address) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_DELETE_FILES,
		"blockHeight": native.Height,
		"eventName":   "deleteFiles",
		"fileHashes":  fileHashes,
		"walletAddr":  walletAddr.ToBase58(),
	}
	newEvent(native, EVENT_FS_DELETE_FILES, []common.Address{walletAddr}, event)
}

func SetUserSpaceEvent(native *native.NativeService, walletAddr common.Address, sizeType, size, countType, count uint64) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_SET_USER_SPACE,
		"blockHeight": native.Height,
		"eventName":   "setUserSpace",
		"walletAddr":  walletAddr.ToBase58(),
		"sizeType":    sizeType,
		"size":        size,
		"countType":   countType,
		"count":       count,
	}
	newEvent(native, EVENT_FS_SET_USER_SPACE, []common.Address{walletAddr}, event)
}

func RegisterNodeEvent(native *native.NativeService, walletAddr common.Address, nodeAddr []byte, volume, serviceTime uint64) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_REG_NODE,
		"blockHeight": native.Height,
		"eventName":   "registerNode",
		"walletAddr":  walletAddr.ToBase58(),
		"nodeAdd":     string(nodeAddr),
		"volume":      volume,
		"serviceTime": serviceTime,
	}
	newEvent(native, EVENT_FS_REG_NODE, []common.Address{walletAddr}, event)
}

func UnRegisterNodeEvent(native *native.NativeService, walletAddr common.Address) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_UN_REG_NODE,
		"blockHeight": native.Height,
		"eventName":   "unregisterNode",
		"walletAddr":  walletAddr.ToBase58(),
	}
	newEvent(native, EVENT_FS_UN_REG_NODE, []common.Address{walletAddr}, event)
}

func ProveFileEvent(native *native.NativeService, fileHash []byte, walletAddr common.Address, profit uint64) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_PROVE_FILE,
		"blockHeight": native.Height,
		"eventName":   "proveFile",
		"walletAddr":  walletAddr.ToBase58(),
		"profit":      profit,
	}
	newEvent(native, EVENT_FS_PROVE_FILE, []common.Address{walletAddr}, event)
}

func FilePDPSuccessEvent(native *native.NativeService, fileHash []byte, walletAddr common.Address) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_FILE_PDP_SUCCESS,
		"blockHeight": native.Height,
		"eventName":   "filePdpSuccess",
		"fileHash":    string(fileHash),
		"walletAddr":  walletAddr.ToBase58(),
	}
	newEvent(native, EVENT_FS_FILE_PDP_SUCCESS, []common.Address{walletAddr}, event)
}

func CreateSectorEvent(native *native.NativeService, walletAddr common.Address, sectorId uint64, proveLevel uint64, size uint64) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_CREATE_SECTOR,
		"blockHeight": native.Height,
		"eventName":   "createSector",
		"walletAddr":  walletAddr.ToBase58(),
		"sectorId":    sectorId,
		"proveLevel":  proveLevel,
		"size":        size,
	}
	newEvent(native, EVENT_FS_CREATE_SECTOR, []common.Address{walletAddr}, event)
}

func DeleteSectorEvent(native *native.NativeService, walletAddr common.Address, sectorId uint64) {
	event := map[string]interface{}{
		"eventId":     EVENT_FS_DELETE_SECTOR,
		"blockHeight": native.Height,
		"eventName":   "deleteSector",
		"walletAddr":  walletAddr.ToBase58(),
		"sectorId":    sectorId,
	}
	newEvent(native, EVENT_FS_DELETE_SECTOR, []common.Address{walletAddr}, event)
}

func newEvent(srvc *native.NativeService, id uint32, participants []common.Address, st interface{}) {
	e := event.NotifyEventInfo{}
	e.ContractAddress = srvc.ContextRef.CurrentContext().ContractAddress
	e.EventIdentifier = id
	e.Addresses = append(e.Addresses, participants...)
	e.States = st
	srvc.Notifications = append(srvc.Notifications, &e)
	return
}
