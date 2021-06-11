package sender

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/Unbewohnte/FTU/checksum"
	"github.com/Unbewohnte/FTU/protocol"
)

// gets a local ip. Borrowed from StackOverflow, thank you, whoever I brought it from
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

// gets a remote ip. Borrowed from StackOverflow, thank you, whoever I brought it from
func GetRemoteIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", fmt.Errorf("could not make a request to get your remote IP: %s", err)
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read a response: %s", err)
	}
	return string(ip), nil
}

// The main sender struct
type Sender struct {
	Port                 int
	FileToTransfer       *File
	Listener             net.Listener
	Connection           net.Conn
	IncomingPackets      chan protocol.Packet
	SentFileBytesPackets uint64
	TransferAllowed      bool
	ReceiverIsReady      bool
	Stopped              bool
}

// Creates a new sender with default fields
func NewSender(port int, filepath string) *Sender {
	fileToTransfer, err := getFile(filepath)
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	incomingPacketsChan := make(chan protocol.Packet, 5000)

	remoteIP, err := GetRemoteIP()
	if err != nil {
		panic(err)
	}
	localIP, err := GetLocalIP()
	if err != nil {
		panic(err)
	}

	var filepacketCounter uint64
	fmt.Printf("Created a new sender at %s:%d (remote)\n%s:%d (local)\n", remoteIP, port, localIP, port)
	return &Sender{
		Port:                 port,
		FileToTransfer:       fileToTransfer,
		Listener:             listener,
		Connection:           nil,
		IncomingPackets:      incomingPacketsChan,
		SentFileBytesPackets: filepacketCounter,
		TransferAllowed:      false,
		ReceiverIsReady:      false,
		Stopped:              false,
	}
}

// Closes the connection, warns about it the receiver and exits the mainloop
func (s *Sender) Stop() {
	disconnectionPacket := protocol.Packet{
		Header: protocol.HeaderDisconnecting,
	}
	err := protocol.SendPacket(s.Connection, disconnectionPacket)
	if err != nil {
		panic(fmt.Sprintf("could not send a disconnection packet: %s", err))
	}

	s.Stopped = true
	s.Disconnect()
}

// Closes current connection
func (s *Sender) Disconnect() {
	s.Connection.Close()
}

// Accepts one connection
func (s *Sender) WaitForConnection() {
	connection, err := s.Listener.Accept()
	if err != nil {
		fmt.Printf("Could not accept a connection: %s", err)
		os.Exit(-1)
	}
	s.Connection = connection
	fmt.Println("New connection from ", s.Connection.RemoteAddr())
}

// Closes the listener. Used only when there is still no connection from `AcceptConnections`
func (s *Sender) StopListening() {
	s.Listener.Close()
}

// Sends multiple packets with all information about the file to receiver
// (filename, filesize, checksum)
func (s *Sender) SendOffer() error {
	// filename
	filenamePacket := protocol.Packet{
		Header: protocol.HeaderFilename,
		Body:   []byte(s.FileToTransfer.Filename),
	}
	err := protocol.SendPacket(s.Connection, filenamePacket)
	if err != nil {
		return fmt.Errorf("could not send an information about the file: %s", err)
	}

	// filesize
	filesizePacket := protocol.Packet{
		Header: protocol.HeaderFileSize,
		Body:   []byte(strconv.Itoa(int(s.FileToTransfer.Filesize))),
	}

	err = protocol.SendPacket(s.Connection, filesizePacket)
	if err != nil {
		return fmt.Errorf("could not send an information about the file: %s", err)
	}

	// checksum
	checksumPacket := protocol.Packet{
		Header: protocol.HeaderChecksum,
		Body:   checksum.ChecksumToBytes(s.FileToTransfer.CheckSum),
	}
	err = protocol.SendPacket(s.Connection, checksumPacket)
	if err != nil {
		return fmt.Errorf("could not send an information about the file: %s", err)
	}

	// indicate that we`ve sent everything we needed to send
	donePacket := protocol.Packet{
		Header: protocol.HeaderDone,
	}
	err = protocol.SendPacket(s.Connection, donePacket)
	if err != nil {
		return fmt.Errorf("could not send an information about the file: %s", err)
	}
	return nil
}

// Sends one packet that contains a piece of file to the receiver
func (s *Sender) SendPiece() error {
	// if no data to send - exit
	if s.FileToTransfer.LeftBytes == 0 {
		fmt.Printf("Done. Sent %d file packets\n", s.SentFileBytesPackets)
		s.Stop()
	}

	// empty body
	fileBytesPacket := protocol.Packet{
		Header: protocol.HeaderFileBytes,
	}

	// how many bytes we can send at maximum
	maxFileBytes := protocol.MAXPACKETSIZE - uint(protocol.MeasurePacketSize(fileBytesPacket))

	fileBytes := make([]byte, maxFileBytes)
	// if there is less data to send than the limit - create a buffer of needed size
	if s.FileToTransfer.LeftBytes < uint64(maxFileBytes) {
		fileBytes = make([]byte, uint64(maxFileBytes)-(uint64(maxFileBytes)-s.FileToTransfer.LeftBytes))
	}

	// reading bytes from the point where we left
	read, err := s.FileToTransfer.Handler.ReadAt(fileBytes, int64(s.FileToTransfer.SentBytes))
	if err != nil {
		return fmt.Errorf("could not read from a file: %s", err)
	}

	// filling BODY with bytes
	fileBytesPacket.Body = fileBytes

	err = protocol.SendPacket(s.Connection, fileBytesPacket)
	if err != nil {
		return fmt.Errorf("could not send a file packet : %s", err)
	}

	// doing a "logging" for the next time
	s.FileToTransfer.LeftBytes -= uint64(read)
	s.FileToTransfer.SentBytes += uint64(read)
	s.SentFileBytesPackets++

	return nil
}

// Listens in an endless loop; reads incoming packages and puts them into channel
func (s *Sender) ReceivePackets() {
	for {
		incomingPacket, err := protocol.ReadFromConn(s.Connection)
		if err != nil {
			// in current implementation there is no way to receive a working file even if only one packet is missing
			fmt.Printf("Error reading a packet: %s\nExiting...", err)
			os.Exit(-1)
		}
		s.IncomingPackets <- incomingPacket
	}
}

// The "head" of the sender. "Glues" all things together.
// Current structure allows the sender to receive any type of packet
// in any order and react correspondingly
func (s *Sender) MainLoop() {

	go s.ReceivePackets()

	// send an information about the shared file to the receiver
	s.SendOffer()

	for {
		if s.Stopped {
			// exit the mainloop
			break
		}

		if s.TransferAllowed && s.ReceiverIsReady {
			err := s.SendPiece()
			if err != nil {
				fmt.Printf("could not send a piece of file: %s", err)
				s.Stop()
			}
			s.ReceiverIsReady = false
		}

		// no incoming packets ? Skipping the packet handling part
		if len(s.IncomingPackets) == 0 {
			continue
		}

		incomingPacket := <-s.IncomingPackets

		// handling each packet header differently
		switch incomingPacket.Header {

		case protocol.HeaderAccept:
			// allowed to send file packets
			s.TransferAllowed = true

		case protocol.HeaderReady:
			s.ReceiverIsReady = true

		case protocol.HeaderReject:
			s.Stop()

		case protocol.HeaderDisconnecting:
			// receiver is dropping the file transfer ?
			fmt.Println("Receiver has disconnected")
			s.Stop()
		}
	}
}