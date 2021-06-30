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

package savefs

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/saveio/themis/common"
)

func TestFileInfo_Serialize(t *testing.T) {
	fileHashStr := []byte("QmevhnWdtmz89BMXuuX5pSY2uZtqKLz7frJsrCojT5kmb6")
	addr := common.ADDRESS_EMPTY
	fileInfo := FileInfo{
		FileHash:       fileHashStr,
		FileOwner:      addr,
		FileDesc:       []byte("desc"),
		Privilege:      1,
		FileBlockNum:   2,
		FileBlockSize:  3,
		ProveInterval:  4,
		ProveTimes:     5,
		ExpiredHeight:  6,
		CopyNum:        7,
		Deposit:        8,
		FileProveParam: []byte{0x1},
		ProveBlockNum:  9,
		BlockHeight:    10,
		ValidFlag:      true,
		StorageType:    1,
		RealFileSize:   50,
	}

	buf := make([]byte, 0)
	source := common.NewZeroCopySink(buf)
	fileInfo.Serialization(source)

	source2 := common.NewZeroCopySource(source.Bytes())
	fileInfo2 := &FileInfo{}
	err := fileInfo2.Deserialization(source2)
	if err != nil {
		fmt.Printf("err : %s\n", err)
		t.Fatal(err)
	}
	fmt.Printf("fileInfo2.FileHash: %v\n", string(fileInfo2.FileHash[:]))
	fmt.Printf("fileInfo2.FileOwner: %v\n", string(fileInfo2.FileOwner[:]))
	fmt.Printf("fileInfo2.FileBlockNum: %v\n", fileInfo2.FileBlockNum)
	fmt.Printf("fileInfo2.FileBlockSize: %v\n", fileInfo2.FileBlockSize)
	fmt.Printf("fileInfo2.ProveInterval: %v\n", fileInfo2.ProveInterval)
	fmt.Printf("fileInfo2.ProveTimes: %v\n", fileInfo2.ProveTimes)
	fmt.Printf("fileInfo2.CopyNum: %v\n", fileInfo2.CopyNum)
	fmt.Printf("fileInfo2.Deposit: %v\n", fileInfo2.Deposit)
	fmt.Printf("fileInfo2.FileProveParam: %v\n", fileInfo2.FileProveParam)
	fmt.Printf("fileInfo2.ProveBlockNum: %v\n", fileInfo2.ProveBlockNum)
	fmt.Printf("fileInfo2.BlockHeight: %v\n", fileInfo2.BlockHeight)
	fmt.Printf("fileInfo2.ValidFlag: %v\n", fileInfo2.ValidFlag)
	fmt.Printf("fileInfo2.RealFileSize: %v\n", fileInfo2.RealFileSize)
}

func TestFileInfSeri(t *testing.T) {
	addr, err := common.AddressFromBase58("AHjjdbVLhfTyiNFEq2X8mFnnirZY1yK8Rq")
	if err != nil {
		t.Fatal(err)
	}
	addr1, _ := common.AddressFromBase58("ALQ6RWJENsELE7ATuzHz4zgHrq573xJsnM")
	addr2, _ := common.AddressFromBase58("AXWoKHswDpUxHjWqCZ5w5yCRgg6CPEViLJ")
	addr3, _ := common.AddressFromBase58("AQFBu9i2jGm3M6iVyvdbRLDnRYZAfgG7GC")
	addr4, _ := common.AddressFromBase58("AcNz5AnMWZhpmkp2fwCALNdYxcVbfjeS2E")
	fileInfo := FileInfo{
		FileHash:       []byte("QmeHGSRV6m9NmsrQJ3Rno2NkRnCzCFQJBXLoHFggZz8a7K"),
		FileOwner:      addr,
		FileDesc:       []byte("2019-08-19_17.22.59_LOG.log"),
		Privilege:      1,
		FileBlockNum:   86,
		FileBlockSize:  256,
		ProveInterval:  300,
		ProveTimes:     0,
		ExpiredHeight:  22545,
		CopyNum:        2,
		Deposit:        0,
		FileProveParam: []byte{0x81, 0x1, 0x23, 0x5d, 0xac, 0xba, 0x40, 0x87, 0x30, 0xf8, 0xb3, 0x5a, 0x8f, 0x5a, 0x61, 0x14, 0x87, 0x41, 0xee, 0x22, 0xbe, 0xd9, 0x17, 0x67, 0x89, 0x57, 0x3f, 0xea, 0x7b, 0xd7, 0xb0, 0x40, 0x72, 0xb3, 0x19, 0xb2, 0x9f, 0xc2, 0x7e, 0x6c, 0x64, 0xa4, 0xc5, 0x51, 0x4b, 0x43, 0x1c, 0x2b, 0x8, 0xa3, 0x8c, 0x8d, 0x6f, 0xeb, 0xd4, 0x83, 0xdf, 0x5a, 0x9e, 0x38, 0x85, 0x4a, 0x7b, 0x24, 0x1d, 0x7d, 0x7b, 0xe2, 0xef, 0xa5, 0x66, 0x3b, 0xb1, 0x46, 0x1, 0x89, 0xaa, 0x24, 0xb2, 0xe2, 0x91, 0xb9, 0x73, 0xe7, 0xa8, 0xbb, 0xaa, 0x4c, 0x4f, 0xc4, 0x99, 0xbc, 0x81, 0x94, 0x5f, 0xb9, 0xae, 0xe5, 0x31, 0x2e, 0xfd, 0x85, 0x80, 0xc6, 0xba, 0x8a, 0x6e, 0xd2, 0xcb, 0x1d, 0x67, 0x24, 0xe4, 0x7f, 0x7e, 0x19, 0xbe, 0xc8, 0xf1, 0xb4, 0xb, 0xb1, 0xfa, 0x6, 0x9f, 0xef, 0x19, 0xdf, 0x1c, 0x47, 0x40, 0x50, 0xdb, 0xdc, 0x96, 0xd1, 0x22, 0x5e, 0xaa, 0x62, 0x9a, 0xa3, 0xc4, 0x79, 0xe2, 0xcd, 0xe2, 0xb2, 0x8, 0x7b, 0xb0, 0x4f, 0xaa, 0xf2, 0x8d, 0x59, 0x1b, 0x4c, 0xe, 0x24, 0x51, 0x52, 0x3a, 0x2d, 0xa3, 0x24, 0xb3, 0xb, 0x99, 0xd, 0x4f, 0x47, 0x74, 0x63, 0xc, 0xd3, 0x65, 0x5, 0xb2, 0x7b, 0x7f, 0xcc, 0x2e, 0xf9, 0x50, 0x30, 0xd6, 0x1a, 0xaa, 0x85, 0x38, 0xa0, 0xe9, 0x18, 0x49, 0x81, 0x1, 0x87, 0x4e, 0x86, 0xf, 0xcc, 0x99, 0x45, 0x10, 0x65, 0x53, 0xe6, 0x73, 0x47, 0x7d, 0x2, 0xec, 0x69, 0x27, 0xed, 0xaa, 0x9a, 0xb, 0xbf, 0x4c, 0x7, 0xc, 0xa3, 0x72, 0x69, 0xfd, 0x1, 0x38, 0x4f, 0xa9, 0x1f, 0xf4, 0xff, 0x25, 0x53, 0x43, 0x9b, 0xa1, 0x40, 0xcc, 0x7f, 0x98, 0x18, 0xe7, 0x68, 0x92, 0x26, 0x76, 0xa9, 0xe2, 0x2c, 0x39, 0xb2, 0xde, 0x60, 0xfa, 0xb8, 0xcf, 0xb8, 0x8d, 0x5d, 0xc4, 0xc7, 0xb, 0x42, 0xe7, 0xef, 0xd6, 0xbd, 0xaf, 0xb6, 0x87, 0x41, 0xa7, 0x14, 0x30, 0x0, 0x6, 0x9, 0x7b, 0x9, 0x3b, 0x34, 0xaf, 0x73, 0xfe, 0x5e, 0x86, 0xf2, 0x2, 0x9b, 0xf1, 0x13, 0x36, 0xee, 0x2a, 0x75, 0x1, 0xdb, 0xf7, 0x43, 0xfa, 0xfa, 0x9e, 0xee, 0xf5, 0xa, 0xbc, 0x99, 0xe7, 0xab, 0x62, 0x21, 0xcd, 0x9c, 0x4d, 0x94, 0x52, 0x3e, 0xac, 0xfd, 0x91, 0x16, 0xc8, 0x20, 0x43, 0xda, 0xfb, 0xd9, 0x8d, 0x96, 0xa, 0xed, 0x13, 0x26, 0x49, 0xd0, 0xd, 0x89, 0xd7, 0xa0, 0x0, 0x5e, 0xaf, 0xc4, 0x8b, 0xa4, 0xf7, 0x3c, 0xee, 0x14, 0xe2, 0xd2, 0x67, 0x19, 0xcc, 0x82},
		ProveBlockNum:  0,
		BlockHeight:    0,
		ValidFlag:      true,
		StorageType:    0,
		RealFileSize:   21736,
		PrimaryNodes: NodeList{
			AddrNum:  3,
			AddrList: []common.Address{addr1, addr2, addr3},
		},
		CandidateNodes: NodeList{
			AddrNum:  1,
			AddrList: []common.Address{addr4},
		},
	}
	buf := make([]byte, 0)
	source := common.NewZeroCopySink(buf)
	fileInfo.Serialization(source)
	fmt.Printf("%x\n", source.Bytes())

}

func TestFileInfoDerialize(t *testing.T) {
	hexData := "2e516d654847535256366d394e6d7372514a33526e6f324e6b526e437a4346514a42584c6f484667675a7a3861374b141588ba84979fe4869df0c6693aae32d4995c676b1b323031392d30382d31395f31372e32322e35395f4c4f472e6c6f6701010156020001022c0100021158010200fd66018101235dacba408730f8b35a8f5a61148741ee22bed9176789573fea7bd7b04072b319b29fc27e6c64a4c5514b431c2b08a38c8d6febd483df5a9e38854a7b241d7d7be2efa5663bb1460189aa24b2e291b973e7a8bbaa4c4fc499bc81945fb9aee5312efd8580c6ba8a6ed2cb1d6724e47f7e19bec8f1b40bb1fa069fef19df1c474050dbdc96d1225eaa629aa3c479e2cde2b2087bb04faaf28d591b4c0e2451523a2da324b30b990d4f4774630cd36505b27b7fcc2ef95030d61aaa8538a0e918498101874e860fcc9945106553e673477d02ec6927edaa9a0bbf4c070ca37269fd01384fa91ff4ff2553439ba140cc7f9818e768922676a9e22c39b2de60fab8cfb88d5dc4c70b42e7efd6bdafb68741a714300006097b093b34af73fe5e86f2029bf11336ee2a7501dbf743fafa9eeef50abc99e7ab6221cd9c4d94523eacfd9116c82043dafbd98d960aed132649d00d89d7a0005eafc48ba4f73cee14e2d26719cc82000001010002e85401031432ba2482da3b595e667c0a1c243fc6b25595782e14aca8285c0a5975b89d9fc1cc9cc7b95256b06a52145ceb7948d9a9d35a527b7460676fc6c4eb3a5d3e010114e206701e4d9af1426f8bb39d17976044fd6be8ef"
	// hexData := "2e516d59745579666662414a71486b516b4d54527a71446446586e6f744c3656744d50725379706a566e317a427039141588ba84979fe4869df0c6693aae32d4995c676b1b323031392d30382d31395f31362e35312e34395f4c4f472e6c6f6701010152020001022c0100021158010200fd660181015d9b61ae28da2b8879c38d8be285d3ca6fe1643727a515a3443549b4677dd3cb540dcb92f8583e74eef3ac4d81280937f16fd647227ea0b1077762d0dc726b148ea1a0ab007bf7dc381675386d3af861c204078735c18889e2d46da4f05ae09a6a53fd20d9127b6acb4f7e52c534a31d7a49ce37f7082e2edf8b0c79aa67127f4032a8de8b3d988dc3a50e94ba0cf0e7284e66184067a4ee4b5eae03545fd6ba182b55c6b410cd8ac659abf1981fedc71c645403c4003900de7dd6d8b0aa687dd2810181c559ed4e2ce0d54f4e8ecec8f55d87358e109c5587d2b1130e7f4cec05d7a302960be9b92f5ad591537e1130c4b8c497385dc9453c570b2e742b083c76483a4fb2e1075eda8c91ca6925b2224a5dc26ac557df4bd2a1bca3ff9fe57e4ef29a454db04f2768ab094e1ef40d2750c89f5c9c525abbfdcfa3159cf749e82c8b0520de93a1e163bb2ebee03f9e617926c25d7d96511832ca0f63ea68cb11a3febdd600000101000289500103010314aca8285c0a5975b89d9fc1cc9cc7b95256b06a52145ceb7948d9a9d35a527b7460676fc6c4eb3a5d3e14e206701e4d9af1426f8bb39d17976044fd6be8ef010101011432ba2482da3b595e667c0a1c243fc6b25595782e"
	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatal(err)
	}
	var fi FileInfo
	source := common.NewZeroCopySource(data)
	err = fi.Deserialization(source)
	if err != nil {
		t.Fatal(err)
	}
}
