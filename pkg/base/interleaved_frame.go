package base

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const (
	// InterleavedFrameMagicByte is the first byte of an interleaved frame.
	InterleavedFrameMagicByte = 0x24
)

// InterleavedFrame is an interleaved frame, and allows to transfer binary data
// within RTSP/TCP connections. It is used to send and receive RTP and RTCP packets with TCP.
type InterleavedFrame struct {
	// channel ID
	Channel int

	// payload
	//Payload []byte
	bytes.Buffer
	PayLoadLength int
}

// Unmarshal decodes an interleaved frame.
func (f *InterleavedFrame) Unmarshal(br *bufio.Reader) error {
	f.Reset()
	var header [4]byte
	_, err := io.ReadFull(br, header[:])
	if err != nil {
		return err
	}

	if header[0] != InterleavedFrameMagicByte {
		return fmt.Errorf("invalid magic byte (0x%.2x)", header[0])
	}

	// it's useless to check payloadLen since it's limited to 65535
	payloadLen := int(uint16(header[2])<<8 | uint16(header[3]))
	f.PayLoadLength = payloadLen
	f.Channel = int(header[1])
	//f.Payload = make([]byte, payloadLen)

	_, err = io.CopyN(f, br, int64(payloadLen))
	return err
}

// MarshalSize returns the size of an InterleavedFrame.
func (f InterleavedFrame) MarshalSize() int {
	return 4 + f.PayLoadLength
}

// MarshalTo writes an InterleavedFrame.
func (f InterleavedFrame) MarshalTo(buf []byte) (int, error) {
	pos := 0

	pos += copy(buf[pos:], []byte{0x24, byte(f.Channel)})

	payloadLen := f.PayLoadLength
	buf[pos] = byte(payloadLen >> 8)
	buf[pos+1] = byte(payloadLen)
	pos += 2

	pos += copy(buf[pos:], f.Bytes())

	return pos, nil
}

// Marshal writes an InterleavedFrame.
func (f InterleavedFrame) Marshal() ([]byte, error) {
	buf := make([]byte, f.MarshalSize())
	_, err := f.MarshalTo(buf)
	return buf, err
}

func (f *InterleavedFrame) WritePayload(payload []byte) {
	f.PayLoadLength = len(payload)
	f.Reset()
	f.Write(payload)
}
