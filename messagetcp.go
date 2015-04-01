package coap

import (
	"bufio"
	"encoding/binary"
	"errors"
)

// TCPMessage is a CoAP Message that can encode itself for TCP
// transport.
type TcpMessage struct {
	Message
}

func (m *TcpMessage) MarshalBinary() ([]byte, error) {

	bin, err := m.Message.MarshalBinary()
	if err != nil {
		return nil, err
	}

	/*
		A CoAP TCP message looks like:

		     0                   1                   2                   3
		    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		   |        Message Length         |Ver| T |  TKL  |      Code     |
		   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		   |   Token (if any, TKL bytes) ...
		   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		   |   Options (if any) ...
		   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		   |1 1 1 1 1 1 1 1|    Payload (if any) ...
		   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	*/

	bin[2] = bin[0]
	bin[3] = bin[1]

	// insert len
	binary.BigEndian.PutUint16(bin, uint16(len(bin)-2))

	return bin, nil
}

func (m *TcpMessage) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return errors.New("short packet")
	}

	data[0] = data[2]
	data[1] = data[3]

	return m.Message.UnmarshalBinary(data)
}

func Decode(reader *bufio.Reader) (*TcpMessage, error) {
	header := make([]byte, 2)

	nr, err := reader.Read(header)

	if err != nil {
		return nil, err
	}

	if nr < 2 {
		return nil, errors.New("can't read 2 bytes")
	}

	ln := binary.BigEndian.Uint16(header)

	packet := make([]byte, ln)

	m := TcpMessage{}

	err = m.UnmarshalBinary(packet)
	return &m, err
}
