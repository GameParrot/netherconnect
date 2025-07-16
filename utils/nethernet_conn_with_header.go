package utils

import "github.com/df-mc/go-nethernet"

type NethernetConnWithHeader struct {
	*nethernet.Conn
}

func (g *NethernetConnWithHeader) Write(b []byte) (n int, err error) {
	return g.Conn.Write(b[1:])
}

func (g *NethernetConnWithHeader) ReadPacket() ([]byte, error) {
	pk, err := g.Conn.ReadPacket()
	if err != nil {
		return nil, err
	}
	return append([]byte{254}, pk...), nil
}
