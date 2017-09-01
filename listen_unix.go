// +build !windows

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

package main

import (
	"net"
	"runtime"
)

func serve(listen net.Listener) {
	conn := 0
	for {
		c, err := listen.Accept()
		if err != nil {
			msg <- err.Error()
			msg <- "quit"
			return
		}
		conn++
		go handle(c, conn)
	}
}

func listenAndServe() error {
	name := "/tmp/.X11-unix/ETXaudit"
	if runtime.GOOS == "linux" {
		name = "@Exceed TurboX Copy Audit"
	}
	listen, err := net.Listen("unix", name)
	if err != nil {
		return err
	}
	go serve(listen)
	return nil
}
