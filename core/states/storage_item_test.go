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
package states

import (
	"testing"

	"github.com/saveio/themis/common"
)

func TestStorageItem_Serialize_Deserialize(t *testing.T) {

	item := &StorageItem{
		StateBase: StateBase{StateVersion: 1},
		Value:     []byte{1},
	}

	bf := common.NewZeroCopySink(nil)
	item.Serialization(bf)

	var storage = new(StorageItem)
	source := common.NewZeroCopySource(bf.Bytes())
	if err := storage.Deserialization(source); err != nil {
		t.Fatalf("StorageItem deserialize error: %v", err)
	}
}
