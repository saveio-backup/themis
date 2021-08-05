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

package constants

// genesis constants
var (
	//TODO: modify this when on mainnet
	GENESIS_BLOCK_TIMESTAMP = uint32(1530316800)
)

// usdt constants
const (
	USDT_NAME          = "USD Token"
	USDT_SYMBOL        = "USDT"
	USDT_DECIMALS      = 9
	USDT_TOTAL_SUPPLY  = uint64(100000000000000000)
	USDT_FAUCEL_SUPPLY = uint64(10000000000000000)
)

// ont/ong unbound model constants
const UNBOUND_TIME_INTERVAL = uint32(31536000)

var UNBOUND_GENERATION_AMOUNT = [18]uint64{5, 4, 3, 3, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

// the end of unbound timestamp offset from genesis block's timestamp
var UNBOUND_DEADLINE = (func() uint32 {
	return 0

	count := uint64(0)
	for _, m := range UNBOUND_GENERATION_AMOUNT {
		count += m
	}
	count *= uint64(UNBOUND_TIME_INTERVAL)

	numInterval := len(UNBOUND_GENERATION_AMOUNT)

	if UNBOUND_GENERATION_AMOUNT[numInterval-1] != 1 ||
		!(count-uint64(UNBOUND_TIME_INTERVAL) < USDT_TOTAL_SUPPLY && USDT_TOTAL_SUPPLY <= count) {
		panic("incompatible constants setting")
	}

	return UNBOUND_TIME_INTERVAL*uint32(numInterval) - uint32(count-uint64(USDT_TOTAL_SUPPLY))
})()

// multi-sig constants
const MULTI_SIG_MAX_PUBKEY_SIZE = 16

// transaction constants
const TX_MAX_SIG_SIZE = 16

// network magic number
const (
	NETWORK_MAGIC_MAINNET = 0x8c77ab60
	NETWORK_MAGIC_POLARIS = 0x2d8829df
)

// ledger state hash check height
const STATE_HASH_HEIGHT_MAINNET = 0
const STATE_HASH_HEIGHT_POLARIS = 850000

// neovm opcode update check height
const OPCODE_HEIGHT_UPDATE_FIRST_MAINNET = 6300000
const OPCODE_HEIGHT_UPDATE_FIRST_POLARIS = 2100000

// gas round tune operation height
const GAS_ROUND_TUNE_HEIGHT_MAINNET = 8500000
const GAS_ROUND_TUNE_HEIGHT_POLARIS = 10100000

const CONTRACT_DEPRECATE_API_HEIGHT_MAINNET = 8600000
const CONTRACT_DEPRECATE_API_HEIGHT_POLARIS = 13000000

// self gov register height
const BLOCKHEIGHT_SELFGOV_REGISTER_MAINNET = 8600000
const BLOCKHEIGHT_SELFGOV_REGISTER_POLARIS = 12150000

const BLOCKHEIGHT_NEW_ONTID_MAINNET = 9000000
const BLOCKHEIGHT_NEW_ONTID_POLARIS = 12150000

const BLOCKHEIGHT_ONTFS_MAINNET = 8550000
const BLOCKHEIGHT_ONTFS_POLARIS = 12250000

const BLOCKHEIGHT_CC_POLARIS = 13130000

//new node cost height
const BLOCKHEIGHT_NEW_PEER_COST_MAINNET = 9400000
const BLOCKHEIGHT_NEW_PEER_COST_POLARIS = 13400000
