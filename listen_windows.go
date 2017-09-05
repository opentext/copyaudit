/*
Copyright © 2017 OpenText Corp.

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice (including the next
paragraph) shall be included in all copies or substantial portions of the
Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
DEALINGS IN THE SOFTWARE.
*/

// +build !windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

var kernel = syscall.MustLoadDLL("kernel32.dll")
var (
	CreateNamedPipe  = kernel.MustFindProc("CreateNamedPipeW")
	ConnectNamedPipe = kernel.MustFindProc("ConnectNamedPipe")
)

const pipename = "\\\\.\\pipe\\Exceed TurboX Copy Audit"

func serve() {
	const inbound = 1   // PIPE_ACCESS_INBOUND
	const localonly = 8 // PIPE_REJECT_REMOTE_CLIENTS
	const unlim = 255   // PIPE_UNLIMITED_INSTANCES
	name, _ := syscall.UTF16PtrFromString(pipename)
	conn := 0
	for {
		listen, _, err := CreateNamedPipe.Call(uintptr(unsafe.Pointer(name)), inbound, localonly, unlim, 0, 0, 0, 0)
		if syscall.Handle(listen) == syscall.InvalidHandle {
			msg <- err.Error()
			msg <- "quit"
			break
		}

		success, _, err := ConnectNamedPipe.Call(listen, 0)
		if success == 0 {
			msg <- err.Error()
			msg <- "quit"
			break
		}
		conn++
		go handle(os.NewFile(listen, pipename), conn)
	}
}

func listenAndServe() error {
	go serve()
	return nil
}
