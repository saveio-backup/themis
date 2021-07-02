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
package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func logPrint() {
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")
	Fatal("fatal")
	Trace("trace")

	testValue := 1
	Debugf("debug %v", testValue)
	Infof("info %v", testValue)
	Warnf("warn %v", testValue)
	Errorf("error %v", testValue)
	Fatalf("fatal %v", testValue)
	Tracef("trace %v", testValue)
}

func TestLog(t *testing.T) {
	defer func() {
		os.RemoveAll("Log/")
	}()

	InitLog(InfoLog, PATH, Stdout)
	Log().SetDebugLevel(DebugLog)
	logPrint()

	Log().SetDebugLevel(WarnLog)

	logPrint()

	err := ClosePrintLog()
	assert.Nil(t, err)
}

func TestNewLogFile(t *testing.T) {
	defer func() {
		os.RemoveAll("Log/")
	}()
	InitLog(InfoLog, PATH, Stdout)
	logfileNum1, err1 := ioutil.ReadDir("Log/")
	if err1 != nil {
		fmt.Println(err1)
		return
	}
	logPrint()
	ClosePrintLog()
	time.Sleep(time.Second * 2)
	InitLog(InfoLog, PATH, Stdout)
	logfileNum2, err2 := ioutil.ReadDir("Log/")
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	assert.Equal(t, len(logfileNum1), len(logfileNum2)-1)
}

func TestDataRace(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			InitLog(DebugLog)
		}
	}()

	for i := 0; i < 10; i++ {
		Debug("aaaa")
	}

	wg.Wait()
}
