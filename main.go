package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"path"
	"slices"
	"syscall"
	"time"
	_ "unsafe"

	"github.com/gameparrot/netherconnect/auth"
	"github.com/gameparrot/netherconnect/github"
	"golang.org/x/oauth2"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

type appInst struct {
	enableLanMode bool

	log *slog.Logger

	tokenSrc oauth2.TokenSource

	authSession *auth.Session

	currentAddr    string
	currentAddrRaw string
	xuid           string

	servers []server

	nethernetId string
}

const (
	repo = "gameparrot/netherconnect"
	tag  = "v1.3.1"
)

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

	if len(os.Args) > 1 {
		appInst.enableLanMode = slices.Contains(os.Args[1:], "--lan-mode")
	}

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

	if appInst.enableLanMode {
		if err := appInst.startNethernetLan(func(err error) {
			dialog.NewError(err, w).Show()
		}); err != nil {
			if errors.Is(err, syscall.EADDRINUSE) {
				err = errors.New("you must start netherconnect before starting minecraft when in lan mode")
			}
			e := dialog.NewError(err, w)
			e.SetOnClosed(fyne.CurrentApp().Quit)
			e.Show()
			w.ShowAndRun()
			return
		}
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		fyne.Do(func() { w.SetContent(appInst.LoggingInScreen(w)) })
	}()

	go func() {
		latest, err := github.GetLatestRelease(repo)
		if err != nil {
			appInst.log.Error(err.Error())
			return
		}
		if latest.TagName != tag {
			info := dialog.NewConfirm("Update Available", "A new version of NetherConnect is available. Do you want to download it?", func(b bool) {
				if b {
					url, _ := url.Parse("https://github.com/" + repo + "/releases/latest")
					a.OpenURL(url)
				}
			}, w)
			info.Show()
		}
	}()

	w.SetOnClosed(func() {
		jsonData, err := json.Marshal(appInst.servers)
		if err == nil {
			os.WriteFile(path.Join(fyne.CurrentApp().Storage().RootURI().Path(), "servers.json"), jsonData, 0644)
		}
	})

	w.ShowAndRun()
}
