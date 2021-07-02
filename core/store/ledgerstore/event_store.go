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

package ledgerstore

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/common/serialization"
	scom "github.com/saveio/themis/core/store/common"
	"github.com/saveio/themis/core/store/leveldbstore"
	"github.com/saveio/themis/smartcontract/event"
)

//Saving event notifies gen by smart contract execution
type EventStore struct {
	dbDir string                     //Store path
	store *leveldbstore.LevelDBStore //Store handler
}

//NewEventStore return event store instance
func NewEventStore(dbDir string) (*EventStore, error) {
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	return &EventStore{
		dbDir: dbDir,
		store: store,
	}, nil
}

//NewBatch start event commit batch
func (this *EventStore) NewBatch() {
	this.store.NewBatch()
}

//SaveEventNotifyByTx persist event notify by transaction hash
func (this *EventStore) SaveEventNotifyByTx(txHash common.Uint256, notify *event.ExecuteNotify) error {
	result, err := json.Marshal(notify)
	if err != nil {
		return fmt.Errorf("json.Marshal error %s", err)
	}
	key := genEventNotifyByTxKey(txHash)
	this.store.BatchPut(key, result)

	return this.SaveEventNofityByEventID(txHash, notify)
}

func (this *EventStore) SaveEventNofityByEventID(txHash common.Uint256, notify *event.ExecuteNotify) error {
	if notify == nil {
		return fmt.Errorf("notify is nil")
	}

	exist := make(map[string]struct{}, 0)
	for _, notifyInfo := range notify.Notify {
		// event id 0 means no need to store
		if notifyInfo.EventIdentifier == 0 {
			continue
		}

		var addresses []common.Address

		if len(notifyInfo.Addresses) == 0 {
			addresses = append(addresses, common.ADDRESS_EMPTY)
		} else {
			addresses = notifyInfo.Addresses
		}

		for _, address := range addresses {
			existKey := fmt.Sprintf("%s-%s-%d", notifyInfo.ContractAddress.ToBase58(), address.ToBase58(), notifyInfo.EventIdentifier)
			if _, ok := exist[existKey]; ok {
				continue
			}
			// use random to distinguish events with same event id
			random := rand.Uint32()
			key, err := this.getEventNotifyByEventIDKey(notifyInfo.ContractAddress, address, notifyInfo.EventIdentifier, random)
			if err != nil {
				return err
			}
			value := bytes.NewBuffer(nil)
			txHash.Serialize(value)

			this.store.BatchPut(key, value.Bytes())
			exist[existKey] = struct{}{}
		}
	}
	return nil
}

func (this *EventStore) getEventNotifyByEventIDKey(contractAddress common.Address, address common.Address, eventId uint32, random uint32) ([]byte, error) {
	prefix, err := this.getEventNotifyByEventIDKeyPrefix(contractAddress, address, eventId)
	if err != nil {
		return nil, err
	}

	key := bytes.NewBuffer(prefix)
	serialization.WriteUint32(key, random)

	return key.Bytes(), nil
}

func (this *EventStore) getEventNotifyByEventIDKeyPrefix(contractAddress common.Address, address common.Address, eventId uint32) ([]byte, error) {
	key := bytes.NewBuffer(nil)
	err := contractAddress.Serialize(key)
	if err != nil {
		return nil, err
	}

	err = address.Serialize(key)
	if err != nil {
		return nil, err
	}

	// eventId 0 means find all the events
	if eventId != 0 {
		serialization.WriteUint32(key, eventId)
	}

	return key.Bytes(), nil
}

func (this *EventStore) getAllEventNotifyKeyPrefix(contractAddress common.Address) ([]byte, error) {
	key := bytes.NewBuffer(nil)
	err := contractAddress.Serialize(key)
	if err != nil {
		return nil, err
	}
	return key.Bytes(), nil
}

func (this *EventStore) GetEventNotifyTxHashByHeights(contractAddress common.Address, addressBytes []byte, eventId uint32) ([]common.Uint256, error) {
	var txHashes []common.Uint256

	var prefix []byte
	if addressBytes != nil {
		var err error
		var address common.Address
		copy(address[:], addressBytes[:])
		prefix, err = this.getEventNotifyByEventIDKeyPrefix(contractAddress, address, eventId)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		prefix, err = this.getAllEventNotifyKeyPrefix(contractAddress)
		if err != nil {
			return nil, err
		}
	}

	iter := this.store.NewIterator(prefix)

	defer iter.Release()
	for iter.Next() {
		var txHash common.Uint256

		reader := bytes.NewBuffer(iter.Value())
		err := txHash.Deserialize(reader)
		if err != nil {
			return nil, fmt.Errorf("ReadUint32 error %s", err)
		}

		txHashes = append(txHashes, txHash)
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return txHashes, nil
}

func (this *EventStore) GetEventNotifyTxHashByEventID(contractAddress common.Address, address common.Address, eventId uint32) ([]common.Uint256, error) {
	var txHashes []common.Uint256

	prefix, err := this.getEventNotifyByEventIDKeyPrefix(contractAddress, address, eventId)
	if err != nil {
		return nil, err
	}

	iter := this.store.NewIterator(prefix)

	defer iter.Release()
	for iter.Next() {
		var txHash common.Uint256

		reader := bytes.NewBuffer(iter.Value())
		err := txHash.Deserialize(reader)
		if err != nil {
			return nil, fmt.Errorf("ReadUint32 error %s", err)
		}

		txHashes = append(txHashes, txHash)
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return txHashes, nil
}

//SaveEventNotifyByBlock persist transaction hash which have event notify to store
func (this *EventStore) SaveEventNotifyByBlock(height uint32, txHashs []common.Uint256) {
	key := genEventNotifyByBlockKey(height)
	values := common.NewZeroCopySink(nil)
	values.WriteUint32(uint32(len(txHashs)))
	for _, txHash := range txHashs {
		values.WriteHash(txHash)
	}
	this.store.BatchPut(key, values.Bytes())
}

//GetEventNotifyByTx return event notify by trasanction hash
func (this *EventStore) GetEventNotifyByTx(txHash common.Uint256) (*event.ExecuteNotify, error) {
	key := genEventNotifyByTxKey(txHash)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	var notify event.ExecuteNotify
	if err = json.Unmarshal(data, &notify); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error %s", err)
	}
	return &notify, nil
}

//GetEventNotifyByBlock return all event notify of transaction in block
func (this *EventStore) GetEventNotifyByBlock(height uint32) ([]*event.ExecuteNotify, error) {
	key := genEventNotifyByBlockKey(height)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewBuffer(data)
	size, err := serialization.ReadUint32(reader)
	if err != nil {
		return nil, fmt.Errorf("ReadUint32 error %s", err)
	}
	evtNotifies := make([]*event.ExecuteNotify, 0)
	for i := uint32(0); i < size; i++ {
		var txHash common.Uint256
		err = txHash.Deserialize(reader)
		if err != nil {
			return nil, fmt.Errorf("txHash.Deserialize error %s", err)
		}
		evtNotify, err := this.GetEventNotifyByTx(txHash)
		if err != nil {
			log.Errorf("getEventNotifyByTx Height:%d by txhash:%s error:%s", height, txHash.ToHexString(), err)
			continue
		}
		evtNotifies = append(evtNotifies, evtNotify)
	}
	return evtNotifies, nil
}

func (this *EventStore) PruneBlock(height uint32, hashes []common.Uint256) {
	key := genEventNotifyByBlockKey(height)
	this.store.BatchDelete(key)
	for _, hash := range hashes {
		this.store.BatchDelete(genEventNotifyByTxKey(hash))
	}
}

//CommitTo event store batch to store
func (this *EventStore) CommitTo() error {
	return this.store.BatchCommit()
}

//Close event store
func (this *EventStore) Close() error {
	return this.store.Close()
}

//ClearAll all data in event store
func (this *EventStore) ClearAll() error {
	this.NewBatch()
	iter := this.store.NewIterator(nil)
	for iter.Next() {
		this.store.BatchDelete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}
	return this.CommitTo()
}

//SaveCurrentBlock persist current block height and block hash to event store
func (this *EventStore) SaveCurrentBlock(height uint32, blockHash common.Uint256) {
	key := this.getCurrentBlockKey()
	value := common.NewZeroCopySink(nil)
	value.WriteHash(blockHash)
	value.WriteUint32(height)
	this.store.BatchPut(key, value.Bytes())
}

//GetCurrentBlock return current block hash, and block height
func (this *EventStore) GetCurrentBlock() (common.Uint256, uint32, error) {
	key := this.getCurrentBlockKey()
	data, err := this.store.Get(key)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	reader := bytes.NewReader(data)
	blockHash := common.Uint256{}
	err = blockHash.Deserialize(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	height, err := serialization.ReadUint32(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	return blockHash, height, nil
}

func (this *EventStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func genEventNotifyByBlockKey(height uint32) []byte {
	key := make([]byte, 5, 5)
	key[0] = byte(scom.EVENT_NOTIFY)
	binary.LittleEndian.PutUint32(key[1:], height)
	return key
}

func genEventNotifyByTxKey(txHash common.Uint256) []byte {
	data := txHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.EVENT_NOTIFY)
	copy(key[1:], data)
	return key
}
