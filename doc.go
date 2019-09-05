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

Protocol

The copy audit daemon listens on the abstract socket (Linux) or named pipe
(Windows) "Exceed TurboX Copy Audit". On other systems, it listens on the Unix
socket "/tmp/.X11-unix/ETXaudit"

When ETX opens a socket, it will write the 8-byte string "ETXaudit" followed by
the byte 0x9F (aka "CBOR start indefinite length array").  The remainder of the
protocol is based on CBOR. For more details on CBOR encoding, see
http://cbor.io and/or RFC 7049.

Each audit/copy event is a single CBOR Map. The Map may contain any combination
of the following entries (with CBOR type in brackets), except that it shall not
contain more than one of Text, Image, FileStart, FileComplete, or Print.
Additional entries not specified here may be present.

 - Display (Unsigned) is the display number of the ETX proxy.

 - File (String) is the name of the file copied by the user.

 - FileComplete (bool) is true if the file transferred from the proxy to the
   desktop, and false if the file transferred from the desktop to the proxy.

 - FileStart (bool) is true if the file will transfer from the proxy to the
   desktop, and false if the file will transfer from the desktop to the proxy.

 - FileSize (Unsigned) is the length (in bytes) of the file copied by the user.

 - Image (Binary) is the image copied by the user.

 - ImageType (String) is the name of the X11 protocol atom describing the image
   (nominally the image's MIME type).

 - IPAddress (String, or Array of String) is the IP address of the remote
   user's computer (not the ETX proxy) or computers (when sharing).

 - Print (Binary) is the first part of the document printed by the user
   (limited to proxy.CopyAuditPrintLimit bytes).

 - TransferIPAddress (String) is the IP address of the file transfer computer
   (not the user's desktop or the ETX proxy). This field is not present when the
   file transfer is exchanged with the ETX proxy itself.

 - Text (Binary) is the text copied. It is not a CBOR String because it may be
   any text type supported by X.

 - User (String) is the username of the user.

 - XApp (String) is the name of the ETX profile the user is running.

The ETX proxy will send Text, Image, FileStart, FileComplete, or Print last in
the Map, so the other fields can inform the disposition of the Text, Image,
FileStart, FileComplete, or Print.

There is no data sent from the audit daemon to ETX. ETX does not want to wait
for confirmation that the audit message has been logged. Attempting to write
data back to ETX may block (because ETX will never read it).

*/
package main
