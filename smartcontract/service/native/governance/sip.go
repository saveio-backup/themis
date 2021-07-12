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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	cstates "github.com/saveio/themis/core/states"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/global_params"
	"github.com/saveio/themis/smartcontract/service/native/utils"
	"github.com/saveio/themis/vm/wasmvm/util"
)

const (
	//function name
	INIT_SIP_CONFIG = "initSIPConfig"
	REGISTER_SIP    = "registerSIP"
	QUERY_SIP       = "querySIP"

	//key prefix
	SIP_POOL               = "sipPool"
	SIP_SEQ_INDEX          = "sipSeqIndex"
	SIP_INDEX              = "sipIndex"
	SIP_LAST_CHANGE_HEIGHT = "sipLastChangeHeight"
	SIP_VOTE_REVENUE       = "sipVoteRevenue"

	SIP_VOTE_DELAY         = 120960
	SIP_VOTE_PERIOD        = 120960
	MAX_VOTES              = 1008
	SIP_PARAM_CHANGE_DELAY = 120960

	//Sip vote result
	AGREE = byte(1)
	EXEC  = byte(2)

	//Sip parameter name
	SIP_MIN_INIT_STAKE        = "MinInitStake"
	SIP_CONS_BONUS_SPLIT_RATE = "ConsBonusSplitRate"
	SIP_POC_SPLIT_RATE        = "PoCSplitRate"
	SIP_PDP_GAS               = "PDPGas"
)

type SipParamAttr struct {
	NumValue  int
	NeedCheck bool
}

var formatMap map[string]*SipParamAttr

func RegisterSIPContract(native *native.NativeService) {
	native.Register(INIT_SIP_CONFIG, InitSIPConfig)
	native.Register(REGISTER_SIP, RegisterSIP)
	native.Register(QUERY_SIP, QuerySIP)
}

//Init sip
func InitSIPConfig(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	sipMap := &SipMap{
		SipMap: make(map[uint32]*SIP),
	}
	err := putSipMap(native, contract, sipMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitSIPConfig,  putSipMap error: %v", err)
	}

	firstSipIndex := 1
	sipBytes := GetUint32Bytes(uint32(firstSipIndex))

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIP_SEQ_INDEX)), cstates.GenRawStorageItem(sipBytes))

	// initialize vote revenue
	sipVoteRevenue := &SIPVoteRevenue{}
	err = putSipVoteRevenue(native, contract, sipVoteRevenue)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitSIPConfig, put sipVoteRevenue error: %v", err)
	}

	initSipParamHeight(native)

	return utils.BYTE_TRUE, nil
}

//Register new SIP. Need admin
func RegisterSIP(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	param := new(RegisterSipParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}

	//verify param
	if param.Height < native.Height+SIP_VOTE_DELAY+SIP_VOTE_PERIOD {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, too late to register sip with effective height %d", param.Height)
	}
	if param.MinVotes > MAX_VOTES {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, MinVotes %d exceed max votes number %d", param.MinVotes, MAX_VOTES)
	}

	//check detail format
	paramName, _, err := VerifySIPDetail(param.Detail)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, check sip detail format error: %v", err)
	}
	height, err := getParamChangeHeight(native, contract, paramName)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, check sip detail format error: %v", err)
	}
	if native.Height < height+SIP_PARAM_CHANGE_DELAY {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, too frequent to change %s", paramName)
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}
	consGovView, err := GetConsGovView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, get cons gov view error: %v", err)
	}

	if IsDuringGovElect(consGovView, view) {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, check during gov elect error: %v", err)
	}

	//check bonus
	if param.Bonus > 0 {
		err := reserveSipVoteRevenue(native, contract, param.Bonus)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, check bonus fail: %v", err)
		}
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, checkWitness error: %v", err)
	}

	//serialize here
	bf := new(bytes.Buffer)
	err = param.Serialize(bf)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, serialize sip param err %v", err)
	}
	hash := sha256.New()
	hash.Write(bf.Bytes())
	sipDigestBytes := hash.Sum(nil)
	sipDigest := hex.EncodeToString(sipDigestBytes)

	//check if sip exist
	indexBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIP_INDEX), []byte(sipDigest)))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, get indexBytes error: %v", err)
	}
	if indexBytes != nil {
		value, err := cstates.GetValueFromRawStorageItem(indexBytes)
		if err != nil {
			return nil, fmt.Errorf("get value from raw storage item error:%v", err)
		}
		index, err := GetBytesUint32(value)
		if err != nil {
			return nil, fmt.Errorf("GetBytesUint32, get index error: %v", err)
		}
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, sip with digest %s, index %d already exist", sipDigest, index)

	}

	//get sipMap
	sipMap, err := GetSipMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSIP, get sip Map error: %v", err)
	}

	//generate sip index
	sipIndex, err := getSipIndex(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getSipIndex, get sip index error: %v", err)
	}

	log.Debugf("RegisterSIP,  new SIP registered with index <%d>", sipIndex)

	sip := &SIP{
		Digest:    sipDigest,
		Index:     sipIndex,
		Height:    param.Height,
		Detail:    param.Detail,
		Default:   param.Default,
		MinVotes:  param.MinVotes,
		Bonus:     param.Bonus,
		RegHeight: native.Height,
	}

	sipMap.SipMap[sip.Index] = sip
	err = putSipMap(native, contract, sipMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putSipMap error: %v", err)
	}

	//record sip digest
	sipBytes := GetUint32Bytes(uint32(sipIndex))

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIP_INDEX), []byte(sipDigest)), cstates.GenRawStorageItem(sipBytes))

	//update sip Index
	newsipIndex := sipIndex + 1
	err = putSipIndex(native, contract, newsipIndex)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("put sipIndex error: %v", err)
	}

	//EventNotify
	sipRegister := &sipRegisterEvent{
		Index:   sip.Index,
		Height:  sip.Height,
		Detail:  sip.Detail,
		Default: sip.Default,
	}

	SipRegisteredEvent(native, sipRegister)

	return util.Int32ToBytes(sip.Index), nil
}

func QuerySIP(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	param := new(QuerySipParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}

	sipMap, err := GetSipMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuerySIP, get sip Map error: %v", err)
	}

	sip, ok := sipMap.SipMap[param.Index]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("QuerySIP, SIP with index %d not exist", param.Index)
	}

	info := new(bytes.Buffer)
	if err = sip.Serialize(info); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuerySIP sip serialize error:%v", err)
	}

	return info.Bytes(), nil
}

func initSipParamHeight(native *native.NativeService) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	if formatMap == nil {
		InitSipParamAttrs()
	}

	for param, _ := range formatMap {
		putParamChangeHeight(native, contract, param, 0)
	}

}

//update vote with info in winner info
func handleSipVote(native *native.NativeService, voter common.Address, sipIndex []uint32, voteInfo []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	sipMap, err := GetSipMap(native, contract)
	if err != nil {
		return fmt.Errorf("handleVote, get sip Map error: %v", err)
	}

	for i := 0; i < len(sipIndex); i++ {
		sip, ok := sipMap.SipMap[sipIndex[i]]
		if !ok {
			log.Debugf("handleVote, fail to find sip with index %d", sipIndex[i])
			continue
		}

		// only accept vote during vote period dealy
		if native.Height < sip.RegHeight+SIP_VOTE_DELAY {
			continue
		}

		// exceed vote period
		if native.Height > sip.RegHeight+SIP_VOTE_DELAY+SIP_VOTE_PERIOD {
			continue
		}

		if voteInfo[i] == AGREE {
			sip.NumVotes++
			if num, ok := sip.VoterMap[voter]; ok {
				num++
				sip.VoterMap[voter] = num
				log.Debugf("handleVote,  SIP<%d> see %s before, votes %d/%d", sipIndex[i], voter.ToBase58(), num, sip.VoterMap[voter])
			} else {
				sip.VoterMap[voter] = 1
				log.Debugf("handleVote,  SIP<%d> see %s first time, votes %d", sipIndex[i], voter.ToBase58(), sip.VoterMap[voter])
			}
		}

		log.Debugf("handleVote,  SIP<%d> get %d decision from %s, total votes %d", sipIndex[i], voteInfo[i], voter.ToBase58(), sip.NumVotes)
		for address, v := range sip.VoterMap {
			log.Debugf("handleVote,  SIP<%d> get %d votes from address %s", sipIndex[i], v, address.ToBase58())
		}

		if sip.NumVotes > sip.MinVotes {
			sip.Result = AGREE
		}
	}

	err = putSipMap(native, contract, sipMap)
	if err != nil {
		return fmt.Errorf("putSipMap error: %v", err)
	}

	return nil
}

func splitSipBonus(native *native.NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	sipMap, err := GetSipMap(native, contract)
	if err != nil {
		return fmt.Errorf("handleVote, get sip Map error: %v", err)
	}

	voteRevenue, err := getSipVoteRevenue(native, contract)
	if err != nil {
		return fmt.Errorf("splitSipBonus, get sip vote revenue error: %v", err)
	}

	for index, sip := range sipMap.SipMap {

		// still during voting
		if native.Height < sip.RegHeight+SIP_VOTE_DELAY+SIP_VOTE_PERIOD {
			continue
		}

		if sip.Bonus == 0 || sip.BonusDone {
			continue
		}

		numVoter := len(sip.VoterMap)
		log.Debugf("splitSipBonus,  SIP<%d> prepare to split %d bonus to %d nodes", index, sip.Bonus, numVoter)

		consumed := uint64(0)
		for address, _ := range sip.VoterMap {
			amount := sip.Bonus / uint64(numVoter)

			log.Debugf("splitSipBonus,  SIP<%d> transfer %d(10^-9) bonus to %s", index, amount, address.ToBase58())

			err = appCallTransferOnt(native, utils.GovernanceContractAddress, address, uint64(amount))
			if err != nil {
				return fmt.Errorf("splitSipBonus, bonus transfer error: %v", err)
			}
			consumed += amount
		}

		//split bonus, no matter vote result
		voteRevenue.Total -= consumed
		voteRevenue.Reserve -= sip.Bonus
		sip.BonusDone = true

		log.Debugf("splitSipBonus,  SIP<%d> finish bonus split", index)
	}

	err = putSipVoteRevenue(native, contract, voteRevenue)
	if err != nil {
		return fmt.Errorf("splitSipBonus, put sip vote revenue error: %v", err)
	}

	err = putSipMap(native, contract, sipMap)
	if err != nil {
		return fmt.Errorf("putSipMap error: %v", err)
	}

	return nil
}

//trigger action for Sip reach threshold
func triggerSipAction(native *native.NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	sipMap, err := GetSipMap(native, contract)
	if err != nil {
		return fmt.Errorf("triggerSipAction, get sip Map error: %v", err)
	}

	for _, sip := range sipMap.SipMap {
		if sip.NumVotes < sip.MinVotes {
			continue
		}

		if sip.Result == EXEC {
			continue
		}

		if native.Height >= sip.Height {
			param, values, err := VerifySIPDetail(sip.Detail)
			if err != nil {
				return fmt.Errorf("triggerSipAction, fail to parse sip %d", sip.Index)
			}

			err = execSipAction(native, param, values)
			if err != nil {
				return fmt.Errorf("triggerSipAction, sip param %s take effect fail: %v", param, err)
			}
			log.Debugf("triggerSipAction, sip %d take effect on height:%d", sip.Index, native.Height)
			sip.Result = EXEC

		}

	}

	err = putSipMap(native, contract, sipMap)
	if err != nil {
		return fmt.Errorf("putSipMap error: %v", err)
	}
	return nil
}

//format should like below
// MinInitStake=1000
// ConsBonusSplitRate=50/50
// PoCSplitRate=5/10/85
func InitSipParamAttrs() {
	formatMap = make(map[string]*SipParamAttr)
	formatMap[SIP_MIN_INIT_STAKE] = &SipParamAttr{1, false}
	formatMap[SIP_CONS_BONUS_SPLIT_RATE] = &SipParamAttr{2, true}
	formatMap[SIP_POC_SPLIT_RATE] = &SipParamAttr{3, true}
	formatMap[SIP_PDP_GAS] = &SipParamAttr{1, false}
}

func VerifySIPDetail(detail []byte) (string, []uint64, error) {
	if formatMap == nil {
		InitSipParamAttrs()
	}

	//MinInitStake
	detailStr := string(detail)
	parts := strings.Split(detailStr, "=")
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("detail of sip has wrong format")
	}

	name := strings.Replace(parts[0], " ", "", -1)
	if _, ok := formatMap[name]; !ok {
		return "", nil, fmt.Errorf("unknow parameter name: %s", parts[0])
	}

	attr := formatMap[name]
	numValue := attr.NumValue
	valueStrs := strings.Split(parts[1], "/")
	if len(valueStrs) != numValue {
		return "", nil, fmt.Errorf("parameter %s need %d value", parts[0], numValue)
	}

	values := []uint64{}
	sum := uint64(0)
	for i := 0; i < numValue; i++ {
		valueStr := strings.Replace(valueStrs[i], " ", "", -1)
		value, err := strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			return "", nil, fmt.Errorf("fail to get value for parameter %s", name)
		}

		values = append(values, value)
		sum += value
	}

	if attr.NeedCheck {
		if sum != 100 {
			return "", nil, fmt.Errorf("%s need sum of %d value be 100", name, numValue)
		}
	}

	return name, values, nil
}

type SipFunc func(native *native.NativeService, values []uint32) error

// parse detail and run corresponding action
func execSipAction(native *native.NativeService, name string, values []uint64) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("execSipAction, getGlobalParam error: %v", err)
	}

	if name == SIP_MIN_INIT_STAKE {
		globalParam.MinInitStake = uint64(values[0])
		log.Debugf("execSipAction,  set %s to %d", SIP_MIN_INIT_STAKE, globalParam.MinInitStake)
	} else if name == SIP_CONS_BONUS_SPLIT_RATE {
		globalParam.A = uint32(values[0])
		globalParam.B = uint32(values[1])
		log.Debugf("execSipAction,  set %s to %d/%d", SIP_CONS_BONUS_SPLIT_RATE, globalParam.A, globalParam.B)
	} else if name == SIP_POC_SPLIT_RATE {
		globalParam.Cons = uint32(values[0])
		globalParam.Votes = uint32(values[1])
		globalParam.PoC = uint32(values[2])
		log.Debugf("execSipAction,  set %s(Cons/Votes/PoC) to %d/%d/%d", SIP_POC_SPLIT_RATE, globalParam.Cons,
			globalParam.Votes, globalParam.PoC)
	} else if name == SIP_PDP_GAS {
		/*
			err := appCallChangePDPGas(native, values[0])
			if err != nil {
				return fmt.Errorf("execSipAction, change pdp gas error: %v", err)
			}
			log.Debugf("execSipAction,  set %s to %d", SIP_PDP_GAS, values[0])
		*/
	}

	err = putParamChangeHeight(native, contract, name, native.Height)
	if err != nil {
		return fmt.Errorf("execSipAction, put putParamChangeHeight error: %v", err)
	}

	err = putGlobalParam(native, contract, globalParam)
	if err != nil {
		return fmt.Errorf("putGlobalParam, put globalParam error: %v", err)
	}
	return nil

}
