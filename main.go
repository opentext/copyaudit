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
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var dir = flag.String("dir", "", "root directory of audit log")

var msg = make(chan string)

func main() {
	flag.Parse()
	if *dir == "" {
		log.Fatal("You must specify -dir=<audit directory>")
	}
	dirh, err := os.Open(*dir)
	if err != nil {
		log.Println("Cannot open directory `" + *dir + "` for logging")
		log.Fatal(err)
	}
	s, _ := dirh.Stat()
	if !s.IsDir() {
		log.Fatal("Cannot open directory `" + *dir + "` for logging")
	}
	dirh.Close()

	hup := make(chan os.Signal, 2)
	signal.Notify(hup, syscall.SIGHUP)
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, os.Kill)

	err = listenAndServe()
	if err != nil {
		log.Println("Cannot create listening socket")
		log.Fatal(err)
	}

	logfn := filepath.Join(*dir, "copyaudit.log")
	logfile, err := os.OpenFile(logfn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		log.Println("Cannot create log file")
		log.Fatal(err)
	}
	l := log.New(logfile, "", log.Ldate|log.Ltime|log.LUTC)
	l.Println("Starting audit log daemon")

	for {
		select {
		case m := <-msg:
			if m == "quit" {
				// The serve routine already sent us the reason, no need for an extra log message
				os.Exit(1)
			}
			l.Println(m)
		case <-hup:
			l.Println("SIGHUP - reopening log file")
			newlogfile, err := os.OpenFile(logfn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
			if err != nil {
				log.Println("Cannot reopen log file")
				log.Println(err.Error())
				continue
			}
			logfile.Close()
			logfile = newlogfile
			l.SetOutput(logfile)
			l.Println("SIGHUP - reopened log file")
		case sig := <-quit:
			l.Println("Received", sig)
			l.Println("Exiting")
			os.Exit(0)
		}
	}
}
