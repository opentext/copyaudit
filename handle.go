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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func handle(remote io.ReadCloser, cno int) {
	defer remote.Close()

	msg <- fmt.Sprintf("New connection #%d", cno)
	prefix := fmt.Sprintf("%d| ", cno)
	defer func() { msg <- prefix + "connection closed" }()

	header := []byte("ETXaudit\x9F")
	hdr := make([]byte, len(header))

	r := bufio.NewReader(remote)
	io.ReadFull(r, hdr)
	if !bytes.Equal(hdr, header) {
		msg <- "Attempted connection was not an ETXaudit connection"
		return
	}

	msg <- prefix + "connection accepted"
	for {
		itms, err := cborRead(r)
		if err != nil {
			msg <- prefix + err.Error()
			return
		}
		if itms.typ != cborMap {
			msg <- fmt.Sprintf("%sprotocol error: got %s expected cbor map", prefix, itms)
			return
		}

		app := ""
		dpy := ""
		ip := ""
		text := ""
		user := ""
		ext := "jpeg"

		for i := uint64(0); i < itms.value; i++ {
			key, err := cborReadString(r)
			if err != nil {
				msg <- fmt.Sprintf("%sprotocol error: got %s expected string key", prefix, err)
				return
			}
			switch key {
			case "Display":
				dpyno, err := cborRead(r)
				if err != nil {
					msg <- err.Error()
					return
				}
				dpy = fmt.Sprintf("dpy :%d", dpyno.value)
			case "User":
				s, err := cborReadString(r)
				if err == nil {
					user = s
				}
			case "XApp":
				s, err := cborReadString(r)
				if err == nil {
					app = "app " + s
				}
			case "IPAddress":
				s, err := cborReadString(r)
				if err == nil {
					ip = "ip " + s
				}
			case "ImageType":
				if s, err := cborReadString(r); err == nil {
					pos := strings.LastIndexAny(s, "/\\")
					if pos >= 0 {
						ext = s[pos+1:]
					}
				}
			case "Text":
				tok, err := cborRead(r)
				if err != nil {
					msg <- err.Error()
					return
				}
				if tok.typ != cborBinary {
					cborDiscardBody(r, tok)
					continue
				}
				if tok.value > 0x7FFFFFFF {
					msg <- fmt.Sprintf("%slog length %d too large", prefix, tok.value)
					return
				}
				discard := 0
				// Limit text to 1 kB
				if tok.value > 1024 {
					discard = int(tok.value) - 1024
					tok.value = 1024
				}
				blob := make([]byte, int(tok.value))
				io.ReadFull(r, blob)
				r.Discard(discard)
				text = fmt.Sprintf("copy %q", blob)
			case "Image":
				tok, err := cborRead(r)
				if err != nil {
					msg <- err.Error()
					return
				}
				if tok.typ != cborBinary {
					cborDiscardBody(r, tok)
					continue
				}
				if tok.value > 0x7FFFFFFF {
					msg <- fmt.Sprintf("%simage size %d too large", prefix, tok.value)
					return
				}

				path := *dir
				if user != "" {
					path = filepath.Join(path, filepath.Clean(user))
					os.Mkdir(path, 0640)
				}
				fn := fmt.Sprintf("%d.%s", time.Now().UnixNano(), ext)
				path = filepath.Join(path, fn)

				jpeg, err := os.Create(path)
				if err != nil {
					msg <- fmt.Sprintf("%serror creating %s: %s", prefix, path, err.Error())
				} else {
					_, err = io.Copy(jpeg, io.LimitReader(r, int64(tok.value)))
					jpeg.Close()
				}
				if err != nil {
					msg <- fmt.Sprintf("%serror writing %s: %s", prefix, path, err.Error())
					return
				}
				text = "copy " + path
			}
		}
		var out []string
		if user != "" {
			out = append(out, "user "+user)
		}
		if dpy != "" {
			out = append(out, dpy)
		}
		if ip != "" {
			out = append(out, ip)
		}
		if app != "" {
			out = append(out, app)
		}
		if text != "" {
			out = append(out, text)
		}
		msg <- prefix + strings.Join(out, ", ")
	}
}
