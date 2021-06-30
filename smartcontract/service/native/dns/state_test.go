package dns

import (
	"bytes"
	"testing"

	"github.com/saveio/themis/common"
	"github.com/stretchr/testify/assert"
)

func TestName_Serialize_Deserialize(t *testing.T) {
	owner, _ := common.AddressFromHexString("AXxzYEV95ub7Nx32k3JnbCNatZNidvcA1L")
	name := NameInfo{
		Header:      []byte{31, 32, 33, 34, 35},
		URL:         []byte{31, 32, 33, 34, 35, 36, 30},
		Name:        []byte("AUbmcsdlcm,m.123lmsn"),
		NameOwner:   owner,
		Desc:        []byte("hcdkshfiyslfs.f"),
		BlockHeight: 2434543,
		TTL:         102324,
	}
	bf := new(bytes.Buffer)
	if err := name.Serialize(bf); err != nil {
		t.Fatalf("NameInfo serialize error: %v", err)
	}
	deserializeName := NameInfo{}
	if err := deserializeName.Deserialize(bf); err != nil {
		t.Fatalf("NameInfo deserialize error: %v", err)
	}

	if string(name.Header) != string(deserializeName.Header) || string(name.URL) != string(deserializeName.URL) ||
		string(name.Name) != string(deserializeName.Name) || name.NameOwner.ToHexString() != deserializeName.NameOwner.ToHexString() ||
		string(name.Desc) != string(deserializeName.Desc) || string(name.BlockHeight) != string(deserializeName.BlockHeight) ||
		string(name.TTL) != string(deserializeName.TTL) {
		t.Fatal("NameInfo deserialize error")
	}

}

func TestHeader_Serialize_Deserialize(t *testing.T) {
	header := HeaderInfo{}
	bf := new(bytes.Buffer)
	err := header.Serialize(bf)
	assert.Nil(t, err)
	deserializeHeaderInfor := HeaderInfo{}
	err = deserializeHeaderInfor.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, header, deserializeHeaderInfor)
}

func TestRequestName_Serialize_Deserialize(t *testing.T) {
	owner, _ := common.AddressFromHexString("AXxzYEV95ub7Nx32k3JnbCNatZNidvcA1L")
	name := RequestName{
		Type:      8,
		Header:    []byte{31, 32, 33, 34, 35},
		URL:       []byte{31, 32, 33, 34, 35, 36, 30},
		Name:      []byte("AUbmcsdlcm,m.123lmsn"),
		NameOwner: owner,
		Desc:      []byte("hcdkshfiyslfs.f"),
		DesireTTL: 102324,
	}
	bf := new(bytes.Buffer)
	err := name.Serialize(bf)
	assert.Nil(t, err)
	deserializeName := RequestName{}
	err = deserializeName.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, name, deserializeName)
}

func TestRequestHeader_Serialize_Deserialize(t *testing.T) {
	owner, _ := common.AddressFromHexString("AXxzYEV95ub7Nx32k3JnbCNatZNidvcA1L")
	header := RequestHeader{
		Header:    []byte{31, 32, 33, 34, 35},
		NameOwner: owner,
		Desc:      []byte("hcdkshfiyslfs.f"),
		DesireTTL: 102324,
	}
	bf := new(bytes.Buffer)
	err := header.Serialize(bf)
	assert.Nil(t, err)
	deserializeHeader := RequestHeader{}
	err = deserializeHeader.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, header, deserializeHeader)
}

func TestReqInfo_Serialize_Deserialize(t *testing.T) {
	owner, _ := common.AddressFromHexString("AXxzYEV95ub7Nx32k3JnbCNatZNidvcA1L")
	req := ReqInfo{
		Header: []byte{31, 32, 33, 34, 35},
		URL:    []byte{31, 32, 33, 34, 35, 36, 30},
		Owner:  owner,
	}
	bf := new(bytes.Buffer)
	err := req.Serialize(bf)
	assert.Nil(t, err)
	deserializeReq := ReqInfo{}
	err = deserializeReq.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, req, deserializeReq)
}

func TestTranferInfo_Serialize_Deserialize(t *testing.T) {
	from, _ := common.AddressFromHexString("AXxzYEV95ub7Nx32k3JnbCNatZNidvcA1L")
	to, _ := common.AddressFromHexString("AXYzYEV11ub7Nx32k3JnbCNatZNidvcmgs")
	ti := TranferInfo{
		Header: []byte{31, 32, 33, 34, 35},
		URL:    []byte{31, 32, 33, 34, 35, 36, 30},
		From:   from,
		To:     to,
	}
	bf := new(bytes.Buffer)
	err := ti.Serialize(bf)
	assert.Nil(t, err)
	deserializeTi := TranferInfo{}
	err = deserializeTi.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, ti, deserializeTi)
}
