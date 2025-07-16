package main

import (
	"github.com/gameparrot/netherconnect/utils"
	"bytes"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	su "github.com/nyaosorg/go-windows-su"
)

func checkNetIsolation(w fyne.Window) {
	if runtime.GOOS != "windows" {
		// Only an issue on Windows.
		return
	}
	data, _ := exec.Command("CheckNetIsolation", "LoopbackExempt", "-s", `-n="microsoft.minecraftuwp_8wekyb3d8bbwe"`).CombinedOutput()
	if bytes.Contains(data, []byte("microsoft.minecraftuwp_8wekyb3d8bbwe")) {
		return
	}
	utils.ShowAlert(w.Canvas(), "To allow connection, you have to exempt Minecraft from\n loopback restrictions. Click continue to proceed.", "Continue", func() {
		_, _ = su.ShellExecute(su.RUNAS,
			"CheckNetIsolation",
			"LoopbackExempt -a -n=\"Microsoft.MinecraftUWP_8wekyb3d8bbwe\"",
			`C:\`)

	})
}

func (a *appInst) ConnectedScreen(w fyne.Window, alreadyConnected bool) fyne.CanvasObject {
	if !alreadyConnected {
		checkNetIsolation(w)
		a.startTransferServer(func(err error) {
			dialog.NewError(err, w).Show()
		})
	}

	text := canvas.NewText("Server started", fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	text.TextSize = 50

	joinText := widget.NewLabel("To join, add 127.0.0.1 port 19132 to your servers and join or join \"NetherConnect\" in LAN. If joining fails, try again.")

	return container.NewBorder(nil, widget.NewButton("Change server", func() {
		w.SetContent(a.ServerListContent(w, true))
	}), nil, nil, container.NewCenter(container.NewVBox(text, joinText)))
}
