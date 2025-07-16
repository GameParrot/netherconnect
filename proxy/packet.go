package proxy

import (
	"bytes"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PacketData holds the data of a Minecraft packet.
type PacketData struct {
	h       *packet.Header
	full    []byte
	payload *bytes.Buffer
}

// ParseData parses the packet data slice passed into a packetData struct.
func ParseData(data []byte) (*PacketData, error) {
	buf := bytes.NewBuffer(data)
	header := &packet.Header{}
	if err := header.Read(buf); err != nil {
		// We don't return this as an error as it's not in the hand of the user to control this. Instead,
		// we return to reading a new packet.
		return nil, fmt.Errorf("read packet header: %w", err)
	}
	return &PacketData{h: header, full: data, payload: buf}, nil
}

func (p *PacketData) Header() *packet.Header {
	return p.h
}

func (p *PacketData) Payload() *bytes.Buffer {
	return p.payload
}
