package types

import (
	"io"

	"github.com/saveio/themis/common"
	comm "github.com/saveio/themis/p2pserver/common"
)

type SubmitNonceParam struct {
	View        uint32
	Address     []byte
	Id          int64
	Nonce       uint64
	Deadline    uint64
	PlotName    string
	Difficulty  int64
	VoteConsPub []string
	VoteId      []uint32
	VoteInfo    []byte
	MoveUpElect bool
}

//Serialize message payload
func (this SubmitNonceParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.View)
	sink.WriteBytes(this.Address)
	sink.WriteInt64(this.Id)
	sink.WriteUint64(this.Nonce)
	sink.WriteUint64(this.Deadline)
	sink.WriteBytes([]byte(this.PlotName))
	sink.WriteInt64(this.Difficulty)
	sink.WriteUint64(uint64(len(this.VoteConsPub)))
	for _, pub := range this.VoteConsPub {
		sink.WriteBytes([]byte(pub))
	}

	sink.WriteUint64(uint64(len(this.VoteId)))
	for _, id := range this.VoteId {
		sink.WriteUint32(id)
	}
	sink.WriteBytes(this.VoteInfo)
	sink.WriteBool(this.MoveUpElect)
}

func (this *SubmitNonceParam) CmdType() string {
	return comm.SUBMIT_NONECE_PARAM
}

//Deserialize message payload
func (this *SubmitNonceParam) Deserialization(source *common.ZeroCopySource) error {

	var err error
	var eof bool

	this.View, err = source.ReadUint32()
	if err != nil {
		return err
	}
	this.Address, err = source.ReadVarBytes()
	if err != nil {
		return err
	}

	this.Id, eof = source.NextInt64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.Nonce, eof = source.NextUint64()

	if eof {
		return io.ErrUnexpectedEOF
	}

	this.Deadline, eof = source.NextUint64()

	if eof {
		return io.ErrUnexpectedEOF
	}

	plotName, err := source.ReadVarBytes()
	if err != nil {
		return err
	}

	this.PlotName = string(plotName)

	this.Difficulty, eof = source.NextInt64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	voteConsPubLen, eof := source.NextUint64()

	if eof {
		return io.ErrUnexpectedEOF
	}

	voteConsPubs := make([]string, 0, voteConsPubLen)
	for i := 0; i < int(voteConsPubLen); i++ {
		pub, err := source.ReadVarBytes()
		if err != nil {
			return err
		}

		voteConsPubs = append(voteConsPubs, string(pub))
	}
	this.VoteConsPub = voteConsPubs

	voteIdLen, eof := source.NextUint64()

	if eof {
		return io.ErrUnexpectedEOF
	}

	voteIds := make([]uint32, 0, voteConsPubLen)
	for i := 0; i < int(voteIdLen); i++ {
		id, err := source.ReadUint32()
		if err != nil {
			return err
		}

		voteIds = append(voteIds, id)
	}
	this.VoteId = voteIds

	this.VoteInfo, err = source.ReadVarBytes()
	if err != nil {
		return err
	}
	this.MoveUpElect, _, eof = source.NextBool()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}
