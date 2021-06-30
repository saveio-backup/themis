package dns

import (
	"bytes"

	"encoding/json"
	"fmt"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	cstates "github.com/saveio/themis/core/states"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/global_params"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

var GovenAcc common.Address

const (
	//status
	RegisterCandidateStatus Status = iota
	CandidateStatus
	ConsensusStatus
	QuitConsensusStatus
	QuitingStatus
	BlackStatus
)

const (
	VERSION_CONTRACT_DNS = byte(0)
	INIT_NAME            = "DnsInit"
	REGISTER_NAME        = "regName"
	REGISTER_HEADER      = "regHeader"
	TRANSFER_NAME        = "transferName"
	TRANSFER_HEADER_NAME = "transferHeader"
	UPDATE_DNS_NAME      = "updateDNS"
	GET_DNS_NAME         = "getDNS"
	GET_HEADER_NAME      = "getHeader"
	DEL_DNS              = "delDNS"
	DEL_DNS_HEADER       = "delHeader"
	DNS_NODE_REG         = "dNSNodeReg"
	UN_DNS_NODE_REG      = "unRegDNSNode"
	APPROVE_CANDIDATE    = "approveCandidate"
	REJECT_CANDIDATE     = "rejectCandidate"
	QUIT_NODE            = "quitNode"
	ADD_INIT_POS         = "addInitPos"
	REDUCE_INIT_POS      = "reduceInitPos"
	GET_PEER_POOLMAP     = "getPeerPoolMap"
	GET_PEER_POOLITEM    = "getPeerPoolItem"
	GET_DNSNODE_BYADDR   = "getDNSNodeByAddr"
	GET_ALL_DNSNODES     = "GetAllDnsNodes"
	UPDATE_DNSNODE       = "UpdateDNSNodesInfo"
	GET_PLUGIN_LIST      = "GetPluginList"
)

const (
	//key prefix
	DNS_NODE_PREFIX      = "DNSNodesList"
	PEER_POOL_MAP_PREFIX = "peerPoolMapPrefix"
	ALL_DNS_NODE_PREFIX  = "allDNSNodePrefix"
)

var (
	DSP_HEADER        = []byte("dsp")
	DSP_PLUGIN_HEADER = []byte("dsp-plugin")
)

const (
	MIN_NAME_LEN = 4
)

func InitDNS() {
	native.Contracts[utils.OntDNSAddress] = RegisterDNSContract

}

func RegisterDNSContract(native *native.NativeService) {
	native.Register(INIT_NAME, DnsInit)
	native.Register(REGISTER_NAME, RegistName)
	native.Register(REGISTER_HEADER, RegistHeader)
	native.Register(TRANSFER_NAME, TransferName)
	native.Register(TRANSFER_HEADER_NAME, TransferHeader)
	native.Register(UPDATE_DNS_NAME, UpdateName)
	native.Register(GET_DNS_NAME, GetName)
	native.Register(GET_HEADER_NAME, GetHeader)
	native.Register(DEL_DNS, DelDNS)
	native.Register(DEL_DNS_HEADER, DelHeader)
	//dns govern
	native.Register(DNS_NODE_REG, DNSNodeReg)
	native.Register(UN_DNS_NODE_REG, UnRegDNSNode)
	native.Register(APPROVE_CANDIDATE, ApproveDNSCandidate)
	native.Register(REJECT_CANDIDATE, RejectDNSCandidate)
	native.Register(QUIT_NODE, QuitNode)
	native.Register(ADD_INIT_POS, AddInitPos)
	native.Register(REDUCE_INIT_POS, ReduceInitPos)
	native.Register(GET_PEER_POOLMAP, GetPeerPoolMap)
	native.Register(GET_PEER_POOLITEM, GetPeerPoolItem)
	native.Register(GET_DNSNODE_BYADDR, GetDNSNodeByAddr)
	native.Register(GET_ALL_DNSNODES, GetAllDnsNodes)
	native.Register(UPDATE_DNSNODE, UpdateDNSNodesInfo)
	native.Register(GET_PLUGIN_LIST, GetPluginList)
}

func DnsInit(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	GovenAcc, _ = common.AddressFromBase58("AXxzYEV95ub7Nx32k3JnbCNatZNidvcA1L")
	accBuffer := new(bytes.Buffer)
	err := GovenAcc.Serialize(accBuffer)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DnsInit] governance account serialize error!")
	}
	utils.PutBytes(native, append(contract[:], ADMIN...), accBuffer.Bytes())
	headerkey := GenHeaderKey(contract, DSP_HEADER)
	info := new(bytes.Buffer)
	ri := HeaderInfo{
		Header:      DSP_HEADER,
		HeaderOwner: GovenAcc,
		Desc:        []byte("reserved dsp protocol"),
		BlockHeight: 0,
		TTL:         0,
	}
	err = ri.Serialize(info)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DnsInit] RootInfo serialize error!")
	}

	utils.PutBytes(native, headerkey, info.Bytes())
	if err = initPeerPoolMap(native); err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func RegistName(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	rd := bytes.NewReader(native.Input)
	var req RequestName
	if err := req.Deserialize(rd); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS RegistName] NameInfo deserialize error!")
	}

	if !native.ContextRef.CheckWitness(req.NameOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[DNS RegistName] CheckWitness failed!")
	}
	if len(req.Name) < MIN_NAME_LEN {
		return utils.BYTE_FALSE, errors.NewErr("[DNS RegistName] request name length invalid!")
	}
	var ri NameInfo
	switch req.Type {
	case SYSTEM:
		ri = NameInfo{
			Type:        uint64(NameTypeNormal),
			Header:      DSP_HEADER,
			URL:         createDefaultUrl(native, req.Name),
			Name:        req.Name,
			NameOwner:   req.NameOwner,
			Desc:        req.Desc,
			BlockHeight: uint64(native.Height + 1),
			TTL:         req.DesireTTL,
		}
	case CUSTOM_HEADER:
		if _, err := queryHeader(native, req.Header); err != nil {
			return utils.BYTE_FALSE, err
		}
		//TBD:should check the header group
		//and TTL
		ri = NameInfo{
			Type:        uint64(NameTypeNormal),
			Header:      req.Header,
			URL:         createDefaultUrl(native, req.Name),
			Name:        req.Name,
			NameOwner:   req.NameOwner,
			Desc:        req.Desc,
			BlockHeight: uint64(native.Height + 1),
			TTL:         req.DesireTTL,
		}

	case CUSTOM_URL:
		//TBD: more fee
		//unique check
		unique, err := uniqueCheck(native, DSP_HEADER, req.URL, true)
		if !unique {
			return utils.BYTE_FALSE, err
		}
		ri = NameInfo{
			Type:        uint64(NameTypeNormal),
			Header:      DSP_HEADER,
			URL:         req.URL,
			Name:        req.Name,
			NameOwner:   req.NameOwner,
			Desc:        req.Desc,
			BlockHeight: uint64(native.Height + 1),
			TTL:         req.DesireTTL,
		}

	case CUSTOM_HEADER_URL:
		//unique check
		// unique, err := uniqueCheck(native, req.Header, req.URL, true)
		// if !unique {
		// 	return utils.BYTE_FALSE, err
		// }
		//TBD: more fee
		//check header
		// if _, err := queryHeader(native, req.Header); err != nil {
		// 	return utils.BYTE_FALSE, err
		// }
		//TBD:should check the root group
		//and TTL
		ri = NameInfo{
			Type:        uint64(NameTypeNormal),
			Header:      req.Header,
			URL:         req.URL,
			Name:        req.Name,
			NameOwner:   req.NameOwner,
			Desc:        req.Desc,
			BlockHeight: uint64(native.Height + 1),
			TTL:         req.DesireTTL,
		}
	}

	info := new(bytes.Buffer)

	err := ri.Serialize(info)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS RegistName] NameInfo serialize error!")
	}
	namekey := GenNameInfoKey(contract, req.Header, req.URL)
	log.Debugf("register header %s, url %s for info %s", req.Header, req.URL, info.Bytes())
	utils.PutBytes(native, namekey, info.Bytes())

	if req.Type == CUSTOM_HEADER_URL && string(req.Header) == string(DSP_PLUGIN_HEADER) {
		if err = AddPluginToList(native, namekey); err != nil {
			return utils.BYTE_FALSE, err
		}
	}

	NotifyNameInfoAdd(native, contract, "RegistName", req.NameOwner, string(req.Header)+"://"+string(req.URL), ri)
	return utils.BYTE_TRUE, nil
}

func RegistHeader(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	rd := bytes.NewReader(native.Input)
	var req RequestHeader
	if err := req.Deserialize(rd); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS RegistHeader] RequestHeader deserialize error!")
	}

	if !native.ContextRef.CheckWitness(req.NameOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[DNS RegistHeader] CheckWitness failed!")
	}
	//unique check
	unique, err := uniqueCheck(native, req.Header, nil, false)
	if !unique {
		return utils.BYTE_FALSE, err
	}
	info := new(bytes.Buffer)
	ri := HeaderInfo{
		Header:      req.Header,
		HeaderOwner: req.NameOwner,
		Desc:        req.Desc,
		BlockHeight: uint64(native.Height),
		TTL:         0,
	}
	err = ri.Serialize(info)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS RegistHeader] HeaderInfo serialize error!")
	}
	headerKey := GenHeaderKey(contract, req.Header)
	utils.PutBytes(native, headerKey, info.Bytes())
	NotifyHeaderAdd(native, contract, "RegistHeader", req.NameOwner, req.Header)
	return utils.BYTE_TRUE, nil
}

func TransferName(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	rd := bytes.NewReader(native.Input)
	var tf TranferInfo
	if err := tf.Deserialize(rd); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferName] TranferInfo deserialize error!")
	}
	if !native.ContextRef.CheckWitness(tf.From) {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferName] CheckWitness failed!")
	}
	//check header
	if _, err := queryHeader(native, tf.Header); err != nil {
		return utils.BYTE_FALSE, err
	}
	//check header+url
	nameitem, err := queryURL(native, tf.Header, tf.URL)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	var ni NameInfo
	source := bytes.NewReader(nameitem.Value)
	if err := ni.Deserialize(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferName] dns deserialize error!")
	}
	if ni.NameOwner != tf.From {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferName] owner invalid!")
	}
	info := new(bytes.Buffer)
	ri := NameInfo{
		Type:        uint64(NameTypeNormal),
		Header:      ni.Header,
		URL:         ni.URL,
		Name:        ni.Name,
		NameOwner:   tf.To,
		Desc:        ni.Desc,
		BlockHeight: uint64(native.Height + 1),
		TTL:         uint64(ni.TTL + ni.BlockHeight - uint64(native.Height)),
	}
	err = ri.Serialize(info)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferName] NameInfo serialize error!")
	}
	nameKey := GenNameInfoKey(contract, tf.Header, tf.URL)
	utils.PutBytes(native, nameKey, info.Bytes())
	NotifyNameInfoTransfer(native, contract, "TransferName", tf.From, tf.To, string(tf.Header)+"://"+string(tf.URL))
	return utils.BYTE_FALSE, nil
}

func TransferHeader(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	rd := bytes.NewReader(native.Input)
	var tf TranferInfo
	if err := tf.Deserialize(rd); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferHeader] TranferInfo deserialize error!")
	}
	if !native.ContextRef.CheckWitness(tf.From) {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferHeader] CheckWitness failed!")
	}
	//check header
	headerItem, err := queryHeader(native, tf.Header)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	var hi HeaderInfo
	source := bytes.NewReader(headerItem.Value)
	if err := hi.Deserialize(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferHeader] dns deserialize error!")
	}
	if hi.HeaderOwner != tf.From {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferHeader] owner invalid!")
	}
	info := new(bytes.Buffer)

	var ttl uint64
	if (hi.TTL + hi.BlockHeight) <= uint64(native.Height) {
		ttl = 0
	} else {
		ttl = (hi.TTL + hi.BlockHeight) - uint64(native.Height)
	}

	ri := HeaderInfo{
		Header:      tf.Header,
		HeaderOwner: tf.To,
		Desc:        hi.Desc,
		BlockHeight: uint64(native.Height + 1),
		TTL:         ttl,
	}
	err = ri.Serialize(info)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS TransferHeader] HeaderInfo serialize error!")
	}
	headerKey := GenHeaderKey(contract, tf.Header)
	utils.PutBytes(native, headerKey, info.Bytes())
	NotifyHeaderTransfer(native, contract, "TransferHeader", tf.From, tf.To, tf.Header)
	return utils.BYTE_FALSE, nil
}

func UpdateName(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	rd := bytes.NewReader(native.Input)
	var req RequestName
	if err := req.Deserialize(rd); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS UpdateName] ReqInfo deserialize error!")
	}
	if !native.ContextRef.CheckWitness(req.NameOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[DNS UpdateName] CheckWitness failed!")
	}
	//check header
	if _, err := queryHeader(native, req.Header); err != nil {
		return utils.BYTE_FALSE, err
	}
	//check header+url
	nameitem, err := queryURL(native, req.Header, req.URL)
	if err != nil {
		log.Errorf("update dns header %s url %s err %s", req.Header, req.URL, err)
		return utils.BYTE_FALSE, err
	}
	var ni NameInfo
	source := bytes.NewReader(nameitem.Value)
	if err := ni.Deserialize(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS UpdateName] dns deserialize error!")
	}
	info := new(bytes.Buffer)
	if ni.NameOwner == req.NameOwner {
		//TBD: fee of change
		ni.Type = req.Type
		ni.Name = req.Name
		ni.Desc = req.Desc
		ni.TTL = req.DesireTTL
		ni.BlockHeight = uint64(native.Height + 1)
		err = ni.Serialize(info)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[DNS UpdateName] RootInfo serialize error!")
		}
	} else {
		return utils.BYTE_FALSE, errors.NewErr("[DNS UpdateName] permission deny!")
	}
	namekey := GenNameInfoKey(contract, req.Header, req.URL)
	utils.PutBytes(native, namekey, info.Bytes())

	if ni.Type == uint64(NameTypePlugin) {
		if err = AddPluginToList(native, namekey); err != nil {
			return utils.BYTE_FALSE, err
		}
	}
	NotifyNameInfoChange(native, contract, "UpdateName", req.NameOwner, string(req.Header)+"://"+string(req.URL))
	return utils.BYTE_TRUE, nil
}

func GetName(native *native.NativeService) ([]byte, error) {
	rd := bytes.NewReader(native.Input)
	//check header
	var req ReqInfo
	if err := req.Deserialize(rd); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS GetName] ReqInfo deserialize error!")
	}
	if _, err := queryHeader(native, req.Header); err != nil {
		return utils.BYTE_FALSE, err
	}
	nameitem, err := queryURL(native, req.Header, req.URL)
	if err != nil {
		log.Errorf("get dns name header %s url %s err %s", req.Header, req.URL, err)
		return utils.BYTE_FALSE, err
	}
	return nameitem.Value, nil
}

func GetHeader(native *native.NativeService) ([]byte, error) {
	rd := bytes.NewReader(native.Input)
	var req ReqInfo
	if err := req.Deserialize(rd); err != nil {
		return utils.BYTE_FALSE, err
	}

	headerItem, err := queryHeader(native, req.Header)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	return headerItem.Value, nil
}

func DelDNS(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	var req ReqInfo
	rd := bytes.NewReader(native.Input)
	if err := req.Deserialize(rd); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelDNS] ReqInfo deserialize error!")
	}
	if !native.ContextRef.CheckWitness(req.Owner) {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelDNS] CheckWitness failed!")
	}
	nameitem, err := queryURL(native, req.Header, req.URL)
	if err != nil {
		log.Errorf("del dns header %s url %s err %s", req.Header, req.URL, err)
		return utils.BYTE_FALSE, err
	}
	var ni NameInfo
	source := bytes.NewReader(nameitem.Value)
	GovenAcc, err = GetGovenAccount(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelDNS]get governance account failed")
	}
	if err := ni.Deserialize(source); err != nil {
		//invalid NameInfo,delete it by using goven account

		if req.Owner == GovenAcc {

			nameKey := GenNameInfoKey(contract, req.Header, req.URL)
			utils.DelStorageItem(native, nameKey)
			NotifyNameInfoDel(native, contract, "DelDNS", req.Owner, string(req.Header)+"://"+string(req.URL))
			return utils.BYTE_TRUE, nil
		}
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelDNS] dns deserialize error!")
	}
	if ni.NameOwner == req.Owner || req.Owner == GovenAcc {
		nameKey := GenNameInfoKey(contract, req.Header, req.URL)
		utils.DelStorageItem(native, nameKey)
		NotifyNameInfoDel(native, contract, "DelDNS", req.Owner, string(req.Header)+"://"+string(req.URL))
		return utils.BYTE_TRUE, nil
	} else {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelDNS] permission deny!")
	}

}

func DelHeader(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	var req ReqInfo
	rd := bytes.NewReader(native.Input)
	if err := req.Deserialize(rd); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelHeader] ReqInfo deserialize error!")
	}
	if !native.ContextRef.CheckWitness(req.Owner) {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelHeader] CheckWitness failed!")
	}
	headerItem, err := queryHeader(native, req.Header)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	var hi HeaderInfo
	source := bytes.NewReader(headerItem.Value)
	GovenAcc, err = GetGovenAccount(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelHeader]get governance account failed")
	}
	if err := hi.Deserialize(source); err != nil {

		//invalid headerinfo,delete it by using goven account
		if req.Owner == GovenAcc {
			headerKey := GenHeaderKey(contract, req.Header)
			utils.DelStorageItem(native, headerKey)
			NotifyHeaderDel(native, contract, "DelHeader", req.Owner, req.Header)
			return utils.BYTE_TRUE, nil
		}
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelHeader] header deserialize error!")
	}
	if hi.HeaderOwner == req.Owner || req.Owner == GovenAcc {
		headerKey := GenHeaderKey(contract, req.Header)
		utils.DelStorageItem(native, headerKey)
		NotifyHeaderDel(native, contract, "DelHeader", req.Owner, req.Header)
		return utils.BYTE_TRUE, nil
	} else {
		return utils.BYTE_FALSE, errors.NewErr("[DNS DelHeader] permission deny!")
	}
}

func uniqueCheck(native *native.NativeService, header, url []byte, bFull bool) (bool, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	if bFull {
		nameKey := GenNameInfoKey(contract, header, url)
		nameItem, err := utils.GetStorageItem(native, nameKey)

		if err != nil {
			return false, errors.NewErr("[DNS uniqueCheck] get name error!")
		}
		if nameItem != nil {
			return false, errors.NewErr("[DNS uniqueCheck] url already regist")
		}
	} else {
		headerKey := GenHeaderKey(contract, header)
		headerItem, err := utils.GetStorageItem(native, headerKey)

		if err != nil {
			return false, errors.NewErr("[DNS uniqueCheck] get header error!")
		}
		if headerItem != nil {
			return false, errors.NewErr("[DNS uniqueCheck] header already regist")
		}
	}

	return true, nil
}

func queryHeader(native *native.NativeService, header []byte) (*cstates.StorageItem, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	headerKey := GenHeaderKey(contract, header)
	headerItem, err := utils.GetStorageItem(native, headerKey)

	if err != nil {
		return nil, errors.NewErr("[DNS queryHeader] query header error!")
	}

	if headerItem == nil {
		return nil, errors.NewErr("[DNS queryHeader] no invalid header!")
	}
	return headerItem, nil
}

func queryURL(native *native.NativeService, header, url []byte) (*cstates.StorageItem, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	namekey := GenNameInfoKey(contract, header, url)
	nameItem, err := utils.GetStorageItem(native, namekey)
	if err != nil {
		log.Errorf("query head %s, url %s err %s", header, url, err)
		return nil, errors.NewErr("[DNS queryURL] get url error!")
	}
	if nameItem == nil {
		log.Errorf("query head %s, url %s not found item", header, url)
		return nil, errors.NewErr("[DNS queryURL] name item is nil!")
	}
	return nameItem, nil
}

//Register a candidate node, used by contracts.
func DNSNodeReg(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	nodeInfo := new(DNSNodeInfo)
	if err := nodeInfo.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err,
			errors.ErrNoCode, "[DNS RegDnsNode] DNSNodeInfo deserialization error!")
	}

	if nodeInfo.InitDeposit <= 0 {
		return utils.BYTE_FALSE, fmt.Errorf("[DNSNodeReg] RegNode:%s deposit:%d lt zero!", nodeInfo.WalletAddr, nodeInfo.InitDeposit)
	}
	//check owner
	err := utils.ValidateOwner(native, nodeInfo.WalletAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[DNSNodeReg] validateOwner error: %v", err)
	}

	//check peerPubkey
	if err := validatePeerPubKeyFormat(nodeInfo.PeerPubKey); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[DNSNodeReg] invalid peer pubkey")
	}
	//get peerPoolMap

	peerPoolMap, err := getPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[DNSNodeReg] get peerPoolMap error: %v", err)
	}

	//check if exist in PeerPool
	_, ok := peerPoolMap.PeerPoolMap[nodeInfo.PeerPubKey]
	if ok {
		return utils.BYTE_FALSE, fmt.Errorf("[DNSNodeReg] registerCandidate, peerPubkey is already in peerPoolMap")
	}

	peerPoolItem := &PeerPoolItem{
		PeerPubkey:    nodeInfo.PeerPubKey,
		WalletAddress: nodeInfo.WalletAddr,
		Status:        RegisterCandidateStatus,
		TotalInitPos:  nodeInfo.InitDeposit,
	}
	peerPoolMap.PeerPoolMap[nodeInfo.PeerPubKey] = peerPoolItem
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[DNSNodeReg] putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	nodeKey := GenDNSRegKey(nodeInfo.WalletAddr)
	buff := new(bytes.Buffer)
	if err := nodeInfo.Serialize(buff); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[DNSNodeReg] nodeInfo serialize error: %v", err)
	}
	utils.PutBytes(native, nodeKey, buff.Bytes())

	//TODO:get globalParam

	err = appCallTransferOnt(native, nodeInfo.WalletAddr, utils.OntDNSAddress, nodeInfo.InitDeposit)
	if err != nil {
		return utils.BYTE_FALSE,
			errors.NewErr("[DNSNodeReg] appCallTransferOnt error")
	}

	//TODO:update total stake

	triggerDNSNodeRegEvent(native, nodeInfo.IP, nodeInfo.Port, nodeInfo.WalletAddr, nodeInfo.InitDeposit)
	return utils.BYTE_TRUE, nil
}

//Unregister a registered-candidate node, will remove node from pool, and unfreeze deposit ont.
func UnRegDNSNode(native *native.NativeService) ([]byte, error) {
	params := new(UnRegisterCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	address := params.Address
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := getPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	//check if exist in PeerPool
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("[UnRegDNSNode[ peerPubkey is not in peerPoolMap: %v", err)
	}

	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("[UnRegDNSNode[ peer status is not RegisterCandidateStatus")
	}

	//check owner address
	if peerPoolItem.WalletAddress != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("[UnRegDNSNode] address is not peer owner")
	}
	err = appCallTransferOnt(native, utils.OntDNSAddress, address, peerPoolItem.TotalInitPos)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[UnRegDNSNode] appCallTransferOnt, ont transfer error: %v", err)
	}
	delete(peerPoolMap.PeerPoolMap, params.PeerPubkey)
	utils.DelStorageItem(native, GenDNSRegKey(params.Address))
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[UnRegDNSNode] putPeerPoolMap, put peerPoolMap error: %v", err)
	}
	triggerDNSNodeUnRegEvent(native, params.Address)
	return utils.BYTE_TRUE, nil
}

func ApproveDNSCandidate(native *native.NativeService) ([]byte, error) {
	params := new(ApproveCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[ApproveDNSCandidate] contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[ApproveDNSCandidate] get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[ApproveDNSCandidate] ValidateOwner error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//TODO:get globalParam

	//check if peerPoolMap full
	//get peerPoolMap
	peerPoolMap, err := getPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[ApproveDNSCandidate] get peerPoolMap error: %v", err)
	}

	//get peerPool
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("[ApproveDNSCandidate] peerPubkey is not in peerPoolMap")
	}

	//check Pos
	if peerPoolItem.TotalInitPos < 0 {
		return utils.BYTE_FALSE, fmt.Errorf("[ApproveDNSCandidate] initPos must >= %d", 0)
	}

	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("[ApproveDNSCandidate] peer status is not RegisterCandidateStatus")
	}

	peerPoolItem.Status = ConsensusStatus
	//peerPoolItem.TotalInitPos = 0

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[ApproveDNSCandidate] put peerPoolMap error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func RejectDNSCandidate(native *native.NativeService) ([]byte, error) {
	params := new(PubKeyParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[RejectDNSCandidatede] serialize, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[RejectDNSCandidatede]  getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[RejectDNSCandidatede] checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get peerPoolMap
	peerPoolMap, err := getPeerPoolMap(native, contract)
	if err != nil || peerPoolMap == nil {
		return utils.BYTE_FALSE, fmt.Errorf("[RejectDNSCandidatede] get peerPoolMap error: %v", err)
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("[RejectDNSCandidatede]  peerPubkey is not in peerPoolMap")
	}
	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("[RejectDNSCandidatede]  peerPubkey is not RegisterCandidateStatus")
	}

	err = appCallTransferOnt(native, utils.OntDNSAddress, peerPoolItem.WalletAddress, peerPoolItem.TotalInitPos)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
	}

	utils.DelStorageItem(native, GenDNSRegKey(peerPoolItem.WalletAddress))
	//remove peerPubkey from peerPool
	delete(peerPoolMap.PeerPoolMap, params.PeerPubkey)
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Quit in status registered or consensus, used by node owner.
func QuitNode(native *native.NativeService) ([]byte, error) {
	params := new(QuitNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get peerPoolMap
	peerPoolMap, err := getPeerPoolMap(native, contract)
	if err != nil || peerPoolMap == nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, peerPubkey is not in peerPoolMap")
	}

	if address != peerPoolItem.WalletAddress {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, peerPubkey is not registered by this address")
	}
	if peerPoolItem.Status != ConsensusStatus && peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, peerPubkey is not  ConsensusStatus")
	}

	//change peerPool status
	if peerPoolItem.Status == ConsensusStatus {
		peerPoolItem.Status = QuitConsensusStatus
	} else {
		peerPoolItem.Status = QuitingStatus
	}

	err = appCallTransferOnt(native, utils.OntDNSAddress, address, peerPoolItem.TotalInitPos)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
	}
	utils.DelStorageItem(native, GenDNSRegKey(params.Address))
	delete(peerPoolMap.PeerPoolMap, params.PeerPubkey)
	//peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//add  pos of a node
func AddInitPos(native *native.NativeService) ([]byte, error) {
	params := new(ChangeInitPosParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize changeInitPosParam error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddInitPos, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get peerPoolMap
	peerPoolMap, err := getPeerPoolMap(native, contract)
	if err != nil || peerPoolMap == nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddInitPos, get peerPoolMap error: %v", err)
	}
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("addInitPos, peerPubkey is not in peerPoolMap")
	}
	if peerPoolItem.WalletAddress != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("address is not peer owner")
	}
	if peerPoolItem.Status != ConsensusStatus && peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("addInitPos, peerPubkey is not candidate")
	}
	if params.Pos <= 0 {
		return utils.BYTE_FALSE, fmt.Errorf("addInitPos add pos lq 0!")
	}

	peerPoolItem.TotalInitPos += params.Pos
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap error: %v", err)
	}

	//ont transfer
	err = appCallTransferOnt(native, params.Address, utils.OntDNSAddress, params.Pos)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
	}

	//TODO:update total stake
	return utils.BYTE_TRUE, nil
}

//reduce init pos of a node
func ReduceInitPos(native *native.NativeService) ([]byte, error) {
	params := new(ChangeInitPosParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize changeInitPosParam error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get peerPoolMap
	peerPoolMap, err := getPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("reduceInitPos, peerPubkey is not in peerPoolMap")
	}
	if peerPoolItem.WalletAddress != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("address is not peer owner")
	}
	if peerPoolItem.TotalInitPos < params.Pos {
		return utils.BYTE_FALSE, fmt.Errorf("reduceInitPos, initPos can not be negative")
	}
	newInitPos := peerPoolItem.TotalInitPos - params.Pos
	peerPoolMap.PeerPoolMap[params.PeerPubkey].TotalInitPos = newInitPos
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func GetPeerPoolMap(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	key := GenKey(contract)

	peerPoolMapBytes, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, fmt.Errorf("[GetPeerPoolMap] get all peerPoolMap error: %v", err)
	}
	if peerPoolMapBytes == nil {
		return nil, fmt.Errorf("[GetPeerPoolMap] peerPoolMap is nil")
	}

	return peerPoolMapBytes.Value, nil
}

//get peer pool item by pubKey
func GetPeerPoolItem(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	params := new(PubKeyParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetPeerPoolItem] serialize, contract params deserialize error: %v", err)
	}
	//get peerPoolMap
	peerPoolMap, err := getPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetPeerPoolItem] get peerPoolMap error: %v", err)
	}
	if peerPoolMap == nil {
		return nil, errors.NewErr("[GetPeerPoolItem] peerPoolMap is nil")
	}
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("[GetPeerPoolItem] peerPubkey is not in peerPoolMap")
	}
	buf := new(bytes.Buffer)
	if err = peerPoolItem.Serialize(buf); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetPeerPoolItem] peerPoolItem Serialize error:%v", err)
	}
	return buf.Bytes(), nil
}

func GetDNSNodeByAddr(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	wallet, err := utils.DecodeAddress(source)
	if err != nil {
		return nil, errors.NewErr("[DNS QueryDNSNodeByAddr] decodeBytes of walletaddr error")
	}
	key := GenDNSRegKey(wallet)
	nodeInfo, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, errors.NewErr("[DNS QueryDNSNodeByAddr] get nodeInfo from db error")
	}
	if nodeInfo == nil {
		return nil, errors.NewErr("[GetDNSNodeByAddr] nodeInfo is nil")
	}
	return nodeInfo.Value, nil
}

func GetAllDnsNodes(native *native.NativeService) ([]byte, error) {
	var node DNSNodeInfo
	iter := native.CacheDB.NewIterator([]byte(DNS_NODE_PREFIX))
	defer iter.Release()
	if err := iter.Error(); err != nil {
		return nil, err
	}
	nodesMap := make(map[string]DNSNodeInfo)
	for has := iter.First(); has; has = iter.Next() {
		valueItem, err := cstates.GetValueFromRawStorageItem(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("[DNS GetRangeMap] GetValueItem error!:%v", err)
		}
		if err := node.Deserialize(bytes.NewBuffer(valueItem)); err != nil {
			return nil, errors.NewErr("[DNS GetRangeMap] DNSNode deserialization error.")
		}
		nodesMap[node.WalletAddr.ToHexString()] = node
	}

	nm, err := json.Marshal(nodesMap)
	if err != nil {
		return nil, fmt.Errorf("[DNS GetRangeMap] DNSNode Marshal error:%s\n", err)
	}
	return nm, nil
}

//update the ip and port of the wallet .
func UpdateDNSNodesInfo(native *native.NativeService) ([]byte, error) {
	nodeInfo := new(UpdateNodeParam)
	if err := nodeInfo.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err,
			errors.ErrNoCode, "[DNS UpdateDnsNode] DNSNodeInfo deserialization error!")
	}

	//check IP and Port
	if nodeInfo.IP == nil || nodeInfo.Port == nil {
		return utils.BYTE_FALSE, fmt.Errorf("[DNS UpdateDnsNode] IP or port is nil")
	}

	//check owner
	err := utils.ValidateOwner(native, nodeInfo.WalletAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[DNS UpdateDnsNode] validateOwner error: %v", err)
	}

	key := GenDNSRegKey(nodeInfo.WalletAddr)
	oldNode, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, errors.NewErr("[DNS UpdateDnsNode] get nodeInfo from db error")
	}
	if oldNode == nil {
		return nil, errors.NewErr("[DNS UpdateDnsNode] the wallet had not registered")
	}

	var dn DNSNodeInfo
	reader := bytes.NewReader(oldNode.Value)
	err = dn.Deserialize(reader)
	if err != nil {
		return nil, errors.NewErr("[DNS UpdateDnsNode] deserialization oldNode error")
	}
	dn.IP = nodeInfo.IP
	dn.Port = nodeInfo.Port

	buff := new(bytes.Buffer)
	if err := dn.Serialize(buff); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[DNSNodeReg] nodeInfo serialize error: %v", err)
	}
	utils.PutBytes(native, key, buff.Bytes())

	triggerDNSNodeRegEvent(native, dn.IP, dn.Port, dn.WalletAddr, dn.InitDeposit)
	return utils.BYTE_TRUE, nil
}

func GetPluginList(native *native.NativeService) ([]byte, error) {
	pluginList, err := GetDnsPliginList(native)
	if err != nil {
		return EncRet(false, []byte("[DNS GetPluginList] GetPluginList GetDnsPliginList error!")), nil
	}
	bf := new(bytes.Buffer)
	err = pluginList.Serialize(bf)
	if err != nil {
		return EncRet(false, []byte("[DNS GetPluginList] GetPluginList Serialize error!")), nil
	}

	var nameInfos NameInfoList
	for _, plugin := range pluginList.List {
		nameItem, err := utils.GetStorageItem(native, plugin.NameKey)
		if err != nil {
			// return nil, errors.NewErr("[DNS DelDNS] get url error!")
			continue
		}
		if nameItem == nil {
			continue
		}
		var ni NameInfo
		source := bytes.NewReader(nameItem.Value)
		if err := ni.Deserialize(source); err != nil {
			// return utils.BYTE_FALSE, errors.NewErr("[DNS DelDNS] NameInfo deserialize error!")
			continue
		}
		nameInfos.NameNum++
		nameInfos.List = append(nameInfos.List, ni)
	}

	pluginListBf := new(bytes.Buffer)
	err = nameInfos.Serialize(pluginListBf)
	if err != nil {
		return EncRet(false, []byte("[DNS GetPluginList] GetPluginList NameInfoList Serialize error!")), nil
	}
	return EncRet(true, pluginListBf.Bytes()), nil
}
