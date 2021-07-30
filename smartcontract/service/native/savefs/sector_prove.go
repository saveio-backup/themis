package savefs

import (
	"fmt"
	"github.com/saveio/themis/common"
	"io"

	"github.com/saveio/themis/smartcontract/service/native/savefs/pdp"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

const SECTOR_PROVE_BLOCK_NUM = 32

type SectorProve struct {
	NodeAddr        common.Address
	SectorID        uint64
	ChallengeHeight uint64 // challenge height is used to generate the challenge for prove calculation/verification
	ProveData       []byte
}

func (this *SectorProve) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.NodeAddr); err != nil {
		return fmt.Errorf("[SectorProve] [NodeAddr:%v] serialize from error:%v", this.NodeAddr, err)
	}
	if err := utils.WriteVarUint(w, this.SectorID); err != nil {
		return fmt.Errorf("[SectorProve] [SectorID:%v] serialize from error:%v", this.SectorID, err)
	}
	if err := utils.WriteVarUint(w, this.ChallengeHeight); err != nil {
		return fmt.Errorf("[SectorProve] [ChallengeHeight:%v] serialize from error:%v", this.ChallengeHeight, err)
	}
	if err := utils.WriteBytes(w, this.ProveData); err != nil {
		return fmt.Errorf("[SectorProve] [ProveData:%v] serialize from error:%v", this.ProveData, err)
	}
	return nil
}

func (this *SectorProve) Deserialize(r io.Reader) error {
	var err error
	if this.NodeAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("[SectorProve] [NodeAddr] deserialize from error:%v", err)
	}
	if this.SectorID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorProve] [SectorID] deserialize from error:%v", err)
	}
	if this.ChallengeHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorProve] [ChallengeHeight] deserialize from error:%v", err)
	}
	if this.ProveData, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[SectorProve] [ProveData] deserialize from error:%v", err)
	}
	return nil
}

func (this *SectorProve) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.NodeAddr)
	utils.EncodeVarUint(sink, this.SectorID)
	utils.EncodeVarUint(sink, this.ChallengeHeight)
	utils.EncodeBytes(sink, this.ProveData)
}

func (this *SectorProve) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.NodeAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.SectorID, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ChallengeHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ProveData, err = utils.DecodeBytes(source)
	if err != nil {
		return err
	}
	return nil
}

type SectorProveData struct {
	ProveFileNum uint64
	BlockNum     uint64
	Proofs       []byte
	Tags         []pdp.Tag // tags for challenged blocks
	MerklePath   []*pdp.MerklePath
	PlotData     []byte
}

func (this *SectorProveData) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ProveFileNum); err != nil {
		return fmt.Errorf("[SectorProveData] [ProveFileNum:%v] serialize from error:%v", this.ProveFileNum, err)
	}
	if err := utils.WriteVarUint(w, this.BlockNum); err != nil {
		return fmt.Errorf("[SectorProveData] [BlockNum:%v] serialize from error:%v", this.BlockNum, err)
	}
	if err := utils.WriteBytes(w, this.Proofs); err != nil {
		return fmt.Errorf("[SectorProveData] [Proofs:%v] serialize from error:%v", this.Proofs, err)
	}
	if this.BlockNum != uint64(len(this.Tags)) || this.BlockNum != uint64(len(this.MerklePath)) {
		return fmt.Errorf("[SectorProveData] BlockNum, tags and MerklePath length no match")
	}
	for _, tag := range this.Tags {
		if err := utils.WriteBytes(w, tag[:]); err != nil {
			return fmt.Errorf("[SectorProveData] [Tags:%v] serialize from error:%v", tag, err)
		}
	}
	for _, path := range this.MerklePath {
		if err := path.Serialize(w); err != nil {
			return fmt.Errorf("[SectorProveData] [MerklePath:%v] serialize from error:%v", path, err)
		}
	}
	if err := utils.WriteBytes(w, this.PlotData); err != nil {
		return fmt.Errorf("[SectorProveData] [PlotData:%v] serialize from error:%v", this.PlotData, err)
	}
	return nil
}

func (this *SectorProveData) Deserialize(r io.Reader) error {
	var err error
	if this.ProveFileNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorProveData] [ProveFileNum] deserialize from error:%v", err)
	}
	if this.ProveFileNum == 0 {
		return fmt.Errorf("[SectorProveData] ProveFileNum is 0")
	}
	if this.BlockNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[SectorProveData] [BlockNum] deserialize from error:%v", err)
	}
	if this.BlockNum == 0 {
		return fmt.Errorf("[SectorProveData] BlockNum is 0")
	}

	if this.Proofs, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[SectorProveData] [Proofs] deserialize from error:%v", err)
	}

	tags := make([]pdp.Tag, 0)
	path := make([]*pdp.MerklePath, 0)
	for i := uint64(0); i < this.BlockNum; i++ {
		var tag pdp.Tag
		var data []byte
		if data, err = utils.ReadBytes(r); err != nil {
			return fmt.Errorf("[ProveData] [Tag] deserialize from error:%v", err)
		}
		if len(data) != pdp.TAG_LENGTH {
			return fmt.Errorf("[ProveData] [Tag] wrong tag length")
		}
		copy(tag[:], data[:])
		tags = append(tags, tag)
	}
	this.Tags = tags

	for i := uint64(0); i < this.BlockNum; i++ {
		p := new(pdp.MerklePath)
		if err = p.Deserialize(r); err != nil {
			return fmt.Errorf("[ProveData] [MerklePath] deserialize from error:%v", err)
		}
		path = append(path, p)
	}
	this.MerklePath = path

	if this.PlotData, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[SectorProveData] [PlotData] deserialize from error:%v", err)
	}
	return nil
}
