package main

import (
	"context"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MSAuthHandler struct {
	codeCb     func(uri, code string)
	finishedCb func(err error)
}

func (m *MSAuthHandler) AuthCode(uri, code string) {
	m.codeCb(uri, code)
}

func (m *MSAuthHandler) Finished(err error) {
	m.finishedCb(err)
}

func (a *appInst) LoginContent(w fyne.Window) fyne.CanvasObject {
	text := canvas.NewText("Login", fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	text.TextSize = 50

	loginText := widget.NewLabel("")

	var wg sync.WaitGroup

	Auth.handler = &MSAuthHandler{codeCb: func(uri, code string) {
		loginText.SetText("Authenticate at " + uri + " using the code " + code)
	}, finishedCb: func(err error) {
		go func() {
			if err != nil {
				dialog.NewError(fmt.Errorf("failed to sign in: %w", err), w).Show()
				return
			}
			wg.Wait()
			if err := a.login(); err != nil {
				dialog.NewError(fmt.Errorf("failed to sign in: %w", err), w).Show()
				return
			}
			fyne.Do(func() { w.SetContent(a.ServerListContent(w, false)) })
		}()
	}}

	wg.Add(1)
	go func() {
		Auth.Login(context.Background(), &DeviceTypeIOSPreview)
		a.log.Info("logged in")
		wg.Done()
	}()

	return container.NewCenter(container.NewVBox(text, loginText))
}
