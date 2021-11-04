// This file describes various headers of the protocol and how to use them
package protocol

type Header string

// Headers

//// In the following examples "~" is the HEADERDELIMETER

// ENCRKEY.
// The FIRST header to be sent. Sent immediately after the connection has been established
// by sender. Body contains a size of a key and the key itself.
// ie: ENCRKEY~(size)(SUPER SECURE ENCRYPTION KEY)
const HeaderEncryptionKey Header = "ENCRKEY"

// REJECT.
// Sent only by receiver if the receiver has decided to not download the contents.
// The body must contain a file ID in binary.
// ie: REJECT~1111011
const HeaderReject Header = "REJECT"

// ACCEPT.
// The opposite of the previous REJECT. Sent by receiver when
// he has agreed to download the file|directory. The body must contain
// the ID of a file in binary that is allowed to upload
// ie: ACCEPT~1111011
const HeaderAccept Header = "ACCEPT"

// DONE.
// Sent by sender. Warns the receiver that the transfer has been done and
// there is no more information to give.
// ie: DONE~
// Usually after the packet with this header has been sent, the receiver will send
// another packet back with header BYE!, telling that it`s going to disconnect
const HeaderDone Header = "DONE"

// READY.
// Sent by receiver when it has read and processed the last
// FILEBYTES packet. The sender is not allowed to "spam" FILEBYTES
// packets without the permission of receiver.
// ie: READY!~
const HeaderReady Header = "READY"

// BYE!.
// Packet with this header can be sent both by receiver and sender.
// It`s used when the sender or the receiver are going to disconnect
// and will not be able to communicate.
// (Usually it`s when the error has happened OR after the DONE header
// has been sent by sender, warning receiver that there is no data to send)
// The BODY is better to be empty.
// ie: BYE!~
const HeaderDisconnecting Header = "BYE!"

// FILE.
// Sent by sender, indicating that the file is going to be sent.
// The body structure must follow such structure:
// FILE~(idInBinary)(filenameLengthInBinary)(filename)(filesize)(checksumLengthInBinary)checksum
const HeaderFile Header = "FILE"

// FILEBYTES.
// Sent only by sender. The packet`s body must contain
// a file`s Identifier and a portion of its bytes.
// ie: FILEBYTES~(fileIDinBinary)(File`sBinaryData)
const HeaderFileBytes Header = "FILEBYTES"

// ENDFILE
// Sent by sender when the file`s contents fully has been sent.
// The body must contain a file ID.
// ie: ENDFILE~(fileIDIinBinary)
const HeaderEndfile Header = "ENDFILE"

// DIRECTORY. (TODO)
const HeaderDirectory Header = "DIRECTORY"
