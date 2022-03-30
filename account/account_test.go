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
package account

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/saveio/themis/core/signature"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/crypto/ec"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	defer func() {
		os.RemoveAll("Log/")
	}()

	names := []string{
		"",
		"SHA224withECDSA",
		"SHA256withECDSA",
		"SHA384withECDSA",
		"SHA512withECDSA",
		"SHA3-224withECDSA",
		"SHA3-256withECDSA",
		"SHA3-384withECDSA",
		"SHA3-512withECDSA",
		"RIPEMD160withECDSA",
		"SM3withSM2",
		"SHA512withEdDSA",
	}
	accounts := make([]*Account, len(names))
	for k, v := range names {
		accounts[k] = NewAccount(v)
		assert.NotNil(t, accounts[k])
		assert.NotNil(t, accounts[k].PrivateKey)
		assert.NotNil(t, accounts[k].PublicKey)
		assert.NotNil(t, accounts[k].Address)
		assert.NotNil(t, accounts[k].PrivKey())
		assert.NotNil(t, accounts[k].PubKey())
		assert.NotNil(t, accounts[k].Scheme())
		assert.Equal(t, accounts[k].Address, types.AddressFromPubKey(accounts[k].PublicKey))
	}
}

func TestGenerateETHAccount(t *testing.T) {
	pv, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	privateKey := hex.EncodeToString(ethCrypto.FromECDSA(pv))
	addr := ethCrypto.PubkeyToAddress(pv.PublicKey)
	fmt.Printf("pk %s, addr %s\n", privateKey, addr.String())
}

func TestGenerateAccount(t *testing.T) {
	acc1 := NewAccount("SHA256withECDSA")
	fmt.Printf("Account1 themis address %v\n", acc1.Address)
	fmt.Printf("Account1 ethereum address %v\n", acc1.EthAddress.String())
	// acc1EcPk := acc1.PrivKey().(*ec.PrivateKey)
	fmt.Printf("Account1 PublicKey %x, PrivKey %x\n", keypair.SerializePublicKey(acc1.PublicKey), keypair.SerializePrivateKey(acc1.PrivateKey))
	fmt.Printf("Account1 privateKey %x\n", acc1.GetEthPrivateKey())

	acc2 := NewAccountWithPrivateKey(acc1.GetEthPrivateKey())

	fmt.Printf("Account2 themis address %v\n", acc2.Address)
	fmt.Printf("Account2 ethereum address %v\n", acc2.EthAddress.String())
	acc2EcPk := acc2.PrivKey().(*ec.PrivateKey)
	fmt.Printf("Account2 PublicKey %v, PrivKey %v\n", acc2EcPk.PublicKey, acc2EcPk.PrivateKey)
	fmt.Printf("Account2 privateKey %x\n", acc2.GetEthPrivateKey())
}

func TestNewAccountPubKey(t *testing.T) {
	acc1 := NewAccount("SHA256withECDSA")
	fmt.Printf("Account1 themis address %v\n", acc1.Address)
	fmt.Printf("Account1 ethereum address %v\n", acc1.EthAddress.String())
	acc1EcPk := acc1.PrivKey().(*ec.PrivateKey)
	fmt.Printf("Account1 PublicKey %v, PrivKey %v\n", acc1EcPk.PublicKey, acc1EcPk.PrivateKey)
	fmt.Printf("Account1 privateKey %x\n", acc1.GetEthPrivateKey())

	pubKeyData := keypair.SerializePublicKey(acc1.PublicKey)
	fmt.Printf("Pubkey %x\n", pubKeyData)

	pubKey, err := keypair.DeserializePublicKey(pubKeyData)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("pubKey %v\n", pubKey)
}

func TestAccountSignTx(t *testing.T) {

	pk, _ := hex.DecodeString("db0e8a343062823ee611423d11b11e1ed4cdde442ec204f80406fa084c601002")
	acc1 := NewAccountWithPrivateKey(pk)

	fmt.Printf("acc++ %s\n", acc1.Address.ToBase58())
	fmt.Printf("acc from++ %s\n", types.AddressFromPubKey(acc1.GetPublicKey()))
	fmt.Printf("pubkey %x\n", keypair.SerializePublicKey(acc1.PublicKey))
	fmt.Printf("EthAddress++ %s\n", acc1.EthAddress.String())

	realData, err := hex.DecodeString("1d1c7304fbb608f507f891680e2e6f745f109406d1cc037a9608454b3338110e")
	if err != nil {
		t.Fatal(err)
	}
	// data := []byte("1d1c7304fbb608f507f891680e2e6f745f109406d1cc037a9608454b3338110e")
	sigData, err := signature.Sign(acc1, realData)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("sigData %x\n", sigData)
	err = signature.Verify(
		acc1.PublicKey, realData, sigData,
	)
	if err != nil {
		t.Fatal(err)
	}
}
