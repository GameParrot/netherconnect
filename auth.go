package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path"

	"fyne.io/fyne/v2"
	"github.com/bedrock-tool/bedrocktool/utils/xbox"
	"golang.org/x/oauth2"
)

var DeviceTypeIOSPreview = xbox.DeviceType{
	DeviceType: "iOS",
	ClientID:   "00000000403fc600",
	TitleID:    "1904044383",
	Version:    "15.6.1",
	UserAgent:  "XAL iOS 2021.11.20211021.000",
}

type authsrv struct {
	log     *slog.Logger
	handler xbox.MSAuthHandler

	liveToken  *oauth2.Token
	deviceType *xbox.DeviceType
}

var Auth *authsrv = &authsrv{
	log: slog.With("part", "Auth"),
}

// reads token from storage if there is one
func (a *authsrv) Startup() (err error) {
	tokenInfo, err := a.readToken()
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, errors.ErrUnsupported) {
		return nil
	}
	if err != nil {
		return err
	}
	a.liveToken = tokenInfo.Token
	a.deviceType = &DeviceTypeIOSPreview
	return nil
}

// if the user is currently logged in or not
func (a *authsrv) LoggedIn() bool {
	return a.liveToken != nil
}

// performs microsoft login using the handler passed
func (a *authsrv) SetHandler(handler xbox.MSAuthHandler) (err error) {
	a.handler = handler
	return nil
}

func (a *authsrv) Login(ctx context.Context, deviceType *xbox.DeviceType) (err error) {
	if deviceType == nil {
		deviceType = a.deviceType
	}
	if deviceType == nil {
		deviceType = &DeviceTypeIOSPreview
	}
	a.liveToken, err = xbox.RequestLiveTokenWriter(ctx, deviceType, a.handler)
	if err != nil {
		return err
	}
	a.deviceType = deviceType
	err = a.writeToken(tokenInfo{
		Token: a.liveToken,
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *authsrv) Logout() {
	a.liveToken = nil
	base := fyne.CurrentApp().Storage().RootURI().Path()
	os.Remove(path.Join(base, "token.json"))
	os.Remove(path.Join(base, "chain.bin"))
}

func (a *authsrv) refreshLiveToken() error {
	a.log.Info("Refreshing Microsoft Token")
	liveToken, err := xbox.RefreshToken(a.liveToken, a.deviceType)
	if err != nil {
		return err
	}
	a.liveToken = liveToken
	return a.writeToken(tokenInfo{
		Token: a.liveToken,
	})
}

var Ver1token func(f io.ReadSeeker, o any) error
var Tokene = func(w io.Writer, o any) error {
	return json.NewEncoder(w).Encode(o)
}

func readAuth[T any](name string) (*T, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var b = make([]byte, 1)
	_, err = f.ReadAt(b, 0)
	if err != nil {
		return nil, err
	}

	switch b[0] {
	case '{':
		var o T
		e := json.NewDecoder(f)
		err = e.Decode(&o)
		if err != nil {
			return nil, err
		}
		return &o, nil
	case '1':
		if Ver1token != nil {
			var o T
			err = Ver1token(f, &o)
			if err != nil {
				return nil, err
			}
			return &o, nil
		}
	}

	return nil, errors.ErrUnsupported
}

func writeAuth(name string, o any) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return Tokene(f, o)
}

type tokenInfo struct {
	*oauth2.Token
}

// writes the livetoken to storage
func (a *authsrv) writeToken(token tokenInfo) error {
	base := fyne.CurrentApp().Storage().RootURI().Path()
	return writeAuth(path.Join(base, "token.json"), token)
}

// reads the live token from storage, returns os.ErrNotExist if no token is stored
func (a *authsrv) readToken() (*tokenInfo, error) {
	base := fyne.CurrentApp().Storage().RootURI().Path()
	return readAuth[tokenInfo](path.Join(base, "token.json"))
}

var ErrNotLoggedIn = errors.New("not logged in")

// Token implements oauth2.TokenSource, returns ErrNotLoggedIn if there is no token, refreshes it if it expired
func (a *authsrv) Token() (t *oauth2.Token, err error) {
	if a.liveToken == nil {
		return nil, ErrNotLoggedIn
	}
	err = a.refreshLiveToken()
	if err != nil {
		return nil, err
	}
	return a.liveToken, nil
}
