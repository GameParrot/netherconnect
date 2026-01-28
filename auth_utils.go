package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gameparrot/netherconnect/session"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

var DevicePreview = auth.Config{ClientID: "00000000403fc600", DeviceType: "iOS", Version: "0.0.0"}

func (a *appInst) login() error {
	var err error
	a.authSession, err = session.SessionFromTokenSource(a.tokenSrc, DevicePreview, context.Background())
	if err != nil {
		return err
	}
	accountInfoTok, err := a.authSession.RequestXBLToken(context.Background(), "http://xboxlive.com")
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

	token, err := DevicePreview.RequestLiveTokenWriter(&fyneTextWriter{label: label, done: func(err error) {
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
	src := DevicePreview.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		src = DevicePreview.RefreshTokenSource(a.requestToken(w))
	}
	tok, _ := src.Token()
	b, _ := json.Marshal(tok)
	_ = os.WriteFile(path.Join(base, "token.json"), b, 0644)
	a.tokenSrc = src
}
