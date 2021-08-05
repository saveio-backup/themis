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

package genesis

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/constants"
	"github.com/saveio/themis/common/log"
	vconfig "github.com/saveio/themis/consensus/vbft/config"
	"github.com/saveio/themis/core/payload"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/core/utils"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/smartcontract/service/native/dns"
	"github.com/saveio/themis/smartcontract/service/native/global_params"
	"github.com/saveio/themis/smartcontract/service/native/governance"
	"github.com/saveio/themis/smartcontract/service/native/usdt"
	nutils "github.com/saveio/themis/smartcontract/service/native/utils"
	"github.com/saveio/themis/smartcontract/service/neovm"
)

const (
	BlockVersion uint32 = 0
	GenesisNonce uint64 = 2083236893
)

var (
	USDTToken   = newGoverningToken()
	USDTTokenID = USDTToken.Hash()
)

var GenBlockTime = (config.DEFAULT_GEN_BLOCK_TIME * time.Second)

var INIT_PARAM = map[string]string{
	"gasPrice": "0",
}

var GenesisBookkeepers []keypair.PublicKey

// BuildGenesisBlock returns the genesis block with default consensus bookkeeper list
func BuildGenesisBlock(defaultBookkeeper []keypair.PublicKey, genesisConfig *config.GenesisConfig) (*types.Block, error) {
	//getBookkeeper
	GenesisBookkeepers = defaultBookkeeper
	nextBookkeeper, err := types.AddressFromBookkeepers(defaultBookkeeper)
	if err != nil {
		return nil, fmt.Errorf("[Block],BuildGenesisBlock err with GetBookkeeperAddress: %s", err)
	}
	conf := common.NewZeroCopySink(nil)
	if genesisConfig.VBFT != nil {
		err := genesisConfig.VBFT.Serialization(conf)
		if err != nil {
			return nil, err
		}
	}
	govConfig := newGoverConfigInit(conf.Bytes())
	consensusPayload, err := vconfig.GenesisConsensusPayload(govConfig.Hash(), 0)
	if err != nil {
		return nil, fmt.Errorf("consensus genesis init failed: %s", err)
	}
	//blockdata
	genesisHeader := &types.Header{
		Version:          BlockVersion,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        constants.GENESIS_BLOCK_TIMESTAMP,
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextBookkeeper:   nextBookkeeper,
		ConsensusPayload: consensusPayload,

		Bookkeepers: nil,
		SigData:     nil,
	}

	//block
	usdt := newGoverningToken()
	param := newParamContract()
	oid := deployOntIDContract()
	auth := deployAuthContract()
	config := newConfig()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			usdt,
			param,
			oid,
			auth,
			config,
			newGoverningInit(),
			newParamInit(),
			newDNSInit(),
			govConfig,
		},
	}
	genesisBlock.RebuildMerkleRoot()
	genesisHash := genesisBlock.Hash()
	log.Infof("build genesis block %s", genesisHash.ToHexString())
	return genesisBlock, nil
}

func newGoverningToken() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.UsdtContractAddress[:], "USDT", "1.0",
		"xxx Team", "contact@xxx.io", "xxx USDT Token", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct usdt governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis governing token transaction error ")
	}
	return tx
}

func newParamContract() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.ParamContractAddress[:],
		"ParamConfig", "1.0", "Ontology Team", "contact@ont.io",
		"Chain Global Environment Variables Manager ", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis param transaction error ")
	}
	return tx
}

func newConfig() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.GovernanceContractAddress[:], "CONFIG", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network Consensus Config", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct config transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis config transaction error ")
	}
	return tx
}

func deployAuthContract() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.AuthContractAddress[:], "AuthContract", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network Authorization Contract", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis auth transaction error ")
	}
	return tx
}

func deployOntIDContract() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.OntIDContractAddress[:], "OID", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network ONT ID", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis ontid transaction error ")
	}
	return tx
}

func newGoverningInit() *types.Transaction {
	bookkeepers, _ := config.DefConfig.GetBookkeepers()

	var addr common.Address
	if len(bookkeepers) == 1 {
		addr = types.AddressFromPubKey(bookkeepers[0])
	} else {
		m := (5*len(bookkeepers) + 6) / 7
		temp, err := types.AddressFromMultiPubKeys(bookkeepers, m)
		if err != nil {
			panic(fmt.Sprint("wrong bookkeeper config, caused by", err))
		}
		addr = temp
	}

	var distribute []struct {
		addr  common.Address
		value uint64
	}

	if len(bookkeepers) == 1 {
		distribute = []struct {
			addr  common.Address
			value uint64
		}{{addr, constants.USDT_TOTAL_SUPPLY}}
		log.Infof("distribute %v to %v", constants.USDT_TOTAL_SUPPLY, addr)
	} else {
		distribute = []struct {
			addr  common.Address
			value uint64
		}{
			{nutils.GovernanceContractAddress, constants.USDT_TOTAL_SUPPLY - constants.USDT_FAUCEL_SUPPLY},
			{addr, constants.USDT_FAUCEL_SUPPLY},
		}
		log.Infof("distribute %v to %v", constants.USDT_TOTAL_SUPPLY-constants.USDT_FAUCEL_SUPPLY, nutils.GovernanceContractAddress)
		log.Infof("distribute %v to %v", constants.USDT_FAUCEL_SUPPLY, addr)
	}

	args := common.NewZeroCopySink(nil)
	nutils.EncodeVarUint(args, uint64(len(distribute)))
	for _, part := range distribute {
		nutils.EncodeAddress(args, part.addr)
		nutils.EncodeVarUint(args, part.value)
	}

	mutable := utils.BuildNativeTransaction(nutils.UsdtContractAddress, usdt.INIT_NAME, args.Bytes())
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis governing token transaction error ")
	}
	return tx
}

func newParamInit() *types.Transaction {
	params := new(global_params.Params)
	var s []string
	for k := range INIT_PARAM {
		s = append(s, k)
	}

	for k, v := range neovm.INIT_GAS_TABLE {
		INIT_PARAM[k] = strconv.FormatUint(v, 10)
		s = append(s, k)
	}

	sort.Strings(s)
	for _, v := range s {
		params.SetParam(global_params.Param{Key: v, Value: INIT_PARAM[v]})
	}
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	bookkeepers, _ := config.DefConfig.GetBookkeepers()
	var addr common.Address
	if len(bookkeepers) == 1 {
		addr = types.AddressFromPubKey(bookkeepers[0])
	} else {
		m := (5*len(bookkeepers) + 6) / 7
		temp, err := types.AddressFromMultiPubKeys(bookkeepers, m)
		if err != nil {
			panic(fmt.Sprint("wrong bookkeeper config, caused by", err))
		}
		addr = temp
	}
	nutils.EncodeAddress(sink, addr)

	mutable := utils.BuildNativeTransaction(nutils.ParamContractAddress, global_params.INIT_NAME, sink.Bytes())
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis governing token transaction error ")
	}
	return tx
}

func newGoverConfigInit(config []byte) *types.Transaction {
	mutable := utils.BuildNativeTransaction(nutils.GovernanceContractAddress, governance.INIT_CONFIG, config)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis governing token transaction error ")
	}
	return tx
}

func deployDNSContract() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.OntDNSAddress[:], "DnsContract", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network Authorization Contract", payload.NEOVM_TYPE)
	if err != nil {
		panic("constract genesis dns transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis dns transaction error ")
	}
	return tx
}
func newDNSInit() *types.Transaction {
	mutable := utils.BuildNativeTransaction(nutils.OntDNSAddress, dns.INIT_NAME, []byte{})
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis dns transaction error ")
	}
	return tx
}
