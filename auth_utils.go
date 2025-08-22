package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gameparrot/netherconnect/auth"
	"github.com/gameparrot/netherconnect/franchise"
	"github.com/gameparrot/netherconnect/playfab"
	"golang.org/x/oauth2"

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
	liveToken, err := a.tokenSrc.Token()
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

type fyneTextWriter struct {
	label *widget.Label
	done  func(err error)
}

func (f *fyneTextWriter) Write(b []byte) (int, error) {
	str := strings.TrimSuffix(string(b), "\n")
	if str == "Authentication successful." {
		f.done(nil)
		return len(b), nil
	}
	if strings.HasPrefix(str, "error") {
		f.done(errors.New(str))
		return len(b), nil
	}
	fyne.DoAndWait(func() { f.label.SetText(f.label.Text + string(b)) })
	return len(b), nil
}

func (a *appInst) requestToken(w fyne.Window) *oauth2.Token {
	var popup *widget.PopUp
	text := canvas.NewText("Login", fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	text.TextSize = 25

	label := widget.NewLabel("")
	popup = widget.NewModalPopUp(container.NewVBox(
		text,
		label,
	), w.Canvas())
	fyne.Do(func() { popup.Show() })

	token, err := auth.RequestLiveTokenWriter(&fyneTextWriter{label: label, done: func(err error) {
		fyne.Do(func() { popup.Hide() })
		if err != nil {
			err := dialog.NewError(fmt.Errorf("failed to sign in: %w", err), w)
			err.SetOnClosed(func() {
				os.Exit(0)
			})
			err.Show()
		}
	}})
	if err != nil {
		popup.Hide()
		err := dialog.NewError(fmt.Errorf("failed to sign in: %w", err), w)
		err.SetOnClosed(func() {
			os.Exit(0)
		})
		err.Show()
	}
	return token
}

func (a *appInst) tokenSource(w fyne.Window) {
	base := fyne.CurrentApp().Storage().RootURI().Path()

	token := new(oauth2.Token)
	tokenData, err := os.ReadFile(path.Join(base, "token.json"))
	if err == nil {
		_ = json.Unmarshal(tokenData, token)
	} else {
		token = a.requestToken(w)
	}
	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		src = auth.RefreshTokenSource(a.requestToken(w))
	}
	tok, _ := src.Token()
	b, _ := json.Marshal(tok)
	_ = os.WriteFile(path.Join(base, "token.json"), b, 0644)
	a.tokenSrc = src
}
