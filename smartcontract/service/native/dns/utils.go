package dns

import (
	"bytes"
	"crypto/sha256"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/event"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/usdt"
	"github.com/saveio/themis/smartcontract/service/native/utils"

	"fmt"

	vbftconfig "github.com/saveio/themis/consensus/vbft/config"
	"github.com/saveio/themis/crypto/vrf"
)

const (
	EVENT_DNS_REG_NODE = iota + 1
	EVENT_DNS_UN_REG_NODE
	EVENT_DNS_QUIT_NODE
)

const (
	NAME        = "nameinfo"
	HEADER      = "headerinfo"
	ADMIN       = "admininfo"
	PLUGIN_LIST = "pluginlist"
)

func GenNameInfoKey(contract common.Address, header, url []byte) []byte {
	h := keyHash(header, url)
	key := append(contract[:], NAME...)
	key = append(key[:], h[:]...)
	return key
}

func GenHeaderKey(contract common.Address, header []byte) []byte {
	h := keyHash(header, nil)
	key := append(contract[:], HEADER...)
	key = append(key[:], h[:]...)
	return key
}

func GenPluginListKey(contract common.Address) []byte {
	key := append(contract[:], PLUGIN_LIST...)
	return key
}

func GetNameInfoItem(native *native.NativeService, keyHash []byte) (NameInfo, error) {
	item, err := utils.GetStorageItem(native, keyHash)
	ni := NameInfo{}
	if err != nil || item == nil {
		return ni, err
	}
	bf := bytes.NewBuffer(item.Value)
	err = ni.Deserialize(bf)
	return ni, err
}

func GetHeaderInfoItem(native *native.NativeService, keyHash []byte) (HeaderInfo, error) {
	item, err := utils.GetStorageItem(native, keyHash)
	ri := HeaderInfo{}
	if err != nil || item == nil {
		return ri, err
	}
	bf := bytes.NewBuffer(item.Value)
	err = ri.Deserialize(bf)
	return ri, err
}

func NotifyNameInfoAdd(native *native.NativeService, contract common.Address, functionName string,
	owner common.Address, url string, newer NameInfo) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, owner.ToBase58(), "add", string(url), string(newer.Name)},
		})
}

//only name and desc can be changed
func NotifyNameInfoChange(native *native.NativeService, contract common.Address, functionName string,
	owner common.Address, url string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}

	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, owner.ToBase58(), "update", string(url)},
		})
}

func NotifyNameInfoDel(native *native.NativeService, contract common.Address, functionName string,
	owner common.Address, url string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, owner.ToBase58(), "delete", string(url)},
		})
}

func NotifyNameInfoTransfer(native *native.NativeService, contract common.Address, functionName string,
	orig, newer common.Address, url string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}

	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, "transfer", orig.ToBase58(), newer.ToBase58(), url},
		})
}

func NotifyHeaderAdd(native *native.NativeService, contract common.Address, functionName string,
	owner common.Address, header []byte) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, owner.ToBase58(), "add", string(header)},
		})
}

func NotifyHeaderDel(native *native.NativeService, contract common.Address, functionName string,
	owner common.Address, header []byte) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, owner.ToBase58(), "delete", string(header)},
		})
}

func NotifyHeaderTransfer(native *native.NativeService, contract common.Address, functionName string,
	orig, newer common.Address, header []byte) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}

	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, "transfer", orig.ToBase58(), newer.ToBase58(), string(header)},
		})
}

func createDefaultUrl(native *native.NativeService, name []byte) []byte {
	n := sha256.Sum256(name)
	h := append(n[:], native.Tx.Raw[:]...)
	hash := sha256.Sum256(h)
	return hash[:32]
}

func keyHash(header, url []byte) []byte {
	n := sha256.Sum256(header)
	if url == nil {
		return n[:32]
	}
	h := append(n[:], url[:]...)
	hash := sha256.Sum256(h)
	return hash[:32]
}

func GetGovenAccount(native *native.NativeService, contract common.Address) (common.Address, error) {
	accItem, err := utils.GetStorageItem(native, append(contract[:], ADMIN...))
	if err != nil {
		return common.ADDRESS_EMPTY, errors.NewErr("[DNS GetGovenAccount] get account error!")
	}
	if accItem == nil {
		return common.ADDRESS_EMPTY, errors.NewErr("[DNS GetGovenAccount] no invalid account!")
	}
	return common.AddressParseFromBytes(accItem.Value)
}

func GenDNSRegKey(wallet common.Address) []byte {
	prefix := []byte(DNS_NODE_PREFIX)
	key := append(prefix, wallet[:]...)
	return key
}

func GenKey(contract common.Address) []byte {
	prefix := []byte(PEER_POOL_MAP_PREFIX)
	key := append(prefix, contract[:]...)
	return key
}

func byteSliceEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	b = b[:len(a)]
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func triggerDNSNodeRegEvent(native *native.NativeService, ip, port []byte, walletAddr common.Address, deposit uint64) {
	event := map[string]interface{}{
		"blockHeight": native.Height,
		"eventName":   "DNSNodeRegister",
		"ip":          ip,
		"port":        port,
		"walletAddr":  walletAddr,
		"deposit":     deposit,
	}
	newEvent(native, EVENT_DNS_REG_NODE, event)
}

func triggerDNSNodeUnRegEvent(native *native.NativeService, walletAddr common.Address) {
	event := map[string]interface{}{
		"blockHeight": native.Height,
		"eventName":   "DNSNodeUnReg",
		"walletAddr":  walletAddr,
	}
	newEvent(native, EVENT_DNS_UN_REG_NODE, event)
}

func triggerDNSNodeQuitEvent(native *native.NativeService, walletAddr common.Address) {
	event := map[string]interface{}{
		"blockHeight": native.Height,
		"eventName":   "DNSNodeQuit",
		"walletAddr":  walletAddr,
	}
	newEvent(native, EVENT_DNS_QUIT_NODE, event)
}

func newEvent(srvc *native.NativeService, id uint32, st interface{}) {
	e := event.NotifyEventInfo{}
	e.ContractAddress = srvc.ContextRef.CurrentContext().ContractAddress
	e.EventIdentifier = id
	e.States = st
	srvc.Notifications = append(srvc.Notifications, &e)
	return
}

func appCallTransferOnt(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	err := appCallTransfer(native, utils.UsdtContractAddress, from, to, amount)
	if err != nil {
		return fmt.Errorf("appCallTransferOnt, appCallTransfer error: %v", err)
	}
	return nil
}

func appCallTransfer(native *native.NativeService, contract common.Address, from common.Address, to common.Address, amount uint64) error {
	var sts []usdt.State
	sts = append(sts, usdt.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := usdt.Transfers{
		States: sts,
	}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)

	if _, err := native.NativeCall(contract, "transfer", sink.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransfer, appCall error!")
	}
	return nil
}

func validatePeerPubKeyFormat(pubkey string) error {
	pk, err := vbftconfig.Pubkey(pubkey)
	if err != nil {
		return fmt.Errorf("failed to parse pubkey")
	}
	if !vrf.ValidatePublicKey(pk) {
		return fmt.Errorf("invalid for VRF")
	}
	return nil
}

func getPeerPoolMap(native *native.NativeService, contract common.Address) (*PeerPoolMap, error) {
	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	key := GenKey(contract)

	peerPoolMapBytes, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, fmt.Errorf("[DNS][GetPeerPoolMap] get all peerPoolMap error: %v", err)
	}
	if peerPoolMapBytes == nil {
		return nil, fmt.Errorf("[DNS][GetPeerPoolMap] peerPoolMap is nil")
	}
	if err := peerPoolMap.Deserialize(bytes.NewBuffer(peerPoolMapBytes.Value)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize peerPoolMap error: %v", err)
	}
	return peerPoolMap, nil
}

func putPeerPoolMap(native *native.NativeService, contract common.Address, peerPoolMap *PeerPoolMap) error {
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize peerPoolMap error: %v", err)
	}
	key := GenKey(contract)
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}

func initPeerPoolMap(native *native.NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	pp := &PeerPoolMap{make(map[string]*PeerPoolItem)}
	peerPoolItem := new(PeerPoolItem)
	pp.PeerPoolMap[PEER_POOL_MAP_PREFIX] = peerPoolItem
	key := GenKey(contract)
	bf := new(bytes.Buffer)
	if err := pp.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize peerPoolMap error: %v", err)
	}
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}
