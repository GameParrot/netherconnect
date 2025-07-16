package utils

import (
	"errors"
	"log/slog"
	"math/rand"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/gameparrot/netherconnect/internal"
)

type Pinger struct {
	servers   map[netip.AddrPort]func(s ServerStatus)
	serversMu sync.RWMutex
	conn      *net.UDPConn
	clientId  int64
	startTime time.Time
}

type server struct {
	ip netip.AddrPort
	f  func(s ServerStatus)
}

func NewPinger(log *slog.Logger) (*Pinger, error) {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	p := &Pinger{conn: conn, servers: make(map[netip.AddrPort]func(s ServerStatus)), clientId: rand.Int63(), startTime: time.Now()}
	go func() {
		b := make([]byte, 1500)
		for {
			n, addr, err := conn.ReadFromUDP(b)
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				continue
			}
			if n == 0 {
				continue
			}
			ip, _ := netip.AddrFromSlice(addr.IP)
			if ip.Is4In6() {
				ip = ip.Unmap()
			}
			addrPort := netip.AddrPortFrom(ip, uint16(addr.Port))
			p.serversMu.RLock()
			if statusFunc, ok := p.servers[addrPort]; ok {
				pk := &internal.UnconnectedPong{}
				if err := pk.UnmarshalBinary(b[1:n]); err == nil {
					if pongData, err := parsePongData(pk.Data); err == nil {
						statusFunc(pongData)
					}
				} else {
					log.Error("Error reading unconnected pong", "err", err.Error())
				}
			}
			p.serversMu.RUnlock()
		}
	}()
	go func() {
		for {
			p.serversMu.RLock()
			servers := make([]server, len(p.servers))
			i := 0
			for ip, f := range p.servers {
				servers[i] = server{ip: ip, f: f}
				i++
			}
			p.serversMu.RUnlock()

			for _, srv := range servers {
				if err := p.pingServer(srv); err != nil {
					if errors.Is(err, net.ErrClosed) {
						return
					}
				}
				time.Sleep(100 * time.Millisecond)
			}

			time.Sleep(max(0, (5*time.Second)-(100*time.Millisecond*time.Duration(len(servers)))))
		}
	}()
	return p, nil
}

func (p *Pinger) AddServer(ipStr string, statusFunc func(s ServerStatus)) error {
	addr, err := net.ResolveUDPAddr("udp", ipStr)
	if err != nil {
		return err
	}
	ip, _ := netip.AddrFromSlice(addr.IP)
	if ip.Is4In6() {
		ip = ip.Unmap()
	}
	addrPort := netip.AddrPortFrom(ip, uint16(addr.Port))
	p.serversMu.Lock()
	p.servers[addrPort] = statusFunc
	p.serversMu.Unlock()
	p.pingServer(server{ip: addrPort, f: statusFunc})
	return nil
}

func (p *Pinger) RemoveServer(ipStr string) error {
	addr, err := net.ResolveUDPAddr("udp", ipStr)
	if err != nil {
		return err
	}
	ip, _ := netip.AddrFromSlice(addr.IP)
	if ip.Is4In6() {
		ip = ip.Unmap()
	}
	addrPort := netip.AddrPortFrom(ip, uint16(addr.Port))
	p.serversMu.Lock()
	delete(p.servers, addrPort)
	p.serversMu.Unlock()
	return nil
}

func (p *Pinger) pingServer(srv server) error {
	addr, err := net.ResolveUDPAddr("udp", srv.ip.String())
	if err != nil {
		return err
	}
	pk := &internal.UnconnectedPing{PingTime: int64(time.Since(p.startTime).Milliseconds()), ClientGUID: p.clientId}
	pkBytes, _ := pk.MarshalBinary()
	if _, err := p.conn.WriteToUDP(pkBytes, addr); err != nil {
		if errors.Is(err, net.ErrClosed) {
			return err
		}
	}
	return nil
}

func (p *Pinger) Close() error {
	return p.conn.Close()
}
