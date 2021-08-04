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

package governance

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	cstates "github.com/saveio/themis/core/states"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/smartcontract/service/native"
	fs "github.com/saveio/themis/smartcontract/service/native/savefs"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

// miningView
func GetMiningView(native *native.NativeService, contract common.Address) (*MiningView, error) {
	miningViewBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(MINING_VIEW)))
	if err != nil {
		return nil, fmt.Errorf("GetMiningView, get miningViewBytes error: %v", err)
	}
	miningView := new(MiningView)
	if miningViewBytes == nil {
		return nil, fmt.Errorf("getMiningView, get nil miningViewBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(miningViewBytes)
		if err != nil {
			return nil, fmt.Errorf("GetMiningView, deserialize from raw storage item err:%v", err)
		}
		if err := miningView.Deserialize(bytes.NewBuffer(value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize governanceView error: %v", err)
		}
	}
	return miningView, nil
}

func putMiningView(native *native.NativeService, contract common.Address, miningView *MiningView) error {
	bf := new(bytes.Buffer)
	if err := miningView.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize governanceView error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(MINING_VIEW)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

//MiningViewInfo
func GenMiningViewInfoKey(contract common.Address, view uint32) []byte {
	str := fmt.Sprintf(MINE_VIEW_INFO_KEY_PATTERN, view)
	key := append(contract[:], []byte(str)...)
	return key
}

func getMiningViewInfo(native *native.NativeService, contract common.Address, view uint32) (*MiningViewInfo, error) {
	key := GenMiningViewInfoKey(contract, view)
	miningInfoBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getMiningViewInfo, get miningInfoBytes error: %v", err)
	}
	miningViewInfo := new(MiningViewInfo)
	if miningInfoBytes == nil {
		return nil, fmt.Errorf("getMiningViewInfo, get nil miningInfoBytes for view %d", view)
	} else {
		value, err := cstates.GetValueFromRawStorageItem(miningInfoBytes)
		if err != nil {
			return nil, fmt.Errorf("getMiningViewInfo, deserialize from raw storage item err:%v", err)
		}
		if err := miningViewInfo.Deserialize(bytes.NewBuffer(value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize governanceView error: %v", err)
		}
	}
	return miningViewInfo, nil
}

func putMiningViewInfo(native *native.NativeService, contract common.Address, view uint32, miningViewInfo *MiningViewInfo) error {
	key := GenMiningViewInfoKey(contract, view)
	bf := new(bytes.Buffer)
	if err := miningViewInfo.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize miningInfo error: %v", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

//WinnerInfo
func GenWinnerInfoKey(contract common.Address, view uint32) []byte {
	str := fmt.Sprintf(WINNER_INFO_KEY_PATTERN, view)
	key := append(contract[:], []byte(str)...)
	return key
}

func getWinnerInfo(native *native.NativeService, contract common.Address, view uint32) (*WinnerInfo, error) {
	key := GenWinnerInfoKey(contract, view)
	winnerInfoBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getWinnerInfo, get winnerInfoBytes error: %v", err)
	}
	winnerInfo := new(WinnerInfo)
	if winnerInfoBytes == nil {
		return nil, fmt.Errorf("getWinnerInfo, get nil winnerInfoBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(winnerInfoBytes)
		if err != nil {
			return nil, fmt.Errorf("getWinnerInfo, deserialize from raw storage item err:%v", err)
		}
		if err := winnerInfo.Deserialize(bytes.NewBuffer(value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize getWinnerInfo error: %v", err)
		}
	}
	return winnerInfo, nil
}

func putWinnerInfo(native *native.NativeService, contract common.Address, view uint32, winnerInfo *WinnerInfo) error {
	key := GenWinnerInfoKey(contract, view)
	bf := new(bytes.Buffer)
	if err := winnerInfo.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize miningInfo error: %v", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

//WinnersInfo
func GenWinnersInfoKey(contract common.Address, view uint32) []byte {
	str := fmt.Sprintf(WINNERS_INFO_KEY_PATTERN, view)
	key := append(contract[:], []byte(str)...)
	return key
}

func getWinnersInfo(native *native.NativeService, contract common.Address, view uint32) (*WinnersInfo, error) {
	key := GenWinnersInfoKey(contract, view)
	winnersInfoBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getWinnersInfo, get winnerInfoBytes error: %v", err)
	}
	winnersInfo := new(WinnersInfo)
	if winnersInfoBytes == nil {
		return nil, fmt.Errorf("getWinnersInfo, get nil winnerInfoBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(winnersInfoBytes)
		if err != nil {
			return nil, fmt.Errorf("getWinnerInfo, deserialize from raw storage item err:%v", err)
		}
		if err := winnersInfo.Deserialize(bytes.NewBuffer(value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize getWinnerInfo error: %v", err)
		}
	}
	return winnersInfo, nil
}

func putWinnersInfo(native *native.NativeService, contract common.Address, view uint32, winnersInfo *WinnersInfo) error {
	key := GenWinnersInfoKey(contract, view)
	bf := new(bytes.Buffer)
	if err := winnersInfo.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize miningInfo error: %v", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func calGenerationSignature(lastGenerationSignature common.Uint256, lastGenerator uint64) (common.Uint256, error) {
	buf := make([]byte, 40)
	last := lastGenerationSignature.ToArray()
	buf = append(buf, last...)
	binary.LittleEndian.PutUint64(buf[32:], lastGenerator)

	md := common.NewShabal256()
	md.Update(buf, 0, int64(len(buf)))
	newGenSig := md.Digest()

	generationSignature, err := common.Uint256ParseFromBytes(newGenSig)
	if err != nil {
		return common.Uint256{}, err
	}

	return generationSignature, nil
}

func calculateScoop(view uint64, gensig []byte) uint32 {
	data := make([]byte, 8)

	binary.BigEndian.PutUint64(data[:], view)
	data = append(data, gensig[:]...)

	md := common.NewShabal256()
	md.Update(data, 0, int64(len(data)))
	newGenSig := md.Digest()

	scoop := (uint32(newGenSig[30]&0x0F) << 8) | uint32(newGenSig[31])
	return scoop
}

// verify nonce submitted by account id
func verifyNonce(param *SubmitNonceParam, scoop uint32, baseTarget int64, gensig []byte) bool {
	plot := types.NewMiningPlot(param.Id, param.Nonce)
	scoopData := plot.GetScoopData(int(scoop))

	data := append([]byte{}, gensig[:]...) // gensig 32 bytes
	data = append(data, scoopData[:]...)   // scoop 64 bytes

	md := common.NewShabal256()
	md.Update(data, 0, int64(len(data)))
	hash := md.Digest()

	//same with burst calculateHit
	deadline := binary.LittleEndian.Uint64(hash)
	deadline /= uint64(baseTarget)

	log.Debugf("verifyNonce for view: %d, from id: %d, nonce: %d, deadline: %d\n", param.View, param.Id, param.Nonce, param.Deadline)
	log.Debugf("verifyNonce scoop:%d, baseTarget：%d, deadline calculated: %d\n", scoop, baseTarget, deadline)

	return param.Deadline == deadline
}

// period summary info
func GenPeriodSummaryKey(contract common.Address, period uint32) []byte {
	str := fmt.Sprintf(PERIOD_SUMMARY_KEY_PATTERN, period)
	key := append(contract[:], []byte(str)...)

	return key

}

func getPeriodSummary(native *native.NativeService, period uint32) (*PeriodSummary, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	key := GenPeriodSummaryKey(contract, period)
	periodSummaryBytes, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, fmt.Errorf("get periodSummaryBytes error: %v", err)
	}

	periodSummary := &PeriodSummary{
		MinerWinTimes: make(map[common.Address]int64),
	}

	if periodSummaryBytes != nil {
		if err := periodSummary.Deserialize(bytes.NewBuffer(periodSummaryBytes.Value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize periodInfos error: %v", err)
		}
	}
	return periodSummary, nil
}

//calculate bonus
func GetBlockSubsidy(view uint32) uint64 {
	halvings := uint64(view) / HALVING_INTERVAL

	if halvings >= 64 {
		return 0
	}

	amount := uint64(250 * COIN)

	amount >>= halvings
	return amount
}

//calculate plot file size in MB
func CalPlotFileSize(deadline uint64) uint64 {
	if deadline == 0 {
		return 0
	}

	max := big.NewInt(1)
	max.SetUint64(math.MaxUint64)
	max.Add(max, big.NewInt(1))
	max.Div(max, big.NewInt(1).SetUint64(deadline))

	expectNonces := max
	expectNonces.Mul(expectNonces, big.NewInt(types.PLOT_SIZE))
	expectNonces.Div(expectNonces, big.NewInt(1024))
	expectNonces.Div(expectNonces, big.NewInt(1024))

	return expectNonces.Uint64()

}

//query space provided in FS
func queryVolume(native *native.NativeService, miner common.Address) (uint64, error) {
	bf := new(bytes.Buffer)
	if err := utils.WriteAddress(bf, miner); err != nil {
		return 0, err
	}

	data, err := native.NativeCall(utils.OntFSContractAddress, "FsNodeQuery", bf.Bytes())

	if err != nil {
		return 0, fmt.Errorf("queryVolume, appCall error: %v", err)
	}

	fsNodeInfo := &fs.FsNodeInfo{}

	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		fsNodeInfoReader := bytes.NewReader(retInfo.Info)
		err = fsNodeInfo.Deserialize(fsNodeInfoReader)
		if err != nil {
			return 0, fmt.Errorf("FsNodeQuery error: %s", err.Error())
		}
		return fsNodeInfo.Volume, nil
	} else {
		return 0, errors.New(string(retInfo.Info))
	}

}

func IsLastViewOfDay(view uint32) bool {
	return view%NUM_VIEW_PER_DAY == 0
}

func GetMiningPeriod(view uint32) uint32 {
	return (view + NUM_VIEW_PER_PERIOD - 1) / NUM_VIEW_PER_PERIOD
}

func IsLastViewInMiningPeriod(view uint32) bool {
	period := GetMiningPeriod(view)
	return view == period*NUM_VIEW_PER_PERIOD
}

func transferDelayedBonus(native *native.NativeService, view uint32) error {
	if !IsLastViewOfDay(view) {
		return nil
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//go through winner of last 90 days!
	first := int64(view) - NUM_VIEW_PER_DAY*(NUM_DAY_DELAYED+1) + 1
	if first <= 0 {
		first = 1
	}
	last := int64(view) - NUM_VIEW_PER_DAY
	if last <= 0 {
		return nil
	}

	for curView := uint32(first); curView <= uint32(last); curView++ {
		winnersInfo, err := getWinnersInfo(native, contract, curView)
		if err != nil {
			return fmt.Errorf("transferDelayedBonus, get winnersInfo for view %d error: %v", view, err)
		}

		// send delayed bonus to miner
		log.Debugf("transferDelayedBonus send delayed bonus of view: %d", curView)
		bonus := GetBlockSubsidy(curView - 1)
		winnerAddress := []common.Address{}
		for _, winner := range winnersInfo.Winners {
			winnerAddress = append(winnerAddress, winner.Address)
		}
		err = SplitBonus(native, winnerAddress, curView, bonus, true)
		if err != nil {
			return fmt.Errorf("transferDelayedBonus, SplitBonus fail %s", err)
		}

	}

	return nil
}

//consensus vote info
func GetConsVoteMap(native *native.NativeService, contract common.Address, view uint32) (*ConsVoteMap, error) {
	consVoteMap := &ConsVoteMap{
		ConsVoteMap: make(map[string]*ConsVoteItem),
	}

	viewBytes := GetUint32Bytes(view)

	consVoteMapBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CONS_VOTE_INFO), viewBytes))
	if err != nil {
		return nil, fmt.Errorf("getconsVoteMap, get all consVoteMap error: %v", err)
	}
	if consVoteMapBytes == nil {
		return consVoteMap, nil
	}
	item := cstates.StorageItem{}
	source := common.NewZeroCopySource(consVoteMapBytes)
	err = item.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize consVoteMap error:%v", err)
	}
	consVoteMapStore := item.Value
	if err := consVoteMap.Deserialize(bytes.NewBuffer(consVoteMapStore)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize consVoteMap error: %v", err)
	}
	return consVoteMap, nil
}

func putConsVoteMap(native *native.NativeService, contract common.Address, view uint32, consVoteMap *ConsVoteMap) error {
	bf := new(bytes.Buffer)
	if err := consVoteMap.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize consVoteMap error: %v", err)
	}
	viewBytes := GetUint32Bytes(view)

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CONS_VOTE_INFO), viewBytes), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

//cons vote bonus
func getConsVoteRevenue(native *native.NativeService, contract common.Address) (*ConsVoteRevenue, error) {
	voteRevenueBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CONS_VOTE_REVENUE)))
	if err != nil {
		return nil, fmt.Errorf("getConsVoteRevenue, get voteRevenueBytes error: %v", err)
	}
	voteRevenue := new(ConsVoteRevenue)
	if voteRevenueBytes == nil {
		return nil, fmt.Errorf("getConsVoteRevenue, get nil voteRevenueBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(voteRevenueBytes)
		if err != nil {
			return nil, fmt.Errorf("getConsVoteRevenue, deserialize from raw storage item err:%v", err)
		}
		if err := voteRevenue.Deserialize(bytes.NewBuffer(value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize voteRevenue error: %v", err)
		}
	}
	return voteRevenue, nil
}

func putConsVoteRevenue(native *native.NativeService, contract common.Address, voteRevenue *ConsVoteRevenue) error {
	bf := new(bytes.Buffer)
	if err := voteRevenue.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize voteRevenue error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CONS_VOTE_REVENUE)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func increaseConsVoteRevenue(native *native.NativeService, contract common.Address, amount uint64) error {
	voteRevenue, err := getConsVoteRevenue(native, contract)
	if err != nil {
		return fmt.Errorf("increaseConsVoteRevenue, get vote revenue error: %v", err)
	}
	voteRevenue.Total += amount
	putConsVoteRevenue(native, contract, voteRevenue)
	return nil
}

func handleConsElectMoveUp(native *native.NativeService, view uint32) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	consGovView, err := GetConsGovView(native, contract)
	if err != nil {
		return fmt.Errorf("getConsGovView, error: %v", err)
	}

	if !IsDuringRunning(consGovView, view) {
		return nil
	}

	log.Debugf("handleConsElectMoveUp, before update cons gov view %v", consGovView)

	consGovView.ReelectStartView = view + 1
	err = putConsGovView(native, contract, consGovView)
	if err != nil {
		return fmt.Errorf("handleConsElectMoveUp, put cons gov view error: %v", err)
	}

	log.Debugf("handleConsElectMoveUp, after update cons gov view %v", consGovView)

	return nil
}

//notify node to pledge in order to envolve consensus election
func notifyConsPledge(native *native.NativeService, view uint32) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	consGovView, err := GetConsGovView(native, contract)
	if err != nil {
		return fmt.Errorf("getConsGovView, error: %v", err)
	}

	if !IsLastRunningView(consGovView, view) {
		return nil
	}

	//EventNotify
	pledge := &pledgeForConsEvent{
		ConsGovPeriod: consGovView.GovView,
	}

	PledgeForConsEvent(native, pledge)
	return nil
}

//notify miner to vote for consensus votee(for next gov view)
func notifyConsVote(native *native.NativeService, view uint32) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	consGovView, err := GetConsGovView(native, contract)
	if err != nil {
		return fmt.Errorf("notifyConsVote, get cons gov view error: %v", err)
	}
	if !IsLastViewForPledge(consGovView, view) {
		return nil
	}

	voteeNodes := []string{}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return fmt.Errorf("notifyConsVote, get peerPoolMap error: %v", err)
	}

	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.ConsStatus&ElectForConsensus == 0 {
			continue
		} else {
			voteeNodes = append(voteeNodes, peerPoolItem.PeerPubkey)
		}
	}

	//EventNotify
	voteeEvent := &voteeConsNodesEvent{
		ConsGovPeriod: consGovView.GovView + 1,
		Votees:        voteeNodes,
	}

	VoteeConsNodesEvent(native, voteeEvent)

	return nil
}

//update cons vote info with info in winner info
func handleConsVote(native *native.NativeService, view uint32, voter common.Address, votee []string) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	consGovView, err := GetConsGovView(native, contract)
	if err != nil {
		return fmt.Errorf("handleConsVote, get cons gov view error: %v", err)
	}

	if !IsDuringConsElect(consGovView, view) {
		return nil
	}

	consVoteMap, err := GetConsVoteMap(native, contract, consGovView.GovView)
	if err != nil {
		return fmt.Errorf("handleVote, get consVoteMap error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return fmt.Errorf("handleConsVote, get peerPoolMap error: %v", err)
	}

	detail, err := GetConsVoteDetail(native, contract, consGovView.GovView, voter)
	if err != nil {
		return fmt.Errorf("handleConsVote, get vote detail error: %v", err)
	}

	for i := 0; i < len(votee); i++ {
		//check if exist in PeerPool
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[votee[i]]
		if !ok {
			continue
		}
		if peerPoolItem.ConsStatus&ElectForConsensus == 0 {
			log.Debugf("handleConsVote, peer with pubkey %s doesn't ask for elect", votee[i])
			continue

		}

		if _, found := detail.ConsVoteDetail[votee[i]]; !found {
			detail.ConsVoteDetail[votee[i]] = 1
			if len(detail.ConsVoteDetail) > NUM_NODE_PER_CONS_ELECT_PERIOD {
				log.Debugf("handleConsVote, exceed max allowed num of votee during one consensus gov period")
				continue
			}
		}

		var item *ConsVoteItem
		item, ok = consVoteMap.ConsVoteMap[votee[i]]
		if !ok {
			item = &ConsVoteItem{
				PeerPubkey: votee[i],
				NumVotes:   0,
				VoterMap:   make(map[common.Address]uint32),
			}
			consVoteMap.ConsVoteMap[votee[i]] = item
		}

		if num, ok := item.VoterMap[voter]; ok {
			num++
			item.VoterMap[voter] = num
		} else {
			item.VoterMap[voter] = 1
		}

		log.Debugf("handleConsVote, pubkey %s get %d vote from %s", votee[i], item.VoterMap[voter], voter.ToBase58())

		item.NumVotes++
		log.Debugf("handleConsVote, pubkey %s get total votes %d", votee[i], item.NumVotes)

	}

	err = putConsVoteDetail(native, contract, consGovView.GovView, voter, detail)
	if err != nil {
		return fmt.Errorf("putConsVoteDetail error: %v", err)
	}

	err = putConsVoteMap(native, contract, consGovView.GovView, consVoteMap)
	if err != nil {
		return fmt.Errorf("putConsVoteMap error: %v", err)
	}

	return nil
}

func getElectEndView(govView *ConsGovView) uint32 {
	consElectStartView := govView.ReelectStartView + NUM_VIEW_PER_CONS_PLEDGE_PERIOD
	consElectEndView := consElectStartView + NUM_VIEW_PER_CONS_ELECT_PERIOD - 1
	return consElectEndView
}

func IsDuringRunning(govView *ConsGovView, view uint32) bool {
	return view >= govView.RunningStartView && view < govView.ReelectStartView
}

func IsLastRunningView(govView *ConsGovView, view uint32) bool {
	return govView.ReelectStartView-1 == view
}

func IsDuringGovElect(govView *ConsGovView, view uint32) bool {
	govElectStartView := govView.ReelectStartView + NUM_VIEW_PER_CONS_PLEDGE_PERIOD + NUM_VIEW_PER_CONS_ELECT_PERIOD
	govElectEndView := govElectStartView + NUM_VIEW_PER_GOV_ELECT_PERIOD - 1
	return view >= govElectStartView && view <= govElectEndView
}

func IsLastViewForPledge(govView *ConsGovView, view uint32) bool {
	consPledgeEndView := govView.ReelectStartView + NUM_VIEW_PER_CONS_PLEDGE_PERIOD - 1
	return view == consPledgeEndView
}

func IsDuringConsElect(govView *ConsGovView, view uint32) bool {
	consElectStartView := govView.ReelectStartView + NUM_VIEW_PER_CONS_PLEDGE_PERIOD
	consElectEndView := consElectStartView + NUM_VIEW_PER_CONS_ELECT_PERIOD - 1
	return view >= consElectStartView && view <= consElectEndView-NUM_VIEW_COLD_DOWN
}

//func IsLastViewInConsGovPeriod(native *native.NativeService, view uint32) (bool, error) {
func IsLastViewInConsGovPeriod(govView *ConsGovView, view uint32) bool {
	consElectEndView := govView.ReelectStartView + NUM_VIEW_PER_CONS_PLEDGE_PERIOD + NUM_VIEW_PER_CONS_ELECT_PERIOD
	lastView := consElectEndView + NUM_VIEW_PER_GOV_ELECT_PERIOD
	return view == lastView-1
}

//update cons vote info with info in winner info
func electConsGroup(native *native.NativeService, view uint32) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	govView, err := GetConsGovView(native, contract)
	if err != nil {
		return fmt.Errorf("getView, get view error: %v", err)
	}

	electEndView := getElectEndView(govView)

	if view != electEndView {
		return nil
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	//consVoteMap, err := GetConsVoteMap(native, contract, consGovView)
	consVoteMap, err := GetConsVoteMap(native, contract, govView.GovView)
	if err != nil {
		return fmt.Errorf("electConsGroup, get consVoteMap error: %v", err)
	}

	//need collect all nodes take part in elect since some of them may get zero vote!
	var consVoteList []*ConsVoteItem
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.ConsStatus&ElectForConsensus == 0 {
			continue
		}
		if _, ok := consVoteMap.ConsVoteMap[peerPoolItem.PeerPubkey]; ok {
			continue
		}

		consVoteMap.ConsVoteMap[peerPoolItem.PeerPubkey] = &ConsVoteItem{
			PeerPubkey: peerPoolItem.PeerPubkey,
			VoterMap:   make(map[common.Address]uint32)}
	}

	//ignore node with 0 vote
	for _, v := range consVoteMap.ConsVoteMap {
		if v.NumVotes > 0 {
			consVoteList = append(consVoteList, v)
		}
	}

	sort.SliceStable(consVoteList, func(i, j int) bool {
		if consVoteList[i].NumVotes > consVoteList[j].NumVotes {
			return true
		} else if consVoteList[i].NumVotes == consVoteList[j].NumVotes {
			return consVoteList[i].PeerPubkey > consVoteList[j].PeerPubkey
		}
		return false
	})

	//record consusens node in group for next gov view
	items := &ConsGroupItems{
		ConsGroupItems: make(map[string]int),
	}
	for _, v := range consVoteList {
		if len(items.ConsGroupItems) >= int(globalParam.CandidateNum) {
			break
		}
		items.ConsGroupItems[v.PeerPubkey] = 1
	}
	log.Debugf("electConsGroup,  cons group has %d nodes", len(items.ConsGroupItems))
	for k, _ := range items.ConsGroupItems {
		log.Debugf("electConsGroup, cons group node pubkey: %s", k)
	}

	//use default nodes when nodes less than minimum
	if len(items.ConsGroupItems) < int(globalParam.ConsGroupSize) {

		defNodes, err := GetDefConsNodes(native, contract)
		if err != nil {
			return fmt.Errorf("electConsGroup, get default consensus nodes error: %v", err)
		}

		//sort def nodes by deposit then pubkey
		var peers []*PeerStakeInfo
		for pubkey, _ := range defNodes.DefaultConsNodes {
			if peerPoolItem, ok := peerPoolMap.PeerPoolMap[pubkey]; ok {
				if _, ok := items.ConsGroupItems[peerPoolItem.PeerPubkey]; ok {
					continue
				}

				if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
					peers = append(peers, &PeerStakeInfo{
						Index:      peerPoolItem.Index,
						PeerPubkey: peerPoolItem.PeerPubkey,
						Stake:      peerPoolItem.TotalPos + peerPoolItem.InitPos,
					})
				}
			}
		}

		//try non-def nodes
		var peers2 []*PeerStakeInfo
		for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
			if _, ok := items.ConsGroupItems[peerPoolItem.PeerPubkey]; ok {
				continue
			}

			if _, ok := defNodes.DefaultConsNodes[peerPoolItem.PeerPubkey]; ok {
				continue
			}

			if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
				peers2 = append(peers2, &PeerStakeInfo{
					Index:      peerPoolItem.Index,
					PeerPubkey: peerPoolItem.PeerPubkey,
					Stake:      peerPoolItem.TotalPos + peerPoolItem.InitPos,
				})
			}
		}
		peers = append(peers, peers2...)

		// sort peers by deposit and pubkey
		sort.SliceStable(peers, func(i, j int) bool {
			if peers[i].Stake > peers[j].Stake {
				return true
			} else if peers[i].Stake == peers[j].Stake {
				return peers[i].PeerPubkey > peers[j].PeerPubkey
			}
			return false
		})

		for i := 0; i < len(peers); i++ {
			if _, ok := items.ConsGroupItems[peers[i].PeerPubkey]; ok {
				continue
			} else {
				items.ConsGroupItems[peers[i].PeerPubkey] = 1
				if len(items.ConsGroupItems) >= int(globalParam.ConsGroupSize) {
					break
				}
			}
		}
		//still not enough?
		if len(items.ConsGroupItems) < int(globalParam.ConsGroupSize) {
			return fmt.Errorf("electConsGroup, not enough nodes in consensus group after using default nodes")
		}
	}

	//record nodes in consensus group for next consensus gov view
	log.Debugf("electConsGroup,  prepare put cons group for gov view %d", govView.GovView+1)
	for pubkey, _ := range items.ConsGroupItems {
		log.Debugf("electConsGroup,  items pubkey %s", pubkey)
	}

	err = putConsGroupItems(native, contract, govView.GovView+1, items)
	if err != nil {
		return fmt.Errorf("putConsVoteMap error: %v", err)
	}

	//split cons bonus
	err = splitConsVoteBonus(native, consVoteMap, items)
	if err != nil {
		return fmt.Errorf("splitConsVoteBonus error: %v", err)
	}
	return nil
}

func splitConsVoteBonus(native *native.NativeService, vote *ConsVoteMap, items *ConsGroupItems) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	voteDetail := make(map[common.Address]uint32)
	voteSum := uint32(0)

	for pubkey, voteItem := range vote.ConsVoteMap {
		if _, ok := items.ConsGroupItems[pubkey]; !ok {
			continue
		}

		if voteItem.VoterMap == nil {
			continue
		}

		for address, num := range voteItem.VoterMap {
			if totalVote, ok := voteDetail[address]; ok {
				voteDetail[address] = totalVote + num
			} else {
				voteDetail[address] = num
			}
			voteSum += num
		}
	}

	numVoter := len(voteDetail)

	voteRevenue, err := getConsVoteRevenue(native, contract)
	if err != nil {
		return fmt.Errorf("splitConsVoteBonus, get vote revenue error: %v", err)
	}
	total := voteRevenue.Total

	log.Debugf("splitConsVoteBonus, prepare send total %d(10^-9) bonus to %d nodes", total, numVoter)
	consumed := uint64(0)
	for address, num := range voteDetail {
		amount := total * uint64(num) / uint64(voteSum)

		err := appCallTransferOnt(native, utils.GovernanceContractAddress, address, uint64(amount))
		if err != nil {
			return fmt.Errorf("splitConsVoteBonus, bonus transfer error: %v", err)
		}
		consumed += amount

		log.Debugf("splitConsVoteBonus,  transfer %d(10^-9) bonus to %s", amount, address.ToBase58())
	}

	voteRevenue.Total -= consumed
	err = putConsVoteRevenue(native, contract, voteRevenue)
	if err != nil {
		return fmt.Errorf("splitConsVoteBonus, put cons vote revenue error: %v", err)
	}
	return nil
}

//consensus vote info
func GetConsVoteDetail(native *native.NativeService, contract common.Address, view uint32, voter common.Address) (*ConsVoteDetail, error) {
	consVoteDetail := &ConsVoteDetail{
		ConsVoteDetail: make(map[string]int),
	}

	viewBytes := GetUint32Bytes(view)

	consVoteDetailBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CONS_VOTE_DETAIL), viewBytes, voter[:]))
	if err != nil {
		return nil, fmt.Errorf("GetConsVoteDetail, get vote detail error: %v", err)
	}

	if consVoteDetailBytes != nil {
		item := cstates.StorageItem{}
		source := common.NewZeroCopySource(consVoteDetailBytes)
		err = item.Deserialization(source)
		if err != nil {
			return nil, fmt.Errorf("deserialize ConsVoteDetail error:%v", err)
		}
		consVoteDetailStore := item.Value
		if err := consVoteDetail.Deserialize(bytes.NewBuffer(consVoteDetailStore)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize consGroupItems error: %v", err)
		}
	}
	return consVoteDetail, nil
}

func putConsVoteDetail(native *native.NativeService, contract common.Address, view uint32, voter common.Address, detail *ConsVoteDetail) error {
	bf := new(bytes.Buffer)
	if err := detail.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize consVoteDetail error: %v", err)
	}
	viewBytes := GetUint32Bytes(view)

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CONS_VOTE_DETAIL), viewBytes, voter[:]), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

//consensus group info
func GetConsGroupItems(native *native.NativeService, contract common.Address, view uint32) (*ConsGroupItems, error) {
	consGroupItems := &ConsGroupItems{
		ConsGroupItems: make(map[string]int),
	}

	viewBytes := GetUint32Bytes(view)

	consGroupItemsBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CONS_GROUP_INFO), viewBytes))
	if err != nil {
		return nil, fmt.Errorf("getConsGroupItems, get all consGroupItems error: %v", err)
	}
	if consGroupItemsBytes == nil {
		return nil, fmt.Errorf("getConsGroupItems, consGroupItems is nil")
	}
	item := cstates.StorageItem{}
	source := common.NewZeroCopySource(consGroupItemsBytes)
	err = item.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize ConsGroupItems error:%v", err)
	}
	consGroupItemsStore := item.Value
	if err := consGroupItems.Deserialize(bytes.NewBuffer(consGroupItemsStore)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize consGroupItems error: %v", err)
	}
	return consGroupItems, nil
}

func putConsGroupItems(native *native.NativeService, contract common.Address, view uint32, items *ConsGroupItems) error {
	bf := new(bytes.Buffer)
	if err := items.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize consGroupItems error: %v", err)
	}
	viewBytes := GetUint32Bytes(view)

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CONS_GROUP_INFO), viewBytes), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

//default consensus nodes
func GetDefConsNodes(native *native.NativeService, contract common.Address) (*DefaultConsNodes, error) {
	nodes := &DefaultConsNodes{
		DefaultConsNodes: make(map[string]int),
	}

	nodesBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(DEFAULT_CONS_NODE)))
	if err != nil {
		return nil, fmt.Errorf("GetDefConsNodes, get default consensus nodes error: %v", err)
	}
	if nodesBytes == nil {
		return nil, fmt.Errorf("GetDefConsNodes, default consensus nodes is nil")
	}
	item := cstates.StorageItem{}
	source := common.NewZeroCopySource(nodesBytes)
	err = item.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize default consensus nodes error:%v", err)
	}
	nodesStore := item.Value
	if err := nodes.Deserialize(bytes.NewBuffer(nodesStore)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize sipMap error: %v", err)
	}
	return nodes, nil
}

func putDefConsNodes(native *native.NativeService, contract common.Address, nodes *DefaultConsNodes) error {
	bf := new(bytes.Buffer)
	if err := nodes.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize default consensus nodes error: %v", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(DEFAULT_CONS_NODE)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

//query pdp winner from FS
func updateMinerPowerMap(native *native.NativeService, contract common.Address, height uint32) ([]common.Address, error) {
	miningView, err := GetMiningView(native, contract)
	if err != nil {
		return []common.Address{}, fmt.Errorf("SettleView, get view error: %v", err)
	}

	minerPowerMap, err := GetMinerPowerMap(native, contract)
	if err != nil {
		return []common.Address{}, fmt.Errorf("updateMinerPowerMap, get minerPowerMap error: %v", err)
	}

	//increate pdp in new view
	view := miningView.View + 1
	first := native.Height - NUM_BLOCK_PER_VIEW + 1
	if first <= 0 {
		first = 1
	}
	last := native.Height

	for curHeight := uint32(first); curHeight < last; curHeight++ {
		sink := common.ZeroCopySink{}
		sink.WriteUint32(curHeight)

		data, err := native.NativeCall(utils.OntFSContractAddress, "FsGetPocProveList", sink.Bytes())
		if err != nil {
			return []common.Address{}, fmt.Errorf("updateMinerPowerMap, call FsGetPocProveList error: %v", err)
		}

		retInfo := fs.DecRet(data)
		if retInfo.Ret {
			prove := &fs.PocProveList{}
			source := common.NewZeroCopySource(retInfo.Info)
			err = prove.Deserialization(source)

			if err != nil {
				return []common.Address{}, fmt.Errorf("updateMinerPowerMap error: %s", err.Error())
			}

			log.Debugf("updateMinerPowerMap for height: %d, get: pdp record for %d miner\n", curHeight, len(prove.Proves))
			for _, proof := range prove.Proves {
				log.Debugf("updateMinerPowerMap for height: %d, miner %s get: %d power\n", curHeight, proof.Miner.ToBase58(), proof.PlotSize)
				if _, ok := minerPowerMap.MinerPowerMap[proof.Miner]; !ok {
					minerPowerMap.MinerPowerMap[proof.Miner] = &MinerPowerItem{Address: proof.Miner}
				}
				minerPowerMap.MinerPowerMap[proof.Miner].Power += proof.PlotSize
			}
		} else {
			return []common.Address{}, errors.New(string(retInfo.Info))
		}
	}

	//[TODO] Need use same period with sector proof period for plot file
	//reduce pdp in expired view
	if int64(view)-NUM_VIEW_PER_DAY <= 0 {
		first = native.Height
		last = native.Height
	} else {
		last = native.Height - view*NUM_VIEW_PER_DAY*NUM_BLOCK_PER_VIEW
		first = last - NUM_BLOCK_PER_VIEW + 1
	}

	for curHeight := uint32(first); curHeight < last; curHeight++ {
		sink := common.ZeroCopySink{}
		sink.WriteUint32(curHeight)

		data, err := native.NativeCall(utils.OntFSContractAddress, "FsGetPocProveList", sink.Bytes())
		if err != nil {
			return []common.Address{}, fmt.Errorf("updateMinerPowerMap, call FsGetPocProveList error: %v", err)
		}

		retInfo := fs.DecRet(data)
		if retInfo.Ret {
			prove := &fs.PocProveList{}
			source := common.NewZeroCopySource(retInfo.Info)
			err = prove.Deserialization(source)

			if err != nil {
				return []common.Address{}, fmt.Errorf("updateMinerPowerMap error: %s", err.Error())
			}

			log.Debugf("updateMinerPowerMap for height: %d, get: pdp record for %d miner\n", curHeight, len(prove.Proves))
			for _, proof := range prove.Proves {
				log.Debugf("updateMinerPowerMap for height: %d, miner %s get: %d power\n", curHeight, proof.Miner.ToBase58(), proof.PlotSize)
				if _, ok := minerPowerMap.MinerPowerMap[proof.Miner]; !ok {
					continue
				}

				if minerPowerMap.MinerPowerMap[proof.Miner].Power <= proof.PlotSize {
					delete(minerPowerMap.MinerPowerMap, proof.Miner)
				} else {
					minerPowerMap.MinerPowerMap[proof.Miner].Power -= proof.PlotSize
				}
			}
		} else {
			return []common.Address{}, errors.New(string(retInfo.Info))
		}
	}

	winners := []MinerPowerItem{}
	for _, miner := range minerPowerMap.MinerPowerMap {
		winner := MinerPowerItem{
			Address: miner.Address,
			Power:   miner.Power,
		}
		winners = append(winners, winner)
	}

	// sort winner by power
	sort.SliceStable(winners, func(i, j int) bool {
		if winners[i].Power > winners[j].Power {
			return true
		} else {
			return winners[i].Address.ToBase58() > winners[j].Address.ToBase58()
		}
		return false
	})

	winnerAddress := []common.Address{}
	for _, winner := range winners {
		address := winner.Address
		winnerAddress = append(winnerAddress, address)
		log.Debugf("updateMinerPowerMap for view: %d, miner %s get: total %d power\n", view, winner.Address, winner.Power)
	}

	err = putMinerPowerMap(native, contract, minerPowerMap)
	if err != nil {
		return []common.Address{}, fmt.Errorf("putConsVoteMap error: %v", err)
	}

	return winnerAddress, nil
}

func GetMinerPowerMap(native *native.NativeService, contract common.Address) (*MinerPowerMap, error) {
	minerPowerMap := &MinerPowerMap{
		MinerPowerMap: make(map[common.Address]*MinerPowerItem),
	}

	minerPowerMapBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(MINER_POWER_MAP)))
	if err != nil {
		return nil, fmt.Errorf("GetMinerPowerMap, get minerPowerMap error: %v", err)
	}
	if minerPowerMapBytes == nil {
		return nil, fmt.Errorf("GetMinerPowerMap, minerPowerMap is nil")
	}
	item := cstates.StorageItem{}
	err = item.Deserialization(common.NewZeroCopySource(minerPowerMapBytes))
	if err != nil {
		return nil, fmt.Errorf("deserialize minerPowerMap error:%v", err)
	}

	if err := minerPowerMap.Deserialize(bytes.NewBuffer(item.Value)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize minerPowerMap error: %v", err)
	}
	return minerPowerMap, nil
}

func putMinerPowerMap(native *native.NativeService, contract common.Address, minerPowerMap *MinerPowerMap) error {
	bf := new(bytes.Buffer)
	if err := minerPowerMap.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize minerPowerMap error: %v", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(MINER_POWER_MAP)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}
