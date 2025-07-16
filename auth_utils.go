package main

import (
	"github.com/gameparrot/netherconnect/auth"
	"github.com/gameparrot/netherconnect/franchise"
	"github.com/gameparrot/netherconnect/playfab"
	"context"
	"fmt"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func (a *appInst) initDiscovery() error {
	discovery, err := franchise.Discover(protocol.CurrentVersion)
	if err != nil {
		return fmt.Errorf("discover: %w", err)
	}
	if err := discovery.Environment(&a.env, franchise.EnvironmentTypeProduction); err != nil {
		return fmt.Errorf("decode environment: %w", err)
	}

	if err := discovery.Environment(&a.signalingEnv, franchise.EnvironmentTypeProduction); err != nil {
		return err
	}
	return nil
}

func (a *appInst) updateXstsToken() (*auth.XblTokenObtainer, error) {
	if time.Since(a.lastUpdateTime) < 45*time.Minute {
		return nil, nil
	}
	liveToken, err := Auth.Token()
	if err != nil {
		return nil, fmt.Errorf("request Live Connect token: %w", err)
	}

	obtainer, err := auth.NewXblTokenObtainer(liveToken, context.Background())
	if err != nil {
		return nil, fmt.Errorf("request Live Device token: %w", err)
	}

	a.xsts, err = obtainer.RequestXBLToken(context.Background(), "https://multiplayer.minecraft.net/")
	if err != nil {
		return nil, fmt.Errorf("request XBOX Live token: %w", err)
	}

	playfabXBL, err := obtainer.RequestXBLToken(context.Background(), "http://playfab.xboxlive.com/")
	if err != nil {
		return nil, fmt.Errorf("request Playfab token: %w", err)
	}

	a.playfabIdentity, err = playfab.Login{
		Title:         "20CA2",
		CreateAccount: true,
	}.WithXBLToken(playfabXBL).Login()
	if err != nil {
		return nil, fmt.Errorf("error logging in to playfab: %w", err)
	}

	a.lastUpdateTime = time.Now()
	return obtainer, nil
}

func (a *appInst) login() error {
	obtainer, err := a.updateXstsToken()
	if err != nil {
		return err
	}
	accountInfoTok, err := obtainer.RequestXBLToken(context.Background(), "http://xboxlive.com")
	if err != nil {
		return fmt.Errorf("request XBOX site token: %w", err)
	}
	a.xuid = accountInfoTok.AuthorizationToken.DisplayClaims.UserInfo[0].XUID
	return nil
}
