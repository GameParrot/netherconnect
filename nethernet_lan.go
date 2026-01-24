package main

import (
	"math/rand/v2"

	"github.com/df-mc/go-nethernet"
	"github.com/df-mc/go-nethernet/discovery"
)

func (a *appInst) startNethernetLan(errorFunc func(error)) error {
	nid := rand.Uint64()
	d, err := discovery.ListenConfig{
		Log:       a.log,
		NetworkID: nid,
	}.Listen("0.0.0.0:7551")
	if err != nil {
		return err
	}
	d.ServerData(&discovery.ServerData{
		ServerName:     "NetherConnect",
		LevelName:      "NetherConnect",
		GameType:       0,
		PlayerCount:    1,
		MaxPlayerCount: 2,
		TransportLayer: 2,
		ConnectionType: 4,
	})
	l, err := nethernet.ListenConfig{Log: a.log}.Listen(d)
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				errorFunc(err)
				return
			}
			a.handleNetherNetConn(conn.(*nethernet.Conn), nil)
		}
	}()
	return nil
}
