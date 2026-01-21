package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path"
	"time"
	_ "unsafe"

	"github.com/gameparrot/netherconnect/auth"
	"golang.org/x/oauth2"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

type appInst struct {
	log *slog.Logger

	tokenSrc oauth2.TokenSource

	authSession *auth.Session

	currentAddr    string
	currentAddrRaw string
	xuid           string

	servers []server

	nethernetId string
}

func (a *appInst) addFeaturedServers() {
	a.servers = append(a.servers, server{Name: "The Hive", IP: "geo.hivebedrock.network", Port: 19132})
	a.servers = append(a.servers, server{Name: "Mineville", IP: "play.inpvp.net", Port: 19132})
	a.servers = append(a.servers, server{Name: "Lifeboat", IP: "mco.lbsg.net", Port: 19132})
	a.servers = append(a.servers, server{Name: "CubeCraft", IP: "mco.cubecraft.net", Port: 19132})
	a.servers = append(a.servers, server{Name: "Galaxite", IP: "play.galaxite.net", Port: 19132})
	a.servers = append(a.servers, server{Name: "Enchanted Dragons", IP: "play.enchanted.gg", Port: 19132})
}

func main() {
	appInst := &appInst{log: slog.Default()}

	a := app.NewWithID("com.gameparrot.netherconnect")
	a.Settings().SetTheme(&forcedVariant{Theme: theme.DefaultTheme(), variant: theme.VariantDark})
	w := a.NewWindow("NetherConnect")
	w.SetMaster()

	serverFile, err := os.ReadFile(path.Join(fyne.CurrentApp().Storage().RootURI().Path(), "servers.json"))
	if err == nil {
		json.Unmarshal(serverFile, &appInst.servers)
	} else if errors.Is(err, os.ErrNotExist) {
		appInst.addFeaturedServers()
	}

	w.Resize(fyne.NewSize(640, 460))

	go func() {
		time.Sleep(100 * time.Millisecond)
		fyne.Do(func() { w.SetContent(appInst.LoggingInScreen(w)) })
	}()

	w.SetOnClosed(func() {
		jsonData, err := json.Marshal(appInst.servers)
		if err == nil {
			os.WriteFile(path.Join(fyne.CurrentApp().Storage().RootURI().Path(), "servers.json"), jsonData, 0644)
		}
	})

	w.ShowAndRun()
}
