package dns

import (
	"io"

	"fmt"

	"github.com/saveio/themis/common"

	//"github.com/saveio/themis/common/serialization"
	//"github.com/saveio/themis/errors"
	"math"
	"sort"

	"github.com/saveio/themis/common/serialization"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

const (
	SYSTEM            uint64 = 0x00
	CUSTOM_HEADER     uint64 = 0x01
	CUSTOM_URL        uint64 = 0x02
	CUSTOM_HEADER_URL uint64 = 0x04
	UPDATE            uint64 = 0x08
)

type NameType uint64

const (
	NameTypeNormal NameType = iota
	NameTypePlugin
)

type RequestName struct {
	Type      uint64
	Header    []byte
	URL       []byte
	Name      []byte
	NameOwner common.Address
	Desc      []byte
	DesireTTL uint64
}

type NameInfo struct {
	Type        uint64
	Header      []byte
	URL         []byte
	Name        []byte
	NameOwner   common.Address
	Desc        []byte
	BlockHeight uint64
	TTL         uint64 // 0: bypass
}

type RequestHeader struct {
	Header    []byte
	NameOwner common.Address
	Desc      []byte
	DesireTTL uint64
}

type HeaderInfo struct {
	Header      []byte
	HeaderOwner common.Address
	Desc        []byte
	BlockHeight uint64
	TTL         uint64 // 0: bypass
}

type ReqInfo struct {
	Header []byte
	URL    []byte
	Owner  common.Address
}

type TranferInfo struct {
	Header []byte
	URL    []byte
	From   common.Address
	To     common.Address
}

func (this *RequestName) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.Type)); err != nil {
		return fmt.Errorf("[RequestName] [Type:%v] serialize from error:%v", this.Type, err)
	}
	if err := utils.WriteBytes(w, this.Header); err != nil {
		return fmt.Errorf("[RequestName] [Header:%v] serialize from error:%v", this.Header, err)
	}
	if err := utils.WriteBytes(w, this.URL); err != nil {
		return fmt.Errorf("[RequestName] [URL:%v] serialize from error:%v", this.URL, err)
	}
	if err := utils.WriteBytes(w, this.Name); err != nil {
		return fmt.Errorf("[RequestName] [Name:%v] serialize from error:%v", this.Name, err)
	}
	if err := utils.WriteAddress(w, this.NameOwner); err != nil {
		return fmt.Errorf("[RequestName] [NameOwner:%v] serialize from error:%v", this.NameOwner, err)
	}
	if err := utils.WriteBytes(w, this.Desc); err != nil {
		return fmt.Errorf("[RequestName] [Desc:%v] serialize from error:%v", this.Desc, err)
	}
	if err := utils.WriteVarUint(w, uint64(this.DesireTTL)); err != nil {
		return fmt.Errorf("[RequestName] [DesireTTL:%v] serialize from error:%v", this.DesireTTL, err)
	}
	return nil
}

func (this *RequestName) Deserialize(r io.Reader) error {
	var err error
	if this.Type, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[RequestName] [Type] deserialize from error:%v", err)
	}
	if this.Header, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[RequestName] [Header] deserialize from error:%v", err)
	}
	if this.URL, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[RequestName] [URL] deserialize from error:%v", err)
	}
	if this.Name, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[RequestName] [Name] deserialize from error:%v", err)
	}
	if this.NameOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[RequestName] [NameOwner] deserialize from error:%v", err)
	}
	if this.Desc, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[RequestName] [Desc] deserialize from error:%v", err)
	}
	if this.DesireTTL, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[RequestName] [DesireTTL] deserialize from error:%v", err)
	}
	return nil
}

func (this *RequestHeader) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.Header); err != nil {
		return fmt.Errorf("[RequestHeader] [Header:%v] serialize from error:%v", this.Header, err)
	}
	if err := utils.WriteAddress(w, this.NameOwner); err != nil {
		return fmt.Errorf("[RequestHeader] [NameOwner:%v] serialize from error:%v", this.NameOwner, err)
	}
	if err := utils.WriteBytes(w, this.Desc); err != nil {
		return fmt.Errorf("[RequestHeader] [Desc:%v] serialize from error:%v", this.Desc, err)
	}
	if err := utils.WriteVarUint(w, uint64(this.DesireTTL)); err != nil {
		return fmt.Errorf("[RequestHeader] [DesireTTL:%v] serialize from error:%v", this.DesireTTL, err)
	}
	return nil
}

func (this *RequestHeader) Deserialize(r io.Reader) error {
	var err error
	if this.Header, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[RequestHeader] [Header] deserialize from error:%v", err)
	}
	if this.NameOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[RequestHeader] [NameOwner] deserialize from error:%v", err)
	}
	if this.Desc, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[RequestHeader] [Desc] deserialize from error:%v", err)
	}
	if this.DesireTTL, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[RequestHeader] [DesireTTL] deserialize from error:%v", err)
	}
	return nil
}

func (this *NameInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.Type)); err != nil {
		return fmt.Errorf("[NameInfo] [Type:%v] serialize from error:%v", this.Type, err)
	}
	if err := utils.WriteBytes(w, this.Header); err != nil {
		return fmt.Errorf("[NameInfo] [Header:%v] serialize from error:%v", this.Header, err)
	}
	if err := utils.WriteBytes(w, this.URL); err != nil {
		return fmt.Errorf("[NameInfo] [URL:%v] serialize from error:%v", this.URL, err)
	}
	if err := utils.WriteBytes(w, this.Name); err != nil {
		return fmt.Errorf("[NameInfo] [Name:%v] serialize from error:%v", this.Name, err)
	}
	if err := utils.WriteAddress(w, this.NameOwner); err != nil {
		return fmt.Errorf("[NameInfo] [NameOwner:%v] serialize from error:%v", this.NameOwner, err)
	}
	if err := utils.WriteBytes(w, this.Desc); err != nil {
		return fmt.Errorf("[NameInfo] [Desc:%v] serialize from error:%v", this.Desc, err)
	}
	if err := utils.WriteVarUint(w, this.BlockHeight); err != nil {
		return fmt.Errorf("[NameInfo] [BlockHeight:%v] serialize from error:%v", this.BlockHeight, err)
	}
	if err := utils.WriteVarUint(w, uint64(this.TTL)); err != nil {
		return fmt.Errorf("[NameInfo] [TTL:%v] serialize from error:%v", this.TTL, err)
	}

	return nil
}

func (this *NameInfo) Deserialize(r io.Reader) error {
	var err error
	if this.Type, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[NameInfo] [Type] deserialize from error:%v", err)
	}
	if this.Header, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[NameInfo] [Header] deserialize from error:%v", err)
	}
	if this.URL, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[NameInfo] [URL] deserialize from error:%v", err)
	}
	if this.Name, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[NameInfo] [Name] deserialize from error:%v", err)
	}
	if this.NameOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[NameInfo] [NameOwner] deserialize from error:%v", err)
	}
	if this.Desc, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[NameInfo] [Desc] deserialize from error:%v", err)
	}
	if this.BlockHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[NameInfo] [BlockHeight] deserialize from error:%v", err)
	}
	if this.TTL, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[NameInfo] [TTL] deserialize from error:%v", err)
	}
	return nil
}

type NameInfoList struct {
	NameNum uint64
	List    []NameInfo
}

func (this *NameInfoList) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.NameNum); err != nil {
		return fmt.Errorf("[NameInfoList] [NameNum:%v] serialize from error:%v", this.NameNum, err)
	}

	for index := 0; uint64(index) < this.NameNum; index++ {
		if err := this.List[index].Serialize(w); err != nil {
			return fmt.Errorf("[NameInfoList] [List:%v] serialize from error:%v", this.List[index].URL, err)
		}
	}
	return nil
}

func (this *NameInfoList) Deserialize(r io.Reader) error {
	var err error
	if this.NameNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[NameInfoList] [NameNum] deserialize from error:%v", err)
	}
	var tmpInfo NameInfo
	for index := 0; uint64(index) < this.NameNum; index++ {
		if err := tmpInfo.Deserialize(r); err != nil {
			return fmt.Errorf("[NameInfoList] [List] deserialize from error:%v", err)
		}
		this.List = append(this.List, tmpInfo)
	}
	return nil
}

func (this *HeaderInfo) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.Header); err != nil {
		return fmt.Errorf("[HeaderInfo] [Header:%v] serialize from error:%v", this.Header, err)
	}
	if err := utils.WriteAddress(w, this.HeaderOwner); err != nil {
		return fmt.Errorf("[HeaderInfo] [HeaderOwner:%v] serialize from error:%v", this.HeaderOwner, err)
	}
	if err := utils.WriteBytes(w, this.Desc); err != nil {
		return fmt.Errorf("[HeaderInfo] [Desc:%v] serialize from error:%v", this.Desc, err)
	}
	if err := utils.WriteVarUint(w, this.BlockHeight); err != nil {
		return fmt.Errorf("[HeaderInfo] [BlockHeight:%v] serialize from error:%v", this.BlockHeight, err)
	}
	if err := utils.WriteVarUint(w, uint64(this.TTL)); err != nil {
		return fmt.Errorf("[HeaderInfo] [TTL:%v] serialize from error:%v", this.TTL, err)
	}
	return nil
}

func (this *HeaderInfo) Deserialize(r io.Reader) error {
	var err error
	if this.Header, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[HeaderInfo] [Header] deserialize from error:%v", err)
	}
	if this.HeaderOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[HeaderInfo] [HeaderOwner] deserialize from error:%v", err)
	}
	if this.Desc, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[HeaderInfo] [Desc] deserialize from error:%v", err)
	}
	if this.BlockHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[HeaderInfo] [BlockHeight] deserialize from error:%v", err)
	}
	if this.TTL, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[HeaderInfo] [TTL] deserialize from error:%v", err)
	}
	return nil
}

func (this *ReqInfo) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.Header); err != nil {
		return fmt.Errorf("[ReqInfo] [Header:%v] serialize from error:%v", this.Header, err)
	}
	if err := utils.WriteBytes(w, this.URL); err != nil {
		return fmt.Errorf("[ReqInfo] [URL:%v] serialize from error:%v", this.URL, err)
	}
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("[ReqInfo] [Owner:%v] serialize from error:%v", this.Owner, err)
	}
	return nil
}

func (this *ReqInfo) Deserialize(r io.Reader) error {
	var err error
	if this.Header, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[HeaderInfo] [Header] deserialize from error:%v", err)
	}
	if this.URL, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[HeaderInfo] [URL] deserialize from error:%v", err)
	}
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[HeaderInfo] [Owner] deserialize from error:%v", err)
	}
	return nil
}

func (this *TranferInfo) Serialize(w io.Writer) error {
	if err := utils.WriteBytes(w, this.Header); err != nil {
		return fmt.Errorf("[TranferInfo] [Header:%v] serialize from error:%v", this.Header, err)
	}
	if err := utils.WriteBytes(w, this.URL); err != nil {
		return fmt.Errorf("[TranferInfo] [URL:%v] serialize from error:%v", this.URL, err)
	}
	if err := utils.WriteAddress(w, this.From); err != nil {
		return fmt.Errorf("[TranferInfo] [From:%v] serialize from error:%v", this.From, err)
	}
	if err := utils.WriteAddress(w, this.To); err != nil {
		return fmt.Errorf("[TranferInfo] [To:%v] serialize from error:%v", this.To, err)
	}
	return nil
}

func (this *TranferInfo) Deserialize(r io.Reader) error {
	var err error
	if this.Header, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[TranferInfo] [Header] deserialize from error:%v", err)
	}
	if this.URL, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[TranferInfo] [URL] deserialize from error:%v", err)
	}
	if this.From, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[TranferInfo] [From] deserialize from error:%v", err)
	}
	if this.To, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[TranferInfo] [To] deserialize from error:%v", err)
	}
	return nil
}

type Status uint8

func (this *Status) Serialize(w io.Writer) error {
	if err := serialization.WriteUint8(w, uint8(*this)); err != nil {
		return fmt.Errorf("serialization.WriteUint8, serialize status error: %v", err)
	}
	return nil
}

func (this *Status) Deserialize(r io.Reader) error {
	status, err := serialization.ReadUint8(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint8, deserialize status error: %v", err)
	}
	*this = Status(status)
	return nil
}

type DNSNodeInfo struct {
	WalletAddr  common.Address
	IP          []byte
	Port        []byte
	InitDeposit uint64
	PeerPubKey  string
}

func (this *DNSNodeInfo) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.WalletAddr); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [WalletAddr:%v] serialize from error:%v", this.WalletAddr, err)
	}

	if err := utils.WriteBytes(w, this.IP); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [IP:%v] serialize from error:%v", this.IP, err)
	}
	if err := utils.WriteBytes(w, this.Port); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [Port:%v] serialize from error:%v", this.Port, err)
	}
	if err := utils.WriteVarUint(w, this.InitDeposit); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [TotalDeposit:%v] serialize from error:%v", this.InitDeposit, err)
	}
	if err := serialization.WriteString(w, this.PeerPubKey); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [PeerPubKey:%v] serialize from error:%v", this.InitDeposit, err)
	}
	return nil
}

func (this *DNSNodeInfo) Deserialize(r io.Reader) error {
	var err error
	if this.WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [WalletAddr] deserialize from error:%v", err)
	}
	if this.IP, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [IP] deserialize from error:%v", err)
	}
	if this.Port, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [Port] deserialize from error:%v", err)
	}
	if this.InitDeposit, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [TotaoDeposit] deserialize from error:%v", err)
	}
	if this.InitDeposit > math.MaxUint64 {
		return fmt.Errorf("initPos larger than max of uint64")
	}
	if this.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [PeerPubKey] deserialize from error:%v", err)
	}

	return nil
}

type UnregisterDnsNode struct {
	PeerPubkey string
	WalletAddr common.Address
}

func (this *UnregisterDnsNode) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, request peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.WalletAddr[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	return nil
}

func (this *UnregisterDnsNode) Deserialize(r io.Reader) error {
	var err error
	if this.PeerPubkey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	if this.WalletAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	return nil
}

type PeerPoolMap struct {
	PeerPoolMap map[string]*PeerPoolItem
}

func (this *PeerPoolMap) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.PeerPoolMap))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize PeerPoolMap length error: %v", err)
	}
	var peerPoolItemList []*PeerPoolItem
	for _, v := range this.PeerPoolMap {
		peerPoolItemList = append(peerPoolItemList, v)
	}
	sort.SliceStable(peerPoolItemList, func(i, j int) bool {
		return peerPoolItemList[i].PeerPubkey > peerPoolItemList[j].PeerPubkey
	})
	for _, v := range peerPoolItemList {
		if err := v.Serialize(w); err != nil {
			return fmt.Errorf("serialize peerPool error: %v", err)
		}
	}
	return nil
}

func (this *PeerPoolMap) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize PeerPoolMap length error: %v", err)
	}
	peerPoolMap := make(map[string]*PeerPoolItem)
	for i := 0; uint32(i) < n; i++ {
		peerPoolItem := new(PeerPoolItem)
		if err := peerPoolItem.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		peerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
	}
	this.PeerPoolMap = peerPoolMap
	return nil
}

type PeerPoolItem struct {
	PeerPubkey    string         //peer pubkey
	WalletAddress common.Address //peer owner
	Status        Status         //peer status
	TotalInitPos  uint64         //total authorize pos this peer received
}

func (this *PeerPoolItem) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := this.WalletAddress.Serialize(w); err != nil {
		return fmt.Errorf("address.Serialize, serialize address error: %v", err)
	}
	if err := this.Status.Serialize(w); err != nil {
		return fmt.Errorf("this.Status.Serialize, serialize Status error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.TotalInitPos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize totalPos error: %v", err)
	}
	return nil
}

func (this *PeerPoolItem) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address := new(common.Address)

	err = address.Deserialize(r)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	status := new(Status)
	err = status.Deserialize(r)
	if err != nil {
		return fmt.Errorf("status.Deserialize. deserialize status error: %v", err)
	}
	totalPos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize totalPos error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.WalletAddress = *address
	this.Status = *status
	this.TotalInitPos = totalPos
	return nil
}
