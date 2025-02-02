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
	"reflect"
	"testing"

	"github.com/saveio/themis/common"
	vconfig "github.com/saveio/themis/consensus/vbft/config"
	"github.com/saveio/themis/core/types"
)

func TestBlock_getProposer(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	type fields struct {
		Block *types.Block
		Info  *vconfig.VbftBlockInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name:   "test",
			fields: fields{Block: blk.Block, Info: blk.Info},
			want:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blk := &Block{
				Block: tt.fields.Block,
				Info:  tt.fields.Info,
			}
			if got := blk.getProposer(); got != tt.want {
				t.Errorf("Block.getProposer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_getBlockNum(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	type fields struct {
		Block *types.Block
		Info  *vconfig.VbftBlockInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name:   "test",
			fields: fields{Block: blk.Block, Info: blk.Info},
			want:   uint32(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blk := &Block{
				Block: tt.fields.Block,
				Info:  tt.fields.Info,
			}
			if got := blk.getBlockNum(); got != tt.want {
				t.Errorf("Block.getBlockNum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_getPrevBlockHash(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	type fields struct {
		Block *types.Block
		Info  *vconfig.VbftBlockInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   common.Uint256
	}{
		{
			name:   "test",
			fields: fields{Block: blk.Block, Info: blk.Info},
			want:   common.Uint256{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blk := &Block{
				Block: tt.fields.Block,
				Info:  tt.fields.Info,
			}
			if got := blk.getPrevBlockHash(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Block.getPrevBlockHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_getLastConfigBlockNum(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}

	type fields struct {
		Block *types.Block
		Info  *vconfig.VbftBlockInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name:   "test",
			fields: fields{Block: blk.Block, Info: blk.Info},
			want:   uint32(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blk := &Block{
				Block: tt.fields.Block,
				Info:  tt.fields.Info,
			}
			if got := blk.getLastConfigBlockNum(); got != tt.want {
				t.Errorf("Block.getLastConfigBlockNum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_getNewChainConfig(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	type fields struct {
		Block *types.Block
		Info  *vconfig.VbftBlockInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   *vconfig.ChainConfig
	}{
		{
			name:   "test",
			fields: fields{Block: blk.Block, Info: blk.Info},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blk := &Block{
				Block: tt.fields.Block,
				Info:  tt.fields.Info,
			}
			if got := blk.getNewChainConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Block.getNewChainConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSerialize(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	blk.Serialize()
	t.Log("Block Serialize succ")
}

func TestInitVbftBlock(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	_, err = initVbftBlock(blk.Block, nil, common.Uint256{})
	if err != nil {
		t.Errorf("initVbftBlock failed: %v", err)
		return
	}
	t.Log("TestInitVbftBlock succ")
}
