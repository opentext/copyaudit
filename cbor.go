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

// one could use a fancy CBOR parsing library such as the excellent
// github.com/ugorji/go/codec instead. This simple parser is only used to avoid
// external dependencies

package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type cborObject struct {
	typ   uint8
	value uint64
}

func (c cborObject) String() string {
	var typStr = []string{"Unsigned", "Negative", "Binary", "String", "Array", "Map", "Tag", "Misc", "Break"}
	rv := ""
	if int(c.typ) < len(typStr) {
		rv = typStr[int(c.typ)]
	} else {
		rv = fmt.Sprintf("cbor(%d)", c.typ)
	}
	return fmt.Sprintf("%s(%d)", rv, c.value)
}

const (
	// Types defined by the cbor spec
	cborUnsigned = 0
	cborNegative = 1
	cborBinary   = 2
	cborString   = 3
	cborArray    = 4
	cborMap      = 5
	cborTag      = 6
	cborMisc     = 7
	// Types added for convenience (that cannot be encoded by the cbor wire format)
	cborBreak = 8
)

func cborRead(r *bufio.Reader) (c cborObject, err error) {
	t, err := r.ReadByte()
	c.typ = t >> 5
	v := t & 0x1F
	if v < 24 {
		c.value = uint64(v)
	} else if v == 24 {
		var s byte
		s, err = r.ReadByte()
		c.value = uint64(s)
	} else if v == 25 {
		var tmp [2]byte
		_, err = io.ReadFull(r, tmp[:])
		c.value = uint64(binary.BigEndian.Uint16(tmp[:]))
	} else if v == 26 {
		var tmp [4]byte
		_, err = io.ReadFull(r, tmp[:])
		c.value = uint64(binary.BigEndian.Uint32(tmp[:]))
	} else if v == 27 {
		var tmp [8]byte
		_, err = io.ReadFull(r, tmp[:])
		c.value = binary.BigEndian.Uint64(tmp[:])
	} else if v == 31 {
		if c.typ == cborMisc {
			c.typ = cborBreak
		} else {
			err = errors.New("Indefinite CBOR types not supported")
			return
		}
	} else {
		err = errors.New("Invalid start of CBOR object")
		return
	}
	return
}

func cborReadString(r *bufio.Reader) (string, error) {
	c, err := cborRead(r)
	if err != nil {
		return "", err
	}
	if c.typ != cborString {
		err := cborDiscardBody(r, c)
		if err == nil {
			err = errors.New("Expected String, got " + c.String())
		}
		return "", err
	}

	buf := make([]byte, int(c.value))
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func cborDiscard(r *bufio.Reader) (err error) {
	c, err := cborRead(r)
	if err != nil {
		return err
	}
	return cborDiscardBody(r, c)
}

func cborDiscardBody(r *bufio.Reader, c cborObject) error {
	switch c.typ {
	case cborUnsigned, cborNegative, cborMisc:
		return nil
	case cborBinary, cborString:
		_, err := r.Discard(int(c.value))
		return err
	case cborArray:
		for i := uint64(0); i < c.value; i++ {
			err := cborDiscard(r)
			if err != nil {
				return err
			}
		}
		return nil
	case cborMap:
		for i := uint64(0); i < c.value; i++ {
			err1 := cborDiscard(r)
			err2 := cborDiscard(r)
			if err1 != nil {
				return err1
			}
			if err2 != nil {
				return err2
			}
		}
		return nil
	case cborTag:
		return cborDiscard(r)
	}
	return errors.New("internal error " + c.String())
}
