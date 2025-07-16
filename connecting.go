package main

import (
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func (a *appInst) Connect(w fyne.Window, ip string, port string, alreadyConnected bool) fyne.CanvasObject {
	text := canvas.NewText("Connecting...", fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	text.TextSize = 50

	go func() {
		bestIp := GetLowestPingIP(ip, port, a.log)
		a.log.Info("Found best IP: " + bestIp)
		a.currentAddrRaw = net.JoinHostPort(ip, port)
		a.currentAddr = net.JoinHostPort(bestIp, port)
		fyne.Do(func() { w.SetContent(a.ConnectedScreen(w, alreadyConnected)) })
	}()

	return container.NewCenter(text)

}
