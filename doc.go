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

/*
copyaudit

This is an example of a daemon that listens for a stream of copy-and-paste
audit messages from OpenText Exceed TurboX. This example daemon writes each
image to its own file and logs all text copy messages to a single file. This
program is not supported in any way, shape, or form, but is provided as an
example for the end-user to write their own daemon.

Set proxy.CopyAudit=2 in the ETX configuration to enable optional copy
auditing. Set proxy.CopyAudit=1 to cause ETX to exit if it cannot connect to
the copy audit daemon. Additionally set proxy.CopyAuditImage=1 to enable "Copy
Rectangle" image auditing. This is an undocumented feature of ETX. As such, the
flag and the protocol are subject to change at any time without notice.

Protocol (as supported by this example):
 The copy audit daemon listens on the abstract socket (Linux) or named pipe
 (Windows) "Exceed TurboX Copy Audit". On other systems, it listens on the Unix
 socket "/tmp/.X11-unix/ETXaudit"

 When ETX opens a socket, it will write the 8-byte string "ETXaudit" followed
 by the byte 0x9F (aka "CBOR start indefinite length array").

 The remainder of the protocol is based on CBOR. For more details on CBOR
 encoding, see http://cbor.io and/or RFC 7049.

 Each audit/copy event is a single CBOR Map. It will typically contain the
 field "Display", an unsigned number, and the fields "XApp", "User", and
 "IPAddress", all three of which (if present) are strings. The IPAddress field
 is the address of the remote user, not the address of the ETX proxy machine.
 The event may also contain one of "Text" or "Image". Currently, Image is
 always in JPEG format.  Text is sent as a CBOR "byte string" (not CBOR "text
 string") because it may be in any of the encodings supported by ICCCM, and is
 not necessarily UTF-8.

 There is no data sent from the audit daemon to ETX. ETX does not want to wait
 for confirmation that the audit message has been logged. Attempting to write
 data back to ETX may block (because ETX will never read it).

*/
package main
