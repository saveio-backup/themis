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
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

const (
	//function name
	INIT_POC_CONFIG   = "initPoCConfig"
	QUERY_MINING_INFO = "queryMiningInfo"
	SETTLE_VIEW       = "settleView"
	QUERY_WINNER_INFO = "queryWinnerInfo"

	//key prefix
	MINING_VIEW                       = "miningView"
	MINE_VIEW_INFO_KEY_PATTERN        = "miningViewInfo=%d"
	WINNER_INFO_KEY_PATTERN           = "winnerView=%d"
	MINER_DEADLINE_PERIOD_KEY_PATTERN = "minerDeadlinePeriod=%d"
	PERIOD_INFOS                      = "periodInfos"
	PERIOD_SUMMARY_KEY_PATTERN        = "periodSummary=%d"
	CONS_VOTE_INFO                    = "consVoteInfo"
	CONS_VOTE_DETAIL                  = "consVoteDetail"
	CONS_GROUP_INFO                   = "consGroup"
	CONS_VOTE_REVENUE                 = "consVoteRevenue"
	DEFAULT_CONS_NODE                 = "defaultConsNodes"

	DIFF_ADJUST_CHANGE_BLOCK = 2700

	HALVING_INTERVAL = 200000
	BONUS_BASE       = 250
	COIN             = 1000000000

	NUM_BLOCK_PER_VIEW = 120

	NUM_VIEW_PER_PERIOD        = 4032
	NUM_VIEW_PER_VERIFY_PERIOD = 3024
	NUM_VIEW_PER_PLEDGE_PERIOD = 1008

	NUM_VIEW_PER_CONS_GOV_PERIOD    = 8064
	NUM_VIEW_PER_RUNNING_PERIOD     = 4032
	NUM_VIEW_PER_CONS_PLEDGE_PERIOD = 2016
	NUM_VIEW_PER_CONS_ELECT_PERIOD  = 1008
	NUM_VIEW_PER_GOV_ELECT_PERIOD   = 1008
	NUM_VIEW_COLD_DOWN              = 36

	NUM_NODE_PER_CONS_ELECT_PERIOD = 3

	NUM_VIEW_PER_YEAR = 52560
	NUM_VIEW_PER_DAY  = 144
	NUM_DAY_DELAYED   = 90

	DELAYED_PERCENT = 70

	FS_PLOT_EXPECTED_PERCENT = 10
)

func InitPoC() {
}

func RegisterPoCContract(native *native.NativeService) {
	native.Register(INIT_POC_CONFIG, InitPoCConfig)
	native.Register(SETTLE_VIEW, SettleView)
	native.Register(QUERY_MINING_INFO, QueryMiningInfo)
	native.Register(QUERY_WINNER_INFO, QueryWinnerInfo)
}

//Init poc mining contract.
func InitPoCConfig(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	// check if initConfig is already execute
	miningViewBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(MINING_VIEW)))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitPoCConfig, get miningViewBytes error: %v", err)
	}
	if miningViewBytes != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitPoCConfig. InitPoCConfig is already executed")
	}

	//init poc mining view. MiningView will change every 120 block
	miningView := &MiningView{
		View:   0,
		Height: 0,
	}
	err = putMiningView(native, contract, miningView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitPoCConfig, put miningView error: %v", err)
	}

	// init generation mining info of genesis block
	miningViewInfo := &MiningViewInfo{
		GenerationSignature: common.Uint256{},
		Generator:           0,
		BaseTarget:          1,
	}
	newGenerationSignature, err := calGenerationSignature(miningViewInfo.GenerationSignature, miningViewInfo.Generator)
	miningViewInfo.NewGenerationSignature = newGenerationSignature
	scoop := calculateScoop(uint64(miningView.View+1), newGenerationSignature.ToArray())
	miningViewInfo.Scoop = scoop

	err = putMiningViewInfo(native, contract, miningView.View, miningViewInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitPoCConfig, put miningViewInfo error: %v", err)
	}

	// init winner info of genesis block
	winnerInfo := &WinnerInfo{}

	err = putWinnerInfo(native, contract, miningView.View, winnerInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitPoCConfig, put miningViewInfo error: %v", err)
	}
	log.Debugf("PoCInit put winner info for view: %d", miningView.View)

	//init consensus vote bonus
	consVoteRevenue := &ConsVoteRevenue{}
	err = putConsVoteRevenue(native, contract, consVoteRevenue)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitSIPConfig, put consVoteRevenue error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func QueryMiningInfo(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	miningView, err := GetMiningView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetMiningView, get view error: %v", err)
	}

	miningViewInfo, err := getMiningViewInfo(native, contract, miningView.View)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getMiningViewInfo, get mining view info error: %v", err)
	}

	miningInfo := &MiningInfo{
		View:                miningView.View + 1,
		BaseTarget:          miningViewInfo.BaseTarget,
		GenerationSignature: miningViewInfo.NewGenerationSignature,
	}

	buf := new(bytes.Buffer)
	if err = miningInfo.Serialize(buf); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QueryMiningInfo miningInfo Serialize error:%v", err)
	}
	return buf.Bytes(), nil
}

func QueryWinnerInfo(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	req := new(WinnerInfoReq)
	if err := req.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QueryWinnerInfo  deserialization view error: %v", err)
	}

	winnerInfo, err := getWinnerInfo(native, contract, req.View)

	buf := new(bytes.Buffer)
	if err = winnerInfo.Serialize(buf); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QueryWinnerInfo winnerInfo Serialize error:%v", err)
	}
	return buf.Bytes(), nil
}

//Go to next PoC mining view. Adjust target etc
func UpdateTarget(native *native.NativeService, submitInfo *SubmitNonceParam) ([]byte, error) {

	view := submitInfo.View
	preView := view - 1
	contract := native.ContextRef.CurrentContext().ContractAddress
	miningView, err := GetMiningView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, get view error: %v", err)
	}
	if miningView.View != preView {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, miningView: %d, expected view: %d", miningView.View, preView)
	}

	preMiningViewInfo, err := getMiningViewInfo(native, contract, miningView.View)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, get mining info error: %v", err)
	}

	//store new miningViewInfo and winnerInfo
	miningViewInfo := &MiningViewInfo{
		GenerationSignature: preMiningViewInfo.NewGenerationSignature,
		Generator:           uint64(submitInfo.Id),
	}

	winnerInfo := &WinnerInfo{
		View:        view,
		Address:     submitInfo.Address,
		Deadline:    submitInfo.Deadline,
		VoteConsPub: submitInfo.VoteConsPub,
		VoteId:      submitInfo.VoteId,
		VoteInfo:    submitInfo.VoteInfo,
	}

	log.Debugf("[UpdateTarget] dump winnerInfo %v\n", winnerInfo)

	err = putWinnerInfo(native, contract, view, winnerInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, put miningViewInfo error: %v", err)
	}

	miningViewInfo.BaseTarget = 1
	generationSignature, err := calGenerationSignature(miningViewInfo.GenerationSignature, miningViewInfo.Generator)
	miningViewInfo.NewGenerationSignature = generationSignature
	scoop := calculateScoop(uint64(view+1), generationSignature.ToArray())
	miningViewInfo.Scoop = scoop

	err = putMiningViewInfo(native, contract, view, miningViewInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, put miningViewInfo error: %v", err)
	}

	//update miningView
	miningView.View = view
	miningView.Height = native.Height

	err = putMiningView(native, contract, miningView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, put miningView error: %v", err)
	}

	// handle sip vote info after winner info recorded
	err = handleSipVote(native, submitInfo.Address, submitInfo.VoteId, submitInfo.VoteInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, handleSipVote error: %v", err)
	}
	err = splitSipBonus(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, splitSipBonus error: %v", err)
	}
	err = triggerSipAction(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, triggerSipAction error: %v", err)
	}

	//detect moving up consensus election
	if submitInfo.MoveUpElect {
		err = handleConsElectMoveUp(native, uint32(view))
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, handleConsElectMoveUp error: %v", err)
		}
	}

	//notify nodes to pledge for consensus election
	err = notifyConsPledge(native, uint32(view))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, notifyConsPledge error: %v", err)
	}

	//notify miner vote for cons node
	err = notifyConsVote(native, uint32(view))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, notifyConsVote error: %v", err)
	}
	//handle cons vote info after winner info recorded
	err = handleConsVote(native, uint32(view), submitInfo.Address, submitInfo.VoteConsPub)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, handleConsVote error: %v", err)
	}
	//try to elect consensus group for next consensus gov view
	err = electConsGroup(native, uint32(view))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, electConsGroup error: %v", err)
	}

	// update mining period info
	err = updatePeriod(native, uint32(view))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, updatePeriod error: %v", err)
	}

	// transfer delayed bonus
	err = transferDelayedBonus(native, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateTarget, transferDelayedBonus error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//record deadline for each miner,cal avg deadline for winner
func handleDeadline(native *native.NativeService, param *SubmitNonceParam, winner common.Address) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	//update period avg deadline and estimate plot of winner based on win times
	miningView, err := GetMiningView(native, contract)
	if err != nil {
		return fmt.Errorf("handleDeadline, get view error: %v", err)
	}
	view := miningView.View + 1

	recordPeriodSummary(native, param)

	//update winner period info with avg deadline
	if PledgeViewInMiningPeriod(view) {
		summaryPeriod(native, view)
		triggerPledgeForPeriod(native, view)
	}
	return nil
}

//Go to next PoC mining view. Adjust target etc
func SplitBonus(native *native.NativeService, winner common.Address, view uint32, bonus uint64, delayed bool) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	//percent, err := checkPeriodsInfo(native, winner)
	//if err != nil {
	//	return fmt.Errorf("SettleView, checkPeriodsInfo %s", err)
	//}

	// transfer bonus based on globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("SplitBonus, getGlobalParam error: %v", err)
	}

	votesBonus := bonus * uint64(globalParam.Votes) / 100

	//calculate fundBonus and pocBonus
	year := (view - 1) / NUM_VIEW_PER_YEAR
	pocPercent := 80 + year*5
	if pocPercent >= 100 {
		pocPercent = 95
	}
	fundPercent := 100 - pocPercent

	//split bonus to fundation
	fundBonus := bonus * uint64(fundPercent) / 100
	realFundBonus := GetEffectBonus(fundBonus, delayed)
	if delayed {
		total := GetTotalDelayedBonus(fundBonus)
		log.Debugf("SplitBonus send delayed %d(10^-9) of total %d(10^-9) of view %d to fundation", realFundBonus, total, view)
	} else {
		log.Debugf("SplitBonus send non-delay %d(10^-9) of total %d(10^-9) of view %d to fundation", realFundBonus, fundBonus, view)
	}
	SplitBonusToFundation(native, contract, realFundBonus, delayed)

	//split bonus to miner
	pocBonus := bonus * uint64(pocPercent) / 100
	winnerBonus := pocBonus * uint64(globalParam.PoC) / 100
	realWinnerBonus := GetEffectBonus(winnerBonus, delayed)
	if delayed {
		total := GetTotalDelayedBonus(winnerBonus)
		log.Debugf("SplitBonus send delayed %d(10^-9) of total %d(10^-9) of view %d to miner", realWinnerBonus, total, view)
	} else {
		log.Debugf("SplitBonus send non-delay %d(10^-9) of total %d(10^-9) of view %d to miner", realWinnerBonus, winnerBonus, view)
	}
	SplitBonusToMiner(native, winner, realWinnerBonus, delayed)

	//split bonus to cons nodes
	consBonus := pocBonus * uint64(globalParam.Cons) / 100
	realConsBonus := GetEffectBonus(consBonus, delayed)
	if delayed {
		total := GetTotalDelayedBonus(consBonus)
		log.Debugf("SplitBonus send delayed %d(10^-9) of total %d(10^-9) of view %d to cons nodes", realConsBonus, total, view)
	} else {
		log.Debugf("SplitBonus send non-delay %d(10^-9) of total %d(10^-9) of view %d to cons nodes", realConsBonus, consBonus, view)
	}
	SplitBonusToConsensusNodes(native, contract, view, realConsBonus, delayed)

	//sip and cons vote share vote bonus
	if !delayed {
		increaseSipVoteRevenue(native, contract, votesBonus/2)
		increaseConsVoteRevenue(native, contract, votesBonus/2)
		log.Debugf("SplitBonus  %d(10^-9) for consensus, %d(10^-9) for votes", consBonus, votesBonus)
	}

	return nil
}

func GetTotalDelayedBonus(total uint64) uint64 {
	return total * DELAYED_PERCENT / 100
}

func GetEffectBonus(total uint64, delayed bool) uint64 {
	if delayed {
		total = GetTotalDelayedBonus(total)
		total /= NUM_DAY_DELAYED
	} else {
		total = total * (100 - DELAYED_PERCENT) / 100
	}
	return total
}

//[TODO] split bonus to multiple miners based pdp rate,need call FS contract
func SplitBonusToMiner(native *native.NativeService, winner common.Address, bonus uint64, delayed bool) error {
	pocBonus := bonus

	if delayed {
		log.Debugf("SplitBonusToMiner send delayed %d(10^-9) for miner:%s", pocBonus, winner.ToBase58())
	} else {
		log.Debugf("SplitBonusToMiner send non-delay %d(10^-9) for miner:%s", pocBonus, winner.ToBase58())
	}

	err := appCallTransferOnt(native, utils.GovernanceContractAddress, winner, uint64(pocBonus))
	if err != nil {
		return fmt.Errorf("SplitBonusToMiner, bonus transfer error: %v", err)
	}

	return nil
}

func SplitBonusToFundation(native *native.NativeService, contract common.Address, bonus uint64, delayed bool) error {
	// transfer bonus based on globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("SplitBonusToFundation, getGlobalParam error: %v", err)
	}

	fundAddr, err := common.AddressFromBase58(globalParam.FundWalletAddr)

	if delayed {
		log.Debugf("SplitBonusToFundation send delayed %d(10^-9) for fundation:%s", bonus, globalParam.FundWalletAddr)
	} else {
		log.Debugf("SplitBonusToFundation send non-delay %d(10^-9) for fundation:%s", bonus, globalParam.FundWalletAddr)
	}
	err = appCallTransferOnt(native, utils.GovernanceContractAddress, fundAddr, uint64(bonus))
	if err != nil {
		return fmt.Errorf("SplitBonusToFundation, bonus transfer error: %v", err)
	}
	return nil
}

func SplitBonusToConsensusNodes(native *native.NativeService, contract common.Address, view uint32, bonus uint64, delayed bool) error {
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getConfig, get config error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view-1)
	if err != nil {
		return fmt.Errorf("SplitBonusToConsensusNodes, get peerPoolMap error: %v", err)
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}
	balance := bonus

	peersCandidate := []*CandidateSplitInfo{}

	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peersCandidate = append(peersCandidate, &CandidateSplitInfo{
				PeerPubkey: peerPoolItem.PeerPubkey,
				InitPos:    peerPoolItem.InitPos,
				Address:    peerPoolItem.Address,
				Stake:      stake,
			})
		}
	}

	// sort peers by stake
	sort.SliceStable(peersCandidate, func(i, j int) bool {
		if peersCandidate[i].Stake > peersCandidate[j].Stake {
			return true
		} else if peersCandidate[i].Stake == peersCandidate[j].Stake {
			return peersCandidate[i].PeerPubkey > peersCandidate[j].PeerPubkey
		}
		return false
	})

	// cal s of each consensus node
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	// if sum = 0, means consensus peer in config, do not split
	if sum < uint64(config.K) {
		//use same share for peer in config
		sum = uint64(config.K)
	}

	avg := sum / uint64(config.K)
	var sumS uint64
	for i := 0; i < int(config.K); i++ {
		//use same share when sum is 0
		if sum == 0 {
			peersCandidate[i].S, err = splitCurve(native, contract, 1, avg, uint64(globalParam.Yita))
			if err != nil {
				return fmt.Errorf("splitCurve, calculate splitCurve error: %v", err)
			}
		} else {
			peersCandidate[i].S, err = splitCurve(native, contract, peersCandidate[i].Stake, avg, uint64(globalParam.Yita))
			if err != nil {
				return fmt.Errorf("splitCurve, calculate splitCurve error: %v", err)
			}
		}
		sumS += peersCandidate[i].S
	}
	if sumS == 0 {
		return fmt.Errorf("SplitBonusToConsensusNodes, sumS is 0")
	}

	//fee split of consensus peer
	for i := 0; i < int(config.K); i++ {
		nodeAmount := balance * uint64(globalParam.A2) / 100 * peersCandidate[i].S / sumS
		address := peersCandidate[i].Address

		if delayed {
			log.Debugf("SplitBonusToConsensusNodes send delayed %d(10^-9) bonus for cons node:%s", nodeAmount, address.ToBase58())
		} else {
			log.Debugf("SplitBonusToConsensusNodes send non-delay %d(10^-9) bonus for cons node:%s", nodeAmount, address.ToBase58())
		}

		err = appCallTransferOnt(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return fmt.Errorf("SplitBonusToConsensusNodes, bonus transfer error: %v", err)
		}
	}

	//fee split of candidate peer
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	if sum == 0 {
		return nil
	}
	for i := int(config.K); i < len(peersCandidate); i++ {
		nodeAmount := balance * uint64(globalParam.B2) / 100 * peersCandidate[i].Stake / sum
		address := peersCandidate[i].Address

		if delayed {
			log.Debugf("SplitBonusToConsensusNodes send delayed %d(10^-9) bonus for cons node:%s", nodeAmount, address.ToBase58())
		} else {
			log.Debugf("SplitBonusToConsensusNodes send non-delay %d(10^-9) bonus for cons node:%s", nodeAmount, address.ToBase58())
		}

		err = appCallTransferOnt(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return fmt.Errorf("SplitBonusToConsensusNodes, bonus transfer error: %v", err)
		}
	}

	return nil
}

func SettleView(native *native.NativeService) ([]byte, error) {
	var submitInfo *SubmitNonceParam

	contract := native.ContextRef.CurrentContext().ContractAddress
	log.Debugf("[SettleView] Input  %v", native.Input)

	submitInfo = new(SubmitNonceParam)
	if err := submitInfo.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[SettleView] deserialization param error: %v", err)
	}

	miningView, err := GetMiningView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SettleView, get view error: %v", err)
	}

	if submitInfo.View != miningView.View+1 {
		return utils.BYTE_FALSE, fmt.Errorf("SettleView, submit for view: %d, expected %d", submitInfo.View, miningView.View+1)
	}

	if native.Height-miningView.Height < NUM_BLOCK_PER_VIEW {
		return utils.BYTE_FALSE, fmt.Errorf("SettleView, block number error: %d", native.Height)

	}

	// verify nonce
	// id 0 mean dummy param!
	if submitInfo.Id != 0 {
		miningViewInfo, err := getMiningViewInfo(native, contract, miningView.View)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[SettleView], get mining view info error: %v", err)
		}
		genSig := miningViewInfo.NewGenerationSignature.ToArray()
		valid := verifyNonce(submitInfo, miningViewInfo.Scoop, miningViewInfo.BaseTarget, genSig)
		if !valid {
			return utils.BYTE_FALSE, fmt.Errorf("[SettleView], submitted nonce from id doesn't match expected value")
		}
	} else {
		submitInfo.Deadline = math.MaxUint64
	}

	log.Debugf("SettleView for view: %d, id: %d, nonce: %d, deadline: %d\n", submitInfo.View, submitInfo.Id, submitInfo.Nonce, submitInfo.Deadline)

	plotSize := CalPlotFileSize(submitInfo.Deadline)
	log.Debugf("SettleView: plot size calculated for deadline %d is %d MB", submitInfo.Deadline, plotSize)

	//record deadline for each miner for each period
	err = handleDeadline(native, submitInfo, submitInfo.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[SettleView], handleDeadline error: %v", err)
	}

	// send bonus to miner
	bonus := GetBlockSubsidy(miningView.View)
	err = SplitBonus(native, submitInfo.Address, submitInfo.View, bonus, false)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SettleView, SplitBonus fail %s", err)
	}

	// update view state, handle vote and prepare parameter for next view
	_, err = UpdateTarget(native, submitInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SettleView, call UpdateTarget error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//record avg deadlines found in one period and win time for every winner
func recordPeriodSummary(native *native.NativeService, winner *SubmitNonceParam) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	miningView, err := GetMiningView(native, contract)
	if err != nil {
		return fmt.Errorf("recordPeriodSummary, get view error: %v", err)
	}

	view := miningView.View + 1
	period := (view + NUM_VIEW_PER_PERIOD - 1) / NUM_VIEW_PER_PERIOD

	info, err := getPeriodSummary(native, period)
	if err != nil {
		return fmt.Errorf("recordPeriodSummary get period summary info error: %v", err)
	}

	viewInPeriod := view - (period-1)*NUM_VIEW_PER_PERIOD
	avgDeadline := big.NewInt(1)
	avgDeadline.SetUint64(info.AvgDeadline)
	sumDeadline := big.NewInt(1)
	sumDeadline = sumDeadline.Mul(avgDeadline, big.NewInt(1).SetUint64(uint64(viewInPeriod-1)))
	winnerDeadline := big.NewInt(1)
	winnerDeadline.SetUint64(winner.Deadline)
	sumDeadline = sumDeadline.Add(sumDeadline, winnerDeadline)
	newAvg := sumDeadline.Div(sumDeadline, big.NewInt(1).SetUint64(uint64(viewInPeriod)))

	info.AvgDeadline = newAvg.Uint64()
	info.MinerWinTimes[winner.Address]++

	err = putPeriodSummary(native, period, info)
	if err != nil {
		return fmt.Errorf("recordPeriodSummary put period summary info error: %v", err)
	}

	return nil
}

// estimate plot for winner during period!
func summaryPeriod(native *native.NativeService, view uint32) error {
	period := GetMiningPeriod(view)

	summary, err := getPeriodSummary(native, period)
	if err != nil {
		return fmt.Errorf("summaryPeriod get period summary  error: %v", err)
	}

	totalPlotSize := CalPlotFileSize(summary.AvgDeadline)
	viewInPeriod := view % NUM_VIEW_PER_PERIOD
	if viewInPeriod == 0 {
		viewInPeriod = NUM_VIEW_PER_PERIOD
	}

	log.Debugf("summaryPeriod calculate total plot size %d(MB)", totalPlotSize)
	for addr, times := range summary.MinerWinTimes {
		info, err := getPeriodInfos(native, addr)
		if err != nil {
			return fmt.Errorf("summaryPeriod error: %v", err)
		}

		plotSize := totalPlotSize * uint64(times) / uint64(viewInPeriod)
		info.curPeriod.Period = period
		info.curPeriod.PlotSize = plotSize

		log.Debugf("summaryPeriod calculate plot size %d(MB) for miner %s", plotSize, addr.ToBase58())

		err = putPeriodsInfo(native, addr, info)
		if err != nil {
			return fmt.Errorf("summaryPeriod error: %v", err)
		}
	}

	return nil
}

func checkPeriodsInfo(native *native.NativeService, address common.Address) (uint64, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	miningView, err := GetMiningView(native, contract)
	if err != nil {
		return 0, fmt.Errorf("checkPeriodsInfo, get view error: %v", err)
	}

	view := miningView.View + 1
	period := (view + NUM_VIEW_PER_PERIOD - 1) / NUM_VIEW_PER_PERIOD
	needCheckPlot := period > 1 && (view-(period-1)*NUM_VIEW_PER_PERIOD) <= NUM_VIEW_PER_VERIFY_PERIOD

	if needCheckPlot {
		vol, err := queryVolume(native, address)
		if err != nil {
			//make settle continue
			return 0, nil
		}
		log.Debugf("checkPeriodsInfo: miner %s has volume %d MB", address.ToBase58(), vol)
		//ensure fs space reach plotSize * 10%, using pre period info!!
		info, err := getPeriodInfos(native, address)
		if err != nil {
			return 0, fmt.Errorf("checkPeriodsInfo error: %v", err)
		}

		expectSize := info.prePeriod.PlotSize
		log.Debugf("checkPeriodsInfo: plot size calculated for pre period is %d MB", expectSize)

		//FS vol is k, adjust to M
		realSize := vol / 1024
		fullBonusSize := expectSize * FS_PLOT_EXPECTED_PERCENT / 100
		if realSize >= fullBonusSize {
			return 100, nil
		} else {
			return realSize * 100 / fullBonusSize, nil
		}
	}
	return 100, nil
}
