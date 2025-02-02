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
	"bytes"
	ethComm "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/crypto/ec"
	"github.com/saveio/themis/crypto/keypair"
	s "github.com/saveio/themis/crypto/signature"
)

/* crypto object */
type Account struct {
	PrivateKey keypair.PrivateKey
	PublicKey  keypair.PublicKey
	Address    common.Address
	EthAddress ethComm.Address
	SigScheme  s.SignatureScheme
}

func NewAccount(encrypt string) *Account {
	// Determine the public key algorithm and parameters according to
	// the encrypt.
	// var pkAlgorithm keypair.KeyType
	// var params interface{}
	// var scheme s.SignatureScheme
	// var err error
	// if "" != encrypt {
	// 	scheme, err = s.GetScheme(encrypt)
	// } else {
	// 	scheme = s.SHA256withECDSA
	// }
	// if err != nil {
	// 	log.Warn("unknown signature scheme, use SHA256withECDSA as default.")
	// 	scheme = s.SHA256withECDSA
	// }
	// switch scheme {
	// case s.SHA224withECDSA, s.SHA3_224withECDSA:
	// 	pkAlgorithm = keypair.PK_ECDSA
	// 	params = keypair.P224
	// case s.SHA256withECDSA, s.SHA3_256withECDSA, s.RIPEMD160withECDSA:
	// 	pkAlgorithm = keypair.PK_ECDSA
	// 	params = keypair.P256
	// case s.SHA384withECDSA, s.SHA3_384withECDSA:
	// 	pkAlgorithm = keypair.PK_ECDSA
	// 	params = keypair.P384
	// case s.SHA512withECDSA, s.SHA3_512withECDSA:
	// 	pkAlgorithm = keypair.PK_ECDSA
	// 	params = keypair.P521
	// case s.SM3withSM2:
	// 	pkAlgorithm = keypair.PK_SM2
	// 	params = keypair.SM2P256V1
	// case s.SHA512withEDDSA:
	// 	pkAlgorithm = keypair.PK_EDDSA
	// 	params = keypair.ED25519
	// }

	// only support SHA256withECDSA to generate ethereum account
	pv, err := ethCrypto.GenerateKey()
	if err != nil {
		return nil
	}
	return NewAccountWithPrivateKey(ethCrypto.FromECDSA(pv))
}

func (this *Account) PrivKey() keypair.PrivateKey {
	return this.PrivateKey
}

func (this *Account) PubKey() keypair.PublicKey {
	return this.PublicKey
}

func (this *Account) Scheme() s.SignatureScheme {
	return this.SigScheme
}

func (this *Account) GetEthPrivateKey() []byte {
	ecdsaPrivateKey := this.PrivateKey.(*ec.PrivateKey)
	return ecdsaPrivateKey.Raw
}

func (this *Account) GetPublicKey() *ec.PublicKey {
	return this.PublicKey.(*ec.PublicKey)
}

func (this *Account) GetEthPublicKey() []byte {
	pubKey := keypair.GetEthPublicKeyFromPrivateKey(this.PrivateKey)
	return ethCrypto.FromECDSAPub(pubKey)
}

func NewAccountWithPrivateKey(privateKey []byte) *Account {

	pri, pub, err := keypair.GenerateKeyPairWithSeed(
		keypair.PK_ECDSA,
		bytes.NewReader(privateKey),
		keypair.P256,
	)
	if err != nil {
		panic(err)
	}
	address := types.AddressFromPubKey(pub)

	ethAddr := keypair.GetEthAddressFromPrivateKey(pri)

	return &Account{
		PrivateKey: pri,
		PublicKey:  pub,
		Address:    address,
		EthAddress: ethAddr,
		SigScheme:  s.SHA256withECDSA,
	}

}

//AccountMetadata all account info without private key
type AccountMetadata struct {
	IsDefault  bool   //Is default account
	Label      string //Lable of account
	KeyType    string //KeyType ECDSA,SM2 or EDDSA
	Curve      string //Curve of key type
	Address    string //Address(base58) of account
	EthAddress string //Ethereum address
	PubKey     string //Public  key
	SigSch     string //Signature scheme
	Salt       []byte //Salt
	Key        []byte //PrivateKey in encrypted
	EthKey     []byte // Ethereum PrivateKey
	EncAlg     string //Encrypt alg of private key
	Hash       string //Hash alg
}

func CheckKeyTypeCurve(keyType keypair.KeyType, curveCode byte) bool {
	switch keyType {
	case keypair.PK_ECDSA:
		switch curveCode {
		case keypair.P224:
		case keypair.P256:
		default:
			return false
		}
	case keypair.PK_SM2:
		switch curveCode {
		case keypair.SM2P256V1:
		default:
			return false
		}
	case keypair.PK_EDDSA:
		switch curveCode {
		case keypair.ED25519:
		default:
			return false
		}
	}
	return true
}

func CheckSigScheme(keyType keypair.KeyType, sigScheme s.SignatureScheme) bool {
	switch keyType {
	case keypair.PK_ECDSA:
		switch sigScheme {
		case s.SHA224withECDSA:
		case s.SHA256withECDSA:
		case s.SHA384withECDSA:
		case s.SHA512withECDSA:
		case s.SHA3_224withECDSA:
		case s.SHA3_256withECDSA:
		case s.SHA3_384withECDSA:
		case s.SHA3_512withECDSA:
		case s.RIPEMD160withECDSA:
		default:
			return false
		}
	case keypair.PK_SM2:
		switch sigScheme {
		case s.SM3withSM2:
		default:
			return false
		}
	case keypair.PK_EDDSA:
		switch sigScheme {
		case s.SHA512withEDDSA:
		default:
			return false
		}
	default:
		return false
	}
	return true
}

func GetKeyTypeString(keyType keypair.KeyType) string {
	switch keyType {
	case keypair.PK_ECDSA:
		return "ECDSA"
	case keypair.PK_SM2:
		return "SM2"
	case keypair.PK_EDDSA:
		return "Ed25519"
	default:
		return "unknown key type"
	}
}
