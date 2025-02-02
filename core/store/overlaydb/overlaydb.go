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

package overlaydb

import (
	"crypto/sha256"

	comm "github.com/saveio/themis/common"
	"github.com/saveio/themis/core/store/common"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type OverlayDB struct {
	store common.PersistStore
	memdb *MemDB
	dbErr error
}

const initCap = 4 * 1024
const initkvNum = 128
const initDummyCap = 4 * 1024
const initDummykvNum = 4

func NewOverlayDB(store common.PersistStore) *OverlayDB {
	return &OverlayDB{
		store: store,
		memdb: NewMemDB(initCap, initkvNum),
	}
}

func NewDummyOverlayDB(store common.PersistStore) *OverlayDB {
	return &OverlayDB{
		store: store,
		memdb: NewMemDB(initDummyCap, initDummykvNum),
	}
}

func (self *OverlayDB) Reset() {
	self.memdb.Reset()
}

func (self *OverlayDB) Error() error {
	return self.dbErr
}

func (self *OverlayDB) SetError(err error) {
	self.dbErr = err
}

// if key is deleted, value == nil
func (self *OverlayDB) Get(key []byte) (value []byte, err error) {
	var unknown bool
	value, unknown = self.memdb.Get(key)
	if !unknown {
		return value, nil
	}

	value, err = self.store.Get(key)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, nil
		}
		self.dbErr = err
		return nil, err
	}

	return
}

func (self *OverlayDB) Put(key []byte, value []byte) {
	self.memdb.Put(key, value)
}

func (self *OverlayDB) Delete(key []byte) {
	self.memdb.Delete(key)
}

func (self *OverlayDB) CommitTo() {
	self.memdb.ForEach(func(key, val []byte) {
		if len(val) == 0 {
			self.store.BatchDelete(key)
		} else {
			self.store.BatchPut(key, val)
		}
	})
}

func (self *OverlayDB) GetWriteSet() *MemDB {
	return self.memdb
}

func (self *OverlayDB) ChangeHash() comm.Uint256 {
	stateDiff := sha256.New()
	self.memdb.ForEach(func(key, val []byte) {
		stateDiff.Write(key)
		stateDiff.Write(val)
	})

	var hash comm.Uint256
	stateDiff.Sum(hash[:0])
	return hash
}

// param key is referenced by iterator
func (self *OverlayDB) NewIterator(key []byte) common.StoreIterator {
	prefixRange := util.BytesPrefix(key)
	backIter := self.store.NewIterator(key)
	memIter := self.memdb.NewIterator(prefixRange)

	return NewJoinIter(memIter, backIter)
}
