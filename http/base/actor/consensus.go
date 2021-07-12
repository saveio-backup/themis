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

package actor

import (
	"github.com/ontio/ontology-eventbus/actor"
	cactor "github.com/saveio/themis/consensus/actor"
)

var consensusSrvPid *actor.PID

func SetConsensusPid(actr *actor.PID) {
	consensusSrvPid = actr
}

//start consensus to consensus actor
func ConsensusSrvStart() error {
	if consensusSrvPid != nil {
		consensusSrvPid.Tell(&cactor.StartConsensus{})
	}
	return nil
}

//halt consensus to consensus actor
func ConsensusSrvHalt() error {
	if consensusSrvPid != nil {
		consensusSrvPid.Tell(&cactor.StopConsensus{})
	}
	return nil
}
var pocConsensusSrvPid *actor.PID

func SetPoCConsensusPid(actr *actor.PID) {
	pocConsensusSrvPid = actr
}

func RemovePlotFile(plotfile string) error {
	if pocConsensusSrvPid != nil {
		pocConsensusSrvPid.Tell(&cactor.PlotFileAction{
			PlotFile: plotfile,
		})
	}
	return nil
}

func SetSipVoteInfo(sipIndex uint32, agree byte) error {
	if pocConsensusSrvPid != nil {
		pocConsensusSrvPid.Tell(&cactor.SipVoteDecision{
			SipIndex: sipIndex,
			Agree:    agree,
		})
	}
	return nil
}

func SetConsVoteInfo(nodesPubkey []string) error {
	if pocConsensusSrvPid != nil {
		pocConsensusSrvPid.Tell(&cactor.ConsVoteDecision{
			NodesPubkey: nodesPubkey,
		})
	}
	return nil
}

func TriggerConsElect() error {
	if pocConsensusSrvPid != nil {
		pocConsensusSrvPid.Tell(&cactor.TriggerConsElect{})
	}
	return nil
}
