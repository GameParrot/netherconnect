package main

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/gameparrot/netherconnect/signaling"

	"github.com/df-mc/go-nethernet"
)

func (a *appInst) startNethernetListener() (nethernet.Signaling, error) {
	sd := signaling.Dialer{
		NetworkID: rand.Uint64(),
		Log:       a.log,
	}
	mcTok, err := a.authSession.MCToken(context.Background())
	if err != nil {
		return nil, fmt.Errorf("obtain MCToken: %w", err)
	}
	signalingConn, err := sd.DialContext(context.Background(), mcTok)
	if err != nil {
		return nil, err
	}

	return signalingConn, nil
}
