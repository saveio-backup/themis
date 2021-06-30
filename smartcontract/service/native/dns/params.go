/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-18
 */
package dns

import (
	"fmt"
	"io"
	"math"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/serialization"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type RegisterCandidateParam struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint32
	Caller     []byte
	KeyNo      uint32
}

func (this *RegisterCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, request peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.InitPos)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize initPos error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Caller); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize caller error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.KeyNo)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize keyNo error: %v", err)
	}
	return nil
}

func (this *RegisterCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	initPos, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize initPos error: %v", err)
	}
	if initPos > math.MaxUint32 {
		return fmt.Errorf("initPos larger than max of uint32")
	}
	caller, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize caller error: %v", err)
	}
	keyNo, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize keyNo error: %v", err)
	}
	if keyNo > math.MaxUint32 {
		return fmt.Errorf("initPos larger than max of uint32")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.InitPos = uint32(initPos)
	this.Caller = caller
	this.KeyNo = uint32(keyNo)
	return nil
}

type UnRegisterCandidateParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *UnRegisterCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, request peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	return nil
}

func (this *UnRegisterCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type QuitNodeParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *QuitNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, deserialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	return nil
}

func (this *QuitNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type ApproveCandidateParam struct {
	PeerPubkey string
}

func (this *ApproveCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	return nil
}

func (this *ApproveCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type PubKeyParam struct {
	PeerPubkey string
}

func (this *PubKeyParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	return nil
}

func (this *PubKeyParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type WithdrawParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *WithdrawParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, deserialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	return nil
}

func (this *WithdrawParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type WithdrawFeeParam struct {
	Address common.Address
}

func (this *WithdrawFeeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	return nil
}

func (this *WithdrawFeeParam) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.Address = address
	return nil
}

type PromisePos struct {
	PeerPubkey string
	PromisePos uint64
}

func (this *PromisePos) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := utils.WriteVarUint(w, this.PromisePos); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize promisePos error: %v", err)
	}
	return nil
}

func (this *PromisePos) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	promisePos, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadBool, deserialize promisePos error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.PromisePos = promisePos
	return nil
}

type ChangeInitPosParam struct {
	PeerPubkey string
	Address    common.Address
	Pos        uint64
}

func (this *ChangeInitPosParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	if err := utils.WriteVarUint(w, this.Pos); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize pos error: %v", err)
	}
	return nil
}

func (this *ChangeInitPosParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	pos, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadBool, deserialize pos error: %v", err)
	}
	if pos > math.MaxUint32 {
		return fmt.Errorf("pos larger than max of uint32")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.Pos = pos
	return nil
}

type UpdateNodeParam struct {
	WalletAddr common.Address
	IP         []byte
	Port       []byte
}

func (this *UpdateNodeParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.WalletAddr); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [WalletAddr:%v] serialize from error:%v", this.WalletAddr, err)
	}

	if err := utils.WriteBytes(w, this.IP); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [IP:%v] serialize from error:%v", this.IP, err)
	}
	if err := utils.WriteBytes(w, this.Port); err != nil {
		return fmt.Errorf("[DNSNodeInfo] [Port:%v] serialize from error:%v", this.Port, err)
	}
	return nil
}

func (this *UpdateNodeParam) Deserialize(r io.Reader) error {
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

	return nil
}
