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

package vbft

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"time"

	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	consutils "github.com/saveio/themis/consensus/utils"
	vconfig "github.com/saveio/themis/consensus/vbft/config"
	"github.com/saveio/themis/core/ledger"
	"github.com/saveio/themis/core/signature"
	"github.com/saveio/themis/core/states"
	scommon "github.com/saveio/themis/core/store/common"
	"github.com/saveio/themis/core/store/overlaydb"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/crypto/vrf"
	gov "github.com/saveio/themis/smartcontract/service/native/governance"
	nutils "github.com/saveio/themis/smartcontract/service/native/utils"
)

func SignMsg(account *account.Account, msg ConsensusMsg) ([]byte, error) {

	data, err := msg.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg when signing: %s", err)
	}

	return signature.Sign(account, data)
}

func hashData(data []byte) common.Uint256 {
	t := sha256.Sum256(data)
	f := sha256.Sum256(t[:])
	return common.Uint256(f)
}

func HashMsg(msg ConsensusMsg) (common.Uint256, error) {

	// FIXME: has to do marshal on each call

	data, err := SerializeVbftMsg(msg)
	if err != nil {
		return common.Uint256{}, fmt.Errorf("failed to marshal block: %s", err)
	}

	return hashData(data), nil
}

type seedData struct {
	BlockNum          uint32 `json:"block_num"`
	PrevBlockProposer uint32 `json:"prev_block_proposer"`
	VrfValue          []byte `json:"vrf_value"`
}

func getParticipantSelectionSeed(block *Block) vconfig.VRFValue {

	data, err := json.Marshal(&seedData{
		BlockNum:          block.getBlockNum() + 1,
		PrevBlockProposer: block.getProposer(),
		VrfValue:          block.getVrfValue(),
	})
	if err != nil {
		return vconfig.VRFValue{}
	}

	t := sha512.Sum512(data)
	f := sha512.Sum512(t[:])
	return vconfig.VRFValue(f)
}

type vrfData struct {
	BlockNum uint32 `json:"block_num"`
	PrevVrf  []byte `json:"prev_vrf"`
}

func computeVrf(sk keypair.PrivateKey, blkNum uint32, prevVrf []byte) ([]byte, []byte, error) {
	data, err := json.Marshal(&vrfData{
		BlockNum: blkNum,
		PrevVrf:  prevVrf,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("computeVrf failed to marshal vrfData: %s", err)
	}

	return vrf.Vrf(sk, data)
}

func verifyVrf(pk keypair.PublicKey, blkNum uint32, prevVrf, newVrf, proof []byte) error {
	data, err := json.Marshal(&vrfData{
		BlockNum: blkNum,
		PrevVrf:  prevVrf,
	})
	if err != nil {
		return fmt.Errorf("verifyVrf failed to marshal vrfData: %s", err)
	}

	result, err := vrf.Verify(pk, data, newVrf, proof)
	if err != nil {
		return fmt.Errorf("verifyVrf failed: %s", err)
	}
	if !result {
		return fmt.Errorf("verifyVrf failed")
	}
	return nil
}

func GetVbftConfigInfo(memdb *overlaydb.MemDB) (*config.VBFTConfig, error) {
	//get governance view
	goveranceview, err := GetGovernanceView(memdb)
	if err != nil {
		return nil, err
	}

	//get preConfig
	preCfg := new(gov.PreConfig)
	data, err := GetStorageValue(memdb, ledger.DefLedger, nutils.GovernanceContractAddress, []byte(gov.PRE_CONFIG))
	if err != nil && err != scommon.ErrNotFound {
		return nil, err
	}
	if data != nil {
		err = preCfg.Deserialization(common.NewZeroCopySource(data))
		if err != nil {
			return nil, err
		}
	}

	var chainconfig *config.VBFTConfig
	if preCfg.SetView == goveranceview.View {
		chainconfig = &config.VBFTConfig{
			N:                    uint32(preCfg.Configuration.N),
			C:                    uint32(preCfg.Configuration.C),
			K:                    uint32(preCfg.Configuration.K),
			L:                    uint32(preCfg.Configuration.L),
			BlockMsgDelay:        uint32(preCfg.Configuration.BlockMsgDelay),
			HashMsgDelay:         uint32(preCfg.Configuration.HashMsgDelay),
			PeerHandshakeTimeout: uint32(preCfg.Configuration.PeerHandshakeTimeout),
			MaxBlockChangeView:   uint32(preCfg.Configuration.MaxBlockChangeView),
		}
	} else {
		data, err := GetStorageValue(memdb, ledger.DefLedger, nutils.GovernanceContractAddress, []byte(gov.VBFT_CONFIG))
		if err != nil {
			return nil, err
		}
		cfg := new(gov.Configuration)
		err = cfg.Deserialization(common.NewZeroCopySource(data))
		if err != nil {
			return nil, err
		}
		chainconfig = &config.VBFTConfig{
			N:                    uint32(cfg.N),
			C:                    uint32(cfg.C),
			K:                    uint32(cfg.K),
			L:                    uint32(cfg.L),
			BlockMsgDelay:        uint32(cfg.BlockMsgDelay),
			HashMsgDelay:         uint32(cfg.HashMsgDelay),
			PeerHandshakeTimeout: uint32(cfg.PeerHandshakeTimeout),
			MaxBlockChangeView:   uint32(cfg.MaxBlockChangeView),
		}
	}
	return chainconfig, nil
}

func GetPeersConfig(memdb *overlaydb.MemDB) ([]*config.VBFTPeerStakeInfo, error) {
	goveranceview, err := GetGovernanceView(memdb)
	if err != nil {
		return nil, err
	}
	viewBytes := gov.GetUint32Bytes(goveranceview.View)
	key := append([]byte(gov.PEER_POOL), viewBytes...)
	data, err := GetStorageValue(memdb, ledger.DefLedger, nutils.GovernanceContractAddress, key)
	if err != nil {
		return nil, err
	}
	peerMap := &gov.PeerPoolMap{
		PeerPoolMap: make(map[string]*gov.PeerPoolItem),
	}
	err = peerMap.Deserialization(common.NewZeroCopySource(data))
	if err != nil {
		return nil, err
	}
	var peerstakes []*config.VBFTPeerStakeInfo
	for _, id := range peerMap.PeerPoolMap {
		if id.Status == gov.CandidateStatus || id.Status == gov.ConsensusStatus {
			config := &config.VBFTPeerStakeInfo{
				Index:      uint32(id.Index),
				PeerPubkey: id.PeerPubkey,
				InitPos:    id.InitPos + id.TotalPos,
			}
			peerstakes = append(peerstakes, config)
		}
	}
	return peerstakes, nil
}

//get peer for certain view, consider consensus flag
//also need handle cross consensus gov period and consensus group flag
func GetPeersConfigExt(memdb *overlaydb.MemDB) ([]*config.VBFTPeerStakeInfo, error) {
	goveranceview, err := GetGovernanceView(memdb)
	if err != nil {
		return nil, err
	}
	viewBytes := gov.GetUint32Bytes(goveranceview.View)
	storageKey := &states.StorageKey{
		ContractAddress: nutils.GovernanceContractAddress,
		Key:             append([]byte(gov.PEER_POOL), viewBytes...),
	}
	data, err := ledger.DefLedger.GetStorageItem(storageKey.ContractAddress, storageKey.Key)
	if err != nil {
		log.Errorf("GetPeersConfigExt fail to get peer pool for view %d", goveranceview.View)
		return nil, err
	}
	peerMap := &gov.PeerPoolMap{
		PeerPoolMap: make(map[string]*gov.PeerPoolItem),
	}
	err = peerMap.Deserialization(common.NewZeroCopySource(data))
	if err != nil {
		return nil, err
	}
	var items *gov.ConsGroupItems
	govView, err := consutils.GetConsGovView()
	if err != nil {
		return nil, err
	}

	last := gov.IsLastViewInConsGovPeriod(govView, goveranceview.View)
	if last {
		log.Debugf("GetPeersConfigExt reach last view for gov view %d", govView.GovView)

		viewBytes := gov.GetUint32Bytes(govView.GovView + 1)

		storageKey := &states.StorageKey{
			ContractAddress: nutils.GovernanceContractAddress,
			Key:             append([]byte(gov.CONS_GROUP_INFO), viewBytes...),
		}

		data, err := ledger.DefLedger.GetStorageItem(storageKey.ContractAddress, storageKey.Key)
		if err != nil {
			log.Debugf("GetPeersConfigExt fail to get cons group info for gov view %d", govView.GovView+1)
			return nil, err
		}

		items = &gov.ConsGroupItems{
			ConsGroupItems: make(map[string]int),
		}

		err = items.Deserialize(bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}

		log.Debugf("GetPeersConfigExt read cons group for gov view %d", govView.GovView+1)
		for pubkey, _ := range items.ConsGroupItems {
			log.Debugf("GetPeersConfigExt cons group item pubkey %s", pubkey)
		}
	}

	var peerstakes []*config.VBFTPeerStakeInfo
	for _, id := range peerMap.PeerPoolMap {
		if last {
			if _, ok := items.ConsGroupItems[id.PeerPubkey]; !ok {
				continue
			}
		} else if id.ConsStatus&gov.InConsensusGroup == 0 {
			continue
		}

		if id.Status == gov.CandidateStatus || id.Status == gov.ConsensusStatus {
			config := &config.VBFTPeerStakeInfo{
				Index:      uint32(id.Index),
				PeerPubkey: id.PeerPubkey,
				InitPos:    id.InitPos + id.TotalPos,
			}
			peerstakes = append(peerstakes, config)
		}
	}
	return peerstakes, nil
}

func isUpdate(memdb *overlaydb.MemDB, view uint32) (bool, error) {
	goveranceview, err := GetGovernanceView(memdb)
	if err != nil {
		return false, err
	}
	if goveranceview.View > view {
		return true, nil
	}
	return false, nil
}

func getRawStorageItemFromMemDb(memdb *overlaydb.MemDB, addr common.Address, key []byte) (value []byte, unkown bool) {
	rawKey := make([]byte, 0, 1+common.ADDR_LEN+len(key))
	rawKey = append(rawKey, byte(scommon.ST_STORAGE))
	rawKey = append(rawKey, addr[:]...)
	rawKey = append(rawKey, key...)
	return memdb.Get(rawKey)
}

func GetStorageValue(memdb *overlaydb.MemDB, backend *ledger.Ledger, addr common.Address, key []byte) (value []byte, err error) {
	if memdb == nil {
		return backend.GetStorageItem(addr, key)
	}
	rawValue, unknown := getRawStorageItemFromMemDb(memdb, addr, key)
	if unknown {
		return backend.GetStorageItem(addr, key)
	}
	if len(rawValue) == 0 {
		return nil, scommon.ErrNotFound
	}

	value, err = states.GetValueFromRawStorageItem(rawValue)
	return
}

func GetGovernanceView(memdb *overlaydb.MemDB) (*gov.GovernanceView, error) {
	value, err := GetStorageValue(memdb, ledger.DefLedger, nutils.GovernanceContractAddress, []byte(gov.GOVERNANCE_VIEW))
	if err != nil {
		return nil, err
	}
	governanceView := new(gov.GovernanceView)
	err = governanceView.Deserialize(bytes.NewBuffer(value))
	if err != nil {
		return nil, err
	}
	return governanceView, nil
}

func getChainConfig(memdb *overlaydb.MemDB, blkNum uint32, winner *gov.SubmitNonceParam) (*vconfig.ChainConfig, error) {
	config, err := GetVbftConfigInfo(memdb)
	if err != nil {
		return nil, fmt.Errorf("failed to get chainconfig from leveldb: %s", err)
	}

	peersinfo, err := GetPeersConfigExt(memdb)
	if err != nil {
		return nil, fmt.Errorf("failed to get peersinfo from leveldb: %s", err)
	}
	goverview, err := GetGovernanceView(memdb)
	if err != nil {
		return nil, fmt.Errorf("failed to get governanceview failed:%s", err)
	}

	log.Debugf("getChainConfig call createWinnerInfo for view: %v", winner.View)
	winnerInfo, err := createWinnerInfo(winner)
	if err != nil {
		return nil, fmt.Errorf("getChainConfig, failed to get winnerInfo for view:%d, err:%s", goverview.View, err)
	}

	cfg, err := GenesisChainConfig(memdb, config, peersinfo, goverview.TxHash, blkNum, winnerInfo)
	if err != nil {
		return nil, fmt.Errorf("GenesisChainConfig failed: %s", err)
	}
	cfg.View = goverview.View
	return cfg, err
}
func createWinnerInfo(winner *gov.SubmitNonceParam) (*gov.WinnerInfo, error) {

	winnerInfo := &gov.WinnerInfo{
		View:     winner.View,
		Address:  winner.Address,
		Deadline: winner.Deadline,

		//vote info
		VoteConsPub: winner.VoteConsPub,
		VoteId:      winner.VoteId,
		VoteInfo:    winner.VoteInfo,
	}

	//use max uint64 as Deadline for dummy param
	if winner.Id == 0 {
		winnerInfo.Deadline = math.MaxUint64
	}

	return winnerInfo, nil
}

func Shuffle_hash(txid common.Uint256, height uint32, id string, idx int) (uint64, error) {
	data, err := json.Marshal(struct {
		Txid   common.Uint256 `json:"txid"`
		Height uint32         `json:"height"`
		NodeID string         `json:"node_id"`
		Index  int            `json:"index"`
	}{txid, height, id, idx})
	if err != nil {
		return 0, err
	}

	hash := fnv.New64a()
	hash.Write(data)
	return hash.Sum64(), nil
}

//GenesisChainConfig return chainconfig
func GenesisChainConfig(memdb *overlaydb.MemDB, vbftconfig *config.VBFTConfig, peersinfo []*config.VBFTPeerStakeInfo, txhash common.Uint256, height uint32,
	winnerInfo *gov.WinnerInfo) (*vconfig.ChainConfig, error) {

	peers := peersinfo
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].InitPos > peers[j].InitPos {
			return true
		} else if peers[i].InitPos == peers[j].InitPos {
			return peers[i].PeerPubkey > peers[j].PeerPubkey
		}
		return false
	})
	log.Debugf("sorted peers: %v", peers)

	//shuttle with PoC winnerInfo after sort by deposit and pubkey
	if winnerInfo != nil {
		bf := new(bytes.Buffer)
		if err := winnerInfo.Serialize(bf); err != nil {
			return nil, fmt.Errorf("GenesisChainConfig, serialize miningInfo error: %v", err)
		}
		info := bf.Bytes()
		governanceView, err := GetGovernanceView(memdb)
		if err != nil {
			return nil, fmt.Errorf("GenesisChainConfig, get GovernanceView error: %v", err)
		}

		log.Debugf("GenesisChainConfig finish sorting using deposit info of view:%d", governanceView.View)
		for i := 0; i < len(peers); i++ {
			log.Debugf("peer[%d] Index %d", i, peers[i].Index)
		}

		hash := fnv.New64a()
		for i := len(peers) - 1; i > 0; i-- {
			data, err := json.Marshal(struct {
				Info  []byte `json:"info"`
				View  uint32 `json:"view"`
				Index int    `json:"index"`
			}{info, governanceView.View, i})
			if err != nil {
				return nil, fmt.Errorf("GenesisChainConfig, generate random num error: %v", err)
			}

			hash.Reset()
			hash.Write(data)
			h := hash.Sum64()
			j := h % uint64(i)
			peers[i], peers[j] = peers[j], peers[i]
		}
		log.Debugf("GenesisChainConfig finish shuffle using poc winner info of view:%d", governanceView.View)
		for i := 0; i < len(peers); i++ {
			log.Debugf("peer[%d] Index %d\n", i, peers[i].Index)
		}
	}

	// get stake sum of top-k peers
	var sum uint64
	for i := 0; i < int(vbftconfig.K); i++ {
		sum += peers[i].InitPos
		log.Debugf("peer: %d, stack: %d", peers[i].Index, peers[i].InitPos)
	}

	log.Debugf("sum of top K stakes: %d", sum)

	// calculate peer ranks
	scale := vbftconfig.L/vbftconfig.K - 1
	if scale <= 0 {
		return nil, fmt.Errorf("L is equal or less than K")
	}

	peerRanks := make([]uint64, 0)
	for i := 0; i < int(vbftconfig.K); i++ {
		//may need change s to bigger value
		var s uint64 = 1
		if sum > 0 && peers[i].InitPos > 0 {
			s = uint64(math.Ceil(float64(peers[i].InitPos) * float64(scale) * float64(vbftconfig.K) / float64(sum)))
		}
		peerRanks = append(peerRanks, s)
	}

	log.Debugf("peers rank table: %v", peerRanks)

	// calculate pos table
	chainPeers := make(map[uint32]*vconfig.PeerConfig, 0)
	posTable := make([]uint32, 0)
	for i := 0; i < int(vbftconfig.K); i++ {
		nodeId := peers[i].PeerPubkey
		chainPeers[peers[i].Index] = &vconfig.PeerConfig{
			Index: peers[i].Index,
			ID:    nodeId,
		}
		for j := uint64(0); j < peerRanks[i]; j++ {
			posTable = append(posTable, peers[i].Index)
		}
	}
	// shuffle
	for i := len(posTable) - 1; i > 0; i-- {
		h, err := Shuffle_hash(txhash, height, chainPeers[posTable[i]].ID, i)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash value: %s", err)
		}
		j := h % uint64(i)
		posTable[i], posTable[j] = posTable[j], posTable[i]
	}
	log.Debugf("init pos table: %v", posTable)

	// generate chain config, and save to ChainConfigFile
	peerCfgs := make([]*vconfig.PeerConfig, 0)
	for i := 0; i < int(vbftconfig.K); i++ {
		peerCfgs = append(peerCfgs, chainPeers[peers[i].Index])
	}

	chainConfig := &vconfig.ChainConfig{
		Version:              1,
		View:                 1,
		N:                    vbftconfig.K,
		C:                    vbftconfig.C,
		BlockMsgDelay:        time.Duration(vbftconfig.BlockMsgDelay) * time.Millisecond,
		HashMsgDelay:         time.Duration(vbftconfig.HashMsgDelay) * time.Millisecond,
		PeerHandshakeTimeout: time.Duration(vbftconfig.PeerHandshakeTimeout) * time.Second,
		Peers:                peerCfgs,
		PosTable:             posTable,
		MaxBlockChangeView:   vbftconfig.MaxBlockChangeView,
	}
	return chainConfig, nil
}
