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

package errors

import (
	"bytes"
	"fmt"
	"runtime"
)

type CallStacker interface {
	GetCallStack() *CallStack
}

type CallStack struct {
	Stacks []uintptr
}

func GetCallStacks(err error) *CallStack {
	if err, ok := err.(CallStacker); ok {
		return err.GetCallStack()
	}
	return nil
}

func CallStacksString(call *CallStack) string {
	buf := bytes.Buffer{}
	if call == nil {
		return "No call stack available"
	}

	for _, stack := range call.Stacks {
		f := runtime.FuncForPC(stack)
		file, line := f.FileLine(stack)
		buf.WriteString(fmt.Sprintf("%s:%d - %s\n", file, line, f.Name()))
	}

	return fmt.Sprintf("%s", buf.Bytes())
}

func getCallStack(skip int, depth int) *CallStack {
	stacks := make([]uintptr, depth)
	stacklen := runtime.Callers(skip, stacks)

	return &CallStack{
		Stacks: stacks[:stacklen],
	}
}
