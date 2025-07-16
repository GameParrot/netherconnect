package main

import (
	"github.com/gameparrot/netherconnect/proxy"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/akmalfairuz/legacy-version/legacyver/legacypacket"
	"github.com/akmalfairuz/legacy-version/legacyver/proto"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *appInst) startTransferServer(errorFunc func(error)) error {
	addrs := []string{"127.0.0.1:19132", "[::1]:19133"}

	for _, addr := range addrs {
		list, err := raknet.Listen(addr)
		if err != nil {
			return fmt.Errorf("start transfer server: %w", err)
		}

		list.PongData(fmt.Appendf(nil, "MCPE;NetherConnect;100;1.0;0;1;%d;NetherConnect;Survival;", list.ID()))

		go func() {
			for {
				conn, err := list.Accept()
				if err != nil {
					errorFunc(err)
					return
				}
				go func() {
					err := a.handleRaknetConn(conn.(*raknet.Conn))
					if err != nil {
						errorFunc(err)
					}
				}()
			}
		}()
	}
	return nil
}

func (a *appInst) handleRaknetConn(conn *raknet.Conn) error {
	protocolId := int32(0)
	decoder := packet.NewDecoder(conn)

decodeLoop:
	for {
		decoded, err := decoder.Decode()
		if err != nil {
			conn.Close()
			return err
		}
		for _, pkBytes := range decoded {
			pkData, err := proxy.ParseData(pkBytes)
			if err != nil {
				conn.Close()
				return err
			}
			if pkData.Header().PacketID == packet.IDRequestNetworkSettings {
				pk := &packet.RequestNetworkSettings{}
				pk.Marshal(protocol.NewReader(bytes.NewReader(pkData.Payload().Bytes()), 0, true))
				decoder.EnableCompression(1024 * 1024 * 8)
				protocolId = pk.ClientProtocol
				break decodeLoop
			}
		}
	}

	encoder := packet.NewEncoder(conn)
	if protocolId < proto.ID818 {
		writePacketToEncoder(&legacypacket.Disconnect{Message: "NetherConnect requires Minecraft 1.21.90 or newer."}, encoder, protocolId)
		return errors.New("netherconnect requires 1.21.90 or newer")
	}
	writePacketToEncoder(&packet.NetworkSettings{CompressionAlgorithm: packet.FlateCompression.EncodeCompression()}, encoder, protocolId)
	encoder.EnableCompression(packet.FlateCompression)
	writePacketToEncoder(&legacypacket.StartGame{PlayerMovementSettings: (&proto.PlayerMovementSettings{}).FromLatest(protocol.PlayerMovementSettings{})}, encoder, protocolId)

	if a.nethernetId == 0 {
		err := a.startNethernet()
		if err != nil {
			return err
		}
	}

	writePacketToEncoder(&legacypacket.Transfer{Address: strconv.FormatUint(a.nethernetId, 10)}, encoder, protocolId)
	time.Sleep(1 * time.Second)
	conn.Close()
	return nil
}

func writePacketToEncoder(pk packet.Packet, enc *packet.Encoder, protocolId int32) {
	header := packet.Header{PacketID: pk.ID()}
	buf := bytes.NewBuffer([]byte{})
	header.Write(buf)
	pk.Marshal(proto.NewWriter(protocol.NewWriter(buf, 0), protocolId))
	enc.Encode([][]byte{buf.Bytes()})
}
