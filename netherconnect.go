package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
	_ "unsafe"

	"github.com/gameparrot/netherconnect/proxy"
	"github.com/gameparrot/netherconnect/utils"

	"github.com/akmalfairuz/legacy-version/legacyver/legacypacket"
	"github.com/akmalfairuz/legacy-version/legacyver/proto"
	"github.com/df-mc/go-nethernet"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *appInst) startNethernet() error {
	singlaingConn, err := a.startNethernetListener()
	if err != nil {
		return fmt.Errorf("start nethernet listener: %w", err)
	}

	l, err := nethernet.ListenConfig{Log: a.log}.Listen(singlaingConn)
	if err != nil {
		return err
	}
	a.log.Info("Network ID: " + strconv.FormatUint(singlaingConn.NetworkID(), 10))

	go func() {
		for {
			conn, err := l.Accept()
			if err == nil {
				go a.handleNetherNetConn(conn.(*nethernet.Conn), l)
			} else {
				return
			}
		}
	}()

	a.nethernetId = singlaingConn.NetworkID()

	return nil
}

func (a *appInst) handleNetherNetConn(rawConn *nethernet.Conn, list *nethernet.Listener) {
	list.Close()
	a.nethernetId = 0

	pendingTransfer := false

	conn := &utils.NethernetConnWithHeader{Conn: rawConn}

	clientConn := proxy.NewProxyConn(conn, true)
	clientConn.SetAuthEnabled(true)
	if err := clientConn.ReadLoop(); err != nil {
		a.log.Error("Client failed to login", "err", err.Error())
		clientConn.WritePacket(&packet.Disconnect{Message: "Error: " + err.Error()})
		time.Sleep(1 * time.Second)
		conn.Close()
		return
	}

	if clientConn.IdentityData().XUID != a.xuid {
		clientConn.WritePacket(&packet.Disconnect{Message: "Error: Account mismatch"})
		time.Sleep(1 * time.Second)
		conn.Close()
		return
	}

	a.log.Info("Client logged in")

	rkConn, err := raknet.Dial(a.currentAddr)
	if err != nil {
		a.log.Error("Failed to dial remote server", "err", err.Error())
		clientConn.WritePacket(&packet.Disconnect{Message: "Error: " + err.Error()})
		time.Sleep(1 * time.Second)
		conn.Close()
		return
	}

	serverConn := proxy.NewProxyConn(rkConn, false)

	cd := clientConn.ClientData()
	cd.DeviceOS = 7
	cd.PlatformType = 0
	cd.GameVersion += ".24"
	cd.DefaultInputMode = 1
	cd.ServerAddress = a.currentAddrRaw

	defer rkConn.Close()
	defer conn.Close()

	if err := serverConn.Login(cd, a.authSession, clientConn.Protocol()); err != nil {
		a.log.Error("Failed to login to server", "err", err.Error())
		clientConn.WritePacket(&packet.Disconnect{Message: "Error: " + err.Error()})
		time.Sleep(1 * time.Second)
		conn.Close()
		return
	}
	a.log.Info("Server logged in")

	go func() {
		defer rkConn.Close()
		defer func() {
			if !pendingTransfer {
				conn.Close()
			}
		}()
		for {
			pks, err := clientConn.ReadPackets()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					a.log.Error("Failed to read packet from client", "err", err.Error())
				}
				return
			}
			if err := serverConn.WritePackets(pks); err != nil {
				if !errors.Is(err, net.ErrClosed) {
					a.log.Error("Failed to write packet to server", "err", err.Error())
				}
				//return
			}
		}
	}()

	for {
		pks, err := serverConn.ReadPackets()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				a.log.Error("Failed to read packet from server", "err", err.Error())
			}
			clientConn.WritePacket(&packet.Disconnect{Message: "Error: " + err.Error()})
			time.Sleep(1 * time.Second)
			conn.Close()
			return
		}

		for _, pk := range pks {
			if data, err := proxy.ParseData(pk); err == nil {
				if data.Header().PacketID == packet.IDTransfer {
					pendingTransfer = true
					transfer := &legacypacket.Transfer{}
					reader := proto.NewReader(protocol.NewReader(data.Payload(), 0, true), clientConn.Protocol())
					transfer.Marshal(reader)

					portStr := strconv.Itoa(int(transfer.Port))
					bestIp := GetLowestPingIP(transfer.Address, portStr, a.log)
					a.log.Info("Found best IP: " + bestIp)
					a.currentAddrRaw = net.JoinHostPort(transfer.Address, portStr)
					a.currentAddr = net.JoinHostPort(bestIp, portStr)

					if a.nethernetId == 0 {
						err := a.startNethernet()
						if err != nil {
							a.log.Error("Failed to start NetherNet", "err", err.Error())
							clientConn.WritePacket(&packet.Disconnect{Message: "Error: " + err.Error()})
							time.Sleep(1 * time.Second)
							return
						}
					}

					clientConn.WritePacket(&legacypacket.Transfer{Address: strconv.FormatUint(a.nethernetId, 10)})
					time.Sleep(1 * time.Second)
					return
				}
			}
		}

		if err := clientConn.WritePackets(pks); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				a.log.Error("Failed to write packet to client", "err", err.Error())
			}
			return
		}
	}

}
