/*
 * Copyright (C) 2019 The themis Authors
 * This file is part of The themis library.
 *
 * The themis is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The themis is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */

package vconfig

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/saveio/themis/common"
)

var (
	Version uint32 = 1
)

type PeerConfig struct {
	Index uint32 `json:"index"`
	ID    string `json:"id"`
}

type ChainConfig struct {
	Version              uint32        `json:"version"` // software version
	View                 uint32        `json:"view"`    // config-updated version
	N                    uint32        `json:"n"`       // network size
	C                    uint32        `json:"c"`       // consensus quorum
	BlockMsgDelay        time.Duration `json:"block_msg_delay"`
	HashMsgDelay         time.Duration `json:"hash_msg_delay"`
	PeerHandshakeTimeout time.Duration `json:"peer_handshake_timeout"`
	Peers                []*PeerConfig `json:"peers"`
	PosTable             []uint32      `json:"pos_table"`
	MaxBlockChangeView   uint32        `json:"MaxBlockChangeView"`
}

//
// VBFT consensus payload, stored on each block header
//
type VbftBlockInfo struct {
	Proposer           uint32       `json:"leader"`
	VrfValue           []byte       `json:"vrf_value"`
	VrfProof           []byte       `json:"vrf_proof"`
	LastConfigBlockNum uint32       `json:"last_config_block_num"`
	NewChainConfig     *ChainConfig `json:"new_chain_config"`
}

const (
	VRF_SIZE            = 64 // bytes
	MAX_PROPOSER_COUNT  = 32
	MAX_ENDORSER_COUNT  = 240
	MAX_COMMITTER_COUNT = 240
)

type VRFValue [VRF_SIZE]byte

var NilVRF = VRFValue{}

func (v VRFValue) Bytes() []byte {
	return v[:]
}

func (v VRFValue) IsNil() bool {
	return bytes.Compare(v.Bytes(), NilVRF.Bytes()) == 0
}

func VerifyChainConfig(cfg *ChainConfig) error {

	// TODO

	return nil
}

//Serialize the ChainConfig
func (cc *ChainConfig) Serialize(w io.Writer) error {
	data, err := json.Marshal(cc)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

func (pc *PeerConfig) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(pc.Index)
	sink.WriteString(pc.ID)
}

func (pc *PeerConfig) Deserialization(source *common.ZeroCopySource) error {
	index, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("Deserialization PeerConfig index err:%s", io.ErrUnexpectedEOF)
	}
	pc.Index = index

	nodeid, _, irregular, eof := source.NextString()
	if irregular || eof {
		return fmt.Errorf("serialization PeerConfig nodeid irregular:%v, eof:%v", irregular, eof)
	}
	pc.ID = nodeid
	return nil
}

func (cc *ChainConfig) Hash() common.Uint256 {
	buf := new(bytes.Buffer)
	cc.Serialize(buf)
	hash := sha256.Sum256(buf.Bytes())
	return hash
}
