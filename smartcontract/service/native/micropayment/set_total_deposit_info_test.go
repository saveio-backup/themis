package micropayment

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/saveio/themis/common"
)

func TestSetTotalDepositInfo_Serialize(t *testing.T) {
	fileHashStr := []byte("QWERTYUIOdfghjokmVBwsxFGHJKLrdfghnkmlHYUHhjkjklljkljklyufsfwwwf")

	var addr, arr1 common.Address
	copy(addr[:], fileHashStr[0:20])
	//copy(addr1[:], fileHashStr[20:40])
	depositleger := SetTotalDepositInfo{
		ChannelID:             110,
		ParticipantWalletAddr: addr,
		PartnerWalletAddr:     arr1,
		SetTotalDeposit:       100,
	}
	bf := new(bytes.Buffer)
	err := depositleger.Serialize(bf)
	if err != nil {
		t.Error(err.Error())
	}
	var depositleger2 *SetTotalDepositInfo
	err = depositleger2.Deserialize(bf)
	if err != nil {
		t.Error(err.Error())
	}

	fmt.Println(depositleger2.SetTotalDeposit)
	fmt.Println(depositleger2.PartnerWalletAddr)
	fmt.Println(depositleger2.ParticipantWalletAddr)
	fmt.Println(depositleger2.ChannelID)
}
