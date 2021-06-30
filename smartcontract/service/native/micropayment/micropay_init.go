package micropayment

import (
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func InitMicroPay() {
	native.Contracts[utils.MicroPayContractAddress] = RegisterMPayContract

}

func RegisterMPayContract(native *native.NativeService) {
	native.Register(MP_OPEN_CHANNEL, OpenChannel)
	native.Register(MP_SET_TOTALDEPOSIT, SetTotalDeposit)
	native.Register(MP_SET_TOTALWITHDRAW, SetTotalWithdraw)
	native.Register(MP_COOPERATIVESETTLE, CooperativeSettle)
	native.Register(MP_CLOSE_CHANNEL, CloseChannel)
	native.Register(MP_UNLOCK, Unlock)
	native.Register(MP_SECRET_REG, RegisterSecret)
	native.Register(MP_SECRET_REG_BATCH, RegisterSecretBatch)
	native.Register(MP_GET_SECRET_REVEAL_BLOCKHEIGHT, GetSecretRevealBlockHeight)
	native.Register(MP_UPDATE_NONCLOSING_BPF, UpdateNonClosingBalanceProof)
	native.Register(MP_SETTLE_CHANNEL, SettleChannel)
	native.Register(MP_GET_CHANNELINFO, GetChannelInfo)
	native.Register(MP_GET_ALL_OPEN_CHANNELS, GetAllOpenChannels)
	native.Register(MP_GET_CHANNELCOUNTER, GetChannelCounter)
	native.Register(MP_GET_CHANNEL_PARTICIPANTINFO, GetChannelParticipantInfo)
	native.Register(MP_GET_CHANNELID, GetChannelIdentifier)
	native.Register(MP_GET_NODE_PUBKEY, GetNodePubKey)
	native.Register(MP_FAST_TRANSFER, FastTransfer)

}
