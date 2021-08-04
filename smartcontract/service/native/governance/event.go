package governance

import (
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/event"
	"github.com/saveio/themis/smartcontract/service/native"
)

const (
	EVENT_SIP_REGISTER = iota + 1
	EVENT_VOTEE_CONS_NODE
	EVENT_PLEDGE_FOR_CONS
	EVENT_ENSURE_SPACE_FOR_MINING
)

type sipRegisterEvent struct {
	Index    uint32
	Height   uint32
	Detail   []byte
	Default  byte
	MinVotes uint32
	Bonus    uint64
}

type voteeConsNodesEvent struct {
	ConsGovPeriod uint32
	Votees        []string
}

type pledgeForConsEvent struct {
	ConsGovPeriod uint32
}

func newEvent(srvc *native.NativeService, id uint32, st interface{}) {
	e := event.NotifyEventInfo{}
	e.ContractAddress = srvc.ContextRef.CurrentContext().ContractAddress
	e.EventIdentifier = id
	e.Addresses = []common.Address{}
	e.States = st
	srvc.Notifications = append(srvc.Notifications, &e)
	return
}

func SipRegisteredEvent(native *native.NativeService, sipRegister *sipRegisterEvent) {
	event := map[string]interface{}{
		"blockHeight":     native.Height,
		"eventName":       "sipRegistered",
		"effectiveHeight": sipRegister.Height,
		"sipIndex":        sipRegister.Index,
		"detail":          sipRegister.Detail,
		"default":         sipRegister.Default,
		"minVotes":        sipRegister.MinVotes,
		"bonus":           sipRegister.Bonus,
	}
	newEvent(native, EVENT_SIP_REGISTER, event)
}

func VoteeConsNodesEvent(native *native.NativeService, votee *voteeConsNodesEvent) {
	event := map[string]interface{}{
		"blockHeight":   native.Height,
		"eventName":     "voteeConsNode",
		"consGovPeriod": votee.ConsGovPeriod,
		"votees":        votee.Votees,
	}
	newEvent(native, EVENT_VOTEE_CONS_NODE, event)
}

func PledgeForConsEvent(native *native.NativeService, pledge *pledgeForConsEvent) {
	event := map[string]interface{}{
		"blockHeight":   native.Height,
		"eventName":     "pledgeForConsensus",
		"consGovPeriod": pledge.ConsGovPeriod,
	}
	newEvent(native, EVENT_PLEDGE_FOR_CONS, event)
}
