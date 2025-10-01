package main

import (
	"net"
	"os"
	"path"
	"slices"
	"strconv"
	"time"

	"github.com/gameparrot/netherconnect/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type server struct {
	Name         string
	IP           string
	Port         uint16
	LastPlayTime time.Time
}

func (a *appInst) ServerListContent(w fyne.Window, alreadyConnected bool) fyne.CanvasObject {
	text := canvas.NewText("Servers", fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	text.TextSize = 50

	slices.SortFunc(a.servers, func(a, b server) int {
		return b.LastPlayTime.Compare(a.LastPlayTime)
	})

	pinger, _ := utils.NewPinger(a.log)

	var tree *widget.Tree
	tree = &widget.Tree{
		ChildUIDs: func(uid widget.TreeNodeID) (c []widget.TreeNodeID) {
			for i := range a.servers {
				c = append(c, strconv.Itoa(i))
			}
			return c
		},
		IsBranch: func(uid string) bool {
			return uid == ""
		},
		CreateNode: func(branch bool) (o fyne.CanvasObject) {
			return container.New(utils.NewFixedSizeLayoutExpand(fyne.NewSize(0, 50)))
		},
		UpdateNode: func(uid string, branch bool, co fyne.CanvasObject) {
			lii, _ := strconv.Atoi(uid)
			if len(co.(*fyne.Container).Objects) != 0 {
				if co.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*canvas.Text).Text == a.servers[lii].Name {
					return
				}
			}

			nameText := canvas.NewText(a.servers[lii].Name, fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
			nameText.TextSize = 20
			motdText := widget.NewTextGrid()
			motdText.SetText("Loading...")
			playerText := widget.NewLabel("")

			editButton := widget.NewButton("Edit", func() {
				srv := a.servers[lii]
				a.EditServer(w, lii, func(removed bool) {
					if pinger != nil {
						pinger.RemoveServer(net.JoinHostPort(srv.IP, strconv.Itoa(int(srv.Port))))
					}
					if !removed {
						nameText.Text = "" // make sure the server motd is refreshed
					}
					tree.Refresh()
				})
			})

			co.(*fyne.Container).Objects = []fyne.CanvasObject{container.NewBorder(nil, nil, nil, container.NewHBox(playerText, widget.NewButton("Join", func() {
				a.servers[lii].LastPlayTime = time.Now()
				if pinger != nil {
					pinger.Close()
				}
				fyne.Do(func() {
					w.SetContent(a.Connect(w, a.servers[lii].IP, strconv.Itoa(int(a.servers[lii].Port)), alreadyConnected))
				})
			}), editButton), container.NewVBox(nameText, motdText))}
			if pinger != nil {
				pinger.AddServer(net.JoinHostPort(a.servers[lii].IP, strconv.Itoa(int(a.servers[lii].Port))), func(pongData utils.ServerStatus) {
					fyne.Do(func() {
						playerText.SetText(strconv.Itoa(pongData.PlayerCount) + "/" + strconv.Itoa(pongData.MaxPlayers) + "\nv" + pongData.Version + " (" + strconv.Itoa(pongData.ProtocolID) + ")")
					})
					colored := utils.ParseText(pongData.ServerName + "§r  " + pongData.WorldName)
					cells := []widget.TextGridCell{}
					for _, c := range colored {
						style := &widget.CustomTextGridStyle{TextStyle: fyne.TextStyle{Bold: c.Style.Bold(), Italic: c.Style.Italic()}, FGColor: c.Color}
						for _, r := range c.Text {
							cells = append(cells, widget.TextGridCell{Rune: r, Style: style})
						}
					}
					motdText.SetRow(0, widget.TextGridRow{Cells: cells})
				})
			}
		},
	}

	addServer := widget.NewButton("Add server", func() {
		a.AddServer(w, func(srv server) {
			tree.Refresh()
		})
	})

	logout := widget.NewButton("Log out", func() {
		os.Remove(path.Join(fyne.CurrentApp().Storage().RootURI().Path(), "token.json"))
		a.authSession = nil
		a.xuid = ""
		if pinger != nil {
			pinger.Close()
		}
		fyne.Do(func() { w.SetContent(a.LoggingInScreen(w)) })
	})

	return container.NewBorder(container.NewCenter(text), container.NewVBox(addServer, logout), nil, nil, tree)
}
