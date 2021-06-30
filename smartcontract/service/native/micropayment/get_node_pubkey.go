package micropayment

import (
	"fmt"
	"io"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type NodePubKey struct {
	Participant common.Address
	PublicKey   []byte
}

func (this *NodePubKey) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Participant); err != nil {
		return fmt.Errorf("[GetChanInfo] [Participant:%v] serialize from error:%v", this.Participant, err)
	}
	if err := utils.WriteBytes(w, this.PublicKey); err != nil {
		return fmt.Errorf("[GetChanInfo] [PublicKey:%v] serialize from error:%v", this.PublicKey, err)
	}
	return nil
}

func (this *NodePubKey) Deserialize(r io.Reader) error {
	var err error
	if this.Participant, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[GetChanInfo] [Participant] deserialize from error:%v", err)
	}
	if this.PublicKey, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[GetChanInfo] [PublicKey] deserialize from error:%v", err)
	}
	return nil
}

func (this *NodePubKey) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Participant)
	utils.EncodeBytes(sink, this.PublicKey)
}

func (this *NodePubKey) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Participant, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.PublicKey, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}
