package main

import (
	"context"
	"math/rand/v2"
	"strconv"

	"github.com/gameparrot/netherconnect/franchise"
	"github.com/gameparrot/netherconnect/franchise/signaling"

	"github.com/df-mc/go-nethernet"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"golang.org/x/text/language"
)

func (a *appInst) startNethernetListener() (nethernet.Signaling, error) {
	region, _ := language.English.Region()

	conf := &franchise.TokenConfig{
		Device: &franchise.DeviceConfig{
			ApplicationType: franchise.ApplicationTypeMinecraftPE,
			Capabilities:    []string{franchise.CapabilityRayTracing},
			GameVersion:     protocol.CurrentVersion,
			ID:              uuid.New(),
			Memory:          strconv.FormatUint(rand.Uint64(), 10),
			Platform:        franchise.PlatformWindows10,
			PlayFabTitleID:  a.env.PlayFabTitleID,
			StorePlatform:   franchise.StorePlatformUWPStore,
			Type:            franchise.DeviceTypeWindows10,
		},
		User: &franchise.UserConfig{
			Language:     language.English,
			LanguageCode: language.AmericanEnglish,
			RegionCode:   region.String(),
			Token:        a.playfabIdentity.SessionTicket,
			TokenType:    franchise.TokenTypePlayFab,
		},
		Environment: &a.env,
	}

	sd := signaling.Dialer{
		NetworkID: rand.Uint64(),
		Log:       a.log,
	}
	signalingConn, err := sd.DialContext(context.Background(), tokenConfigSource(func() (*franchise.TokenConfig, error) {
		return conf, nil
	}), &a.signalingEnv)
	if err != nil {
		return nil, err
	}

	return signalingConn, nil
}

type tokenConfigSource func() (*franchise.TokenConfig, error)

func (f tokenConfigSource) TokenConfig() (*franchise.TokenConfig, error) { return f() }
