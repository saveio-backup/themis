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

import "testing"

func constructEventTimer() *EventTimer {
	server := constructServer()
	return NewEventTimer(server)
}

func TestStartTimer(t *testing.T) {
	eventtimer := constructEventTimer()
	eventtimer.StartTimer(1, 10)
}

func TestCancelTimer(t *testing.T) {
	eventtimer := constructEventTimer()
	eventtimer.StartTimer(1, 10)
	eventtimer.CancelTimer(1)
}
func TestStartEventTimer(t *testing.T) {
	eventtimer := constructEventTimer()
	err := eventtimer.startEventTimer(EventProposeBlockTimeout, 1)
	t.Logf("TestStartEventTimer: %v", err)
}

func TestCancelEventTimer(t *testing.T) {
	eventtimer := constructEventTimer()
	err := eventtimer.startEventTimer(EventProposeBlockTimeout, 1)
	t.Logf("startEventTimer: %v", err)
	eventtimer.cancelEventTimer(EventProposeBlockTimeout, 1)
}
