package micropayment

import (
	"encoding/hex"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/event"
	"github.com/saveio/themis/smartcontract/service/native"
)

const (
	EVENT_CHANNEL_OPENED = iota + 1
	EVENT_CHANNEL_CLOSED
	EVENT_CHANNEL_SETTLED
	EVENT_CHANNEL_COSETTLED
	EVENT_CHANNEL_UNLOCKED
	EVENT_BALANCE_PROOF_UPDATE
	EVENT_SET_DEPOSIT
	EVENT_WITHDRAW
	EVENT_SECRET_REVEALED
	EVENT_FAST_TRANSFER
	EVENT_SET_FEE
)

type channelOpenedEvent struct {
	channelIdentifier uint64
	participant1      common.Address
	participant2      common.Address
	settleTimeout     uint64
}

type channelNewDepositEvent struct {
	channelIdentifier uint64
	participant       common.Address
	totalDeposit      uint64
}

type channelCloseEvent struct {
	channelID          uint64
	closingParticipant common.Address
	nonce              uint64
}

type channelSettledEvent struct {
	channelID           uint64
	participant1_amount uint64
	participant2_amount uint64
}

type channelWithdrawEvent struct {
	channelIdentifier uint64
	participant       common.Address
	totalWithdraw     uint64
}

type unlockEvent struct {
	channelIdentifier uint64
	participant       common.Address
	partner           common.Address
	computedLocksroot [32]byte
	unlockedAmount    uint64
	returnedTokens    uint64
}

type fastTransferEvent struct {
	paymentId uint64
	asset     common.Address
	from      common.Address
	to        common.Address
	amount    uint64
}

func newEvent(srvc *native.NativeService, id uint32, participants []common.Address, st interface{}) {
	e := event.NotifyEventInfo{}
	e.ContractAddress = srvc.ContextRef.CurrentContext().ContractAddress
	e.EventIdentifier = id
	e.Addresses = append(e.Addresses, participants...)
	e.States = st
	srvc.Notifications = append(srvc.Notifications, &e)
	return
}

func SecretRevealedEvent(native *native.NativeService, secretHash [32]byte, secret [32]byte, height uint32) {
	event := map[string]interface{}{
		"blockHeight": height,
		"eventName":   "SecretRevealed",
		"secretHash":  secretHash,
		"secret":      secret,
	}
	newEvent(native, EVENT_SECRET_REVEALED, nil, event)
}

func triggerEndpointRegisterEvent(native *native.NativeService, ip, port []byte, walletAddr common.Address) {
	event := map[string]interface{}{
		"blockHeight": native.Height,
		"eventName":   "Register",
		"ip":          ip,
		"port":        port,
		"walletAddr":  walletAddr,
	}
	newEvent(native, 0, nil, event)
}

func ChannelWithdrawEvent(native *native.NativeService, withdrawEvent channelWithdrawEvent, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight":   native.Height,
		"eventName":     "SetTotalWithdraw",
		"channelID":     withdrawEvent.channelIdentifier,
		"participant":   withdrawEvent.participant,
		"totalWithdraw": withdrawEvent.totalWithdraw,
	}
	newEvent(native, EVENT_WITHDRAW, participants, event)
}

func ChannelSettledEvent(native *native.NativeService, chanSettledEvent channelSettledEvent, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight":         native.Height,
		"eventName":           "chanSettled",
		"channelID":           chanSettledEvent.channelID,
		"participant1_amount": chanSettledEvent.participant1_amount,
		"participant2_amount": chanSettledEvent.participant2_amount,
	}
	newEvent(native, EVENT_CHANNEL_SETTLED, participants, event)
}

func ChannelCooperativeSettledEvent(native *native.NativeService, chanSettledEvent channelSettledEvent, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight":         native.Height,
		"eventName":           "chanCooperativeSettled",
		"channelID":           chanSettledEvent.channelID,
		"participant1_amount": chanSettledEvent.participant1_amount,
		"participant2_amount": chanSettledEvent.participant2_amount,
	}
	newEvent(native, EVENT_CHANNEL_COSETTLED, participants, event)
}

func ChannelSetFeeEvent(native *native.NativeService, feeInfo FeeInfo, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight":   native.Height, //uint32
		"eventName":     "SetFee",
		"walletAddr":    feeInfo.WalletAddr,      //"github.com/saveio/themis/common"
		"tokenAddr":  	 feeInfo.TokenAddr,
		"flat": 	     feeInfo.Flat,
	}
	newEvent(native, EVENT_SET_FEE, participants, event)
}

func ChannelOpenedEvent(native *native.NativeService, chanOpened channelOpenedEvent, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight":   native.Height, //uint32
		"eventName":     "chanOpened",
		"channelID":     chanOpened.channelIdentifier, //uint64
		"participant1":  chanOpened.participant1,      //"github.com/saveio/themis/common"
		"participant2":  chanOpened.participant2,
		"settleTimeout": chanOpened.settleTimeout, //uint64
	}
	newEvent(native, EVENT_CHANNEL_OPENED, participants, event)
}

func ChannelNewDepositEvent(native *native.NativeService, chDeposit channelNewDepositEvent, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight":  native.Height,
		"eventName":    "SetTotalDeposit",
		"channelID":    chDeposit.channelIdentifier,
		"participant":  chDeposit.participant,
		"totalDeposit": chDeposit.totalDeposit,
	}
	newEvent(native, EVENT_SET_DEPOSIT, participants, event)
}

func ChannelCloseEvent(native *native.NativeService, chClose channelCloseEvent, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight":        native.Height,
		"eventName":          "ChannelClose",
		"channelID":          chClose.channelID,
		"nonce":              chClose.nonce,
		"closingParticipant": chClose.closingParticipant,
	}
	newEvent(native, EVENT_CHANNEL_CLOSED, participants, event)
}
func NonClosingBPFUpdateEvent(native *native.NativeService, chID uint64, closingP common.Address, nonce uint64, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight":        native.Height,
		"eventName":          "NonClosingBPFUpdate",
		"channelID":          chID,
		"nonce":              nonce,
		"closingParticipant": closingP,
	}
	newEvent(native, EVENT_BALANCE_PROOF_UPDATE, participants, event)
}

func topUpChannelEvent(native *native.NativeService, sender common.Address, op string, ChannelID []byte, amount uint64) {
	event := map[string]interface{}{
		"blockHeight": native.Height,
		"eventName":   "topUpChannel",
		"sender":      sender,
		"op":          op,
		"ChannelID":   ChannelID,
		"amount":      amount,
	}
	newEvent(native, 0, nil, event)
}

func UnlockEvent(native *native.NativeService, unlock unlockEvent, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight":       native.Height,
		"eventName":         "ChannelUnlocked",
		"channelID":         unlock.channelIdentifier,
		"participant":       unlock.participant,
		"partner":           unlock.partner,
		"computedLocksroot": unlock.computedLocksroot,
		"unlockedAmount":    unlock.unlockedAmount,
		"returnedTokens":    unlock.returnedTokens,
	}
	newEvent(native, EVENT_CHANNEL_UNLOCKED, participants, event)
}

func NewFastTransferEvent(native *native.NativeService, evt fastTransferEvent, participants []common.Address) {
	event := map[string]interface{}{
		"blockHeight": native.Height,
		"eventName":   "FastTransfer",
		"paymentId":   evt.paymentId,
		"asset":       evt.asset,
		"from":        evt.from,
		"to":          evt.to,
		"amount":      evt.amount,
	}
	newEvent(native, EVENT_FAST_TRANSFER, participants, event)
}

func triggerAttributeEvent(srvc *native.NativeService, op string, id []byte, path [][]byte) {
	var attr interface{}
	if op == "remove" {
		attr = hex.EncodeToString(path[0])
	} else {
		t := make([]string, len(path))
		for i, v := range path {
			t[i] = hex.EncodeToString(v)
		}
		attr = t
	}
	event := map[string]interface{}{
		"blockHeight": srvc.Height,
		"eventName":   "Attribute",
		"op":          op,
		"id":          string(id),
		"attr":        attr,
	}
	newEvent(srvc, 0, nil, event)
}
