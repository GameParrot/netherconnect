package main

import (
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *appInst) LoggingInScreen(w fyne.Window) fyne.CanvasObject {
	go func() {
		if err := a.login(); err != nil {
			text := canvas.NewText("Failed to log in", fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
			text.TextSize = 50
			infoText := widget.NewLabel(err.Error())
			if strings.Contains(infoText.Text, "2148916254") {
				infoText.Text += "\nTry again in a minute."
			}
			infoText.Wrapping = fyne.TextWrapWord
			tryAgainButton := widget.NewButton("Try again", func() {
				w.SetContent(a.LoggingInScreen(w))
			})
			logout := widget.NewButton("Log out", func() {
				Auth.Logout()
				a.xsts = nil
				a.playfabIdentity = nil
				a.lastUpdateTime = time.Time{}
				a.xuid = ""
				fyne.Do(func() { w.SetContent(a.LoginContent(w)) })
			})
			fyne.Do(func() { w.SetContent(container.NewCenter(container.NewVBox(text, infoText, tryAgainButton, logout))) })
		} else {
			fyne.Do(func() { w.SetContent(a.ServerListContent(w, false)) })
		}
	}()

	text := canvas.NewText("Logging in...", fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	text.TextSize = 50

	return container.NewCenter(text)
}
