package main

import (
	"github.com/gameparrot/netherconnect/utils"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *appInst) AddServer(w fyne.Window, addFunc func(srv server)) {
	text := canvas.NewText("Add server", fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	text.TextSize = 50

	var popup *widget.PopUp

	nameBox := utils.NewObjectWithLabel("Name: ", widget.NewEntry())
	ipBox := utils.NewObjectWithLabel("IP: ", widget.NewEntry())
	portBox := utils.NewObjectWithLabel("Port: ", widget.NewEntry())
	portBox.Obj().SetText("19132")

	button := widget.NewButton("Add", func() {
		portNum, err := strconv.Atoi(portBox.Obj().Text)
		if err != nil {
			dialog.NewError(fmt.Errorf("failed to parse port: %w", err), w).Show()
			return
		}
		if portNum < 1 || portNum > 65535 {
			dialog.NewError(errors.New("port must be between 1 and 65535"), w).Show()
			return
		}
		srv := server{Name: nameBox.Obj().Text, IP: ipBox.Obj().Text, Port: uint16(portNum), LastPlayTime: time.Now()}
		a.servers = append(a.servers, srv)

		addFunc(srv)

		popup.Hide()
	})

	cancel := widget.NewButton("Cancel", func() {
		popup.Hide()
	})

	popup = widget.NewModalPopUp(container.NewCenter(container.NewVBox(container.New(utils.NewFixedSizeLayoutExpand(fyne.NewSize(350, 60)), container.NewCenter(text)), nameBox.Container(), ipBox.Container(), portBox.Container(), cancel, button)), w.Canvas())

	popup.Show()
}

func (a *appInst) EditServer(w fyne.Window, index int, completionFunc func(removed bool)) {
	text := canvas.NewText("Edit server", fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	text.TextSize = 50

	var popup *widget.PopUp

	srv := a.servers[index]

	nameBox := utils.NewObjectWithLabel("Name: ", widget.NewEntry())
	ipBox := utils.NewObjectWithLabel("IP: ", widget.NewEntry())
	portBox := utils.NewObjectWithLabel("Port: ", widget.NewEntry())
	nameBox.Obj().SetText(srv.Name)
	ipBox.Obj().SetText(srv.IP)
	portBox.Obj().SetText(strconv.Itoa(int(srv.Port)))

	button := widget.NewButton("Save", func() {
		portNum, err := strconv.Atoi(portBox.Obj().Text)
		if err != nil {
			dialog.NewError(fmt.Errorf("failed to parse port: %w", err), w).Show()
			return
		}
		if portNum < 1 || portNum > 65535 {
			dialog.NewError(errors.New("port must be between 1 and 65535"), w).Show()
			return
		}
		srv := server{Name: nameBox.Obj().Text, IP: ipBox.Obj().Text, Port: uint16(portNum), LastPlayTime: time.Now()}
		a.servers[index] = srv

		completionFunc(false)

		popup.Hide()
	})

	delete := widget.NewButton("Delete", func() {
		a.servers = slices.Delete(a.servers, index, index+1)
		completionFunc(true)
		popup.Hide()
	})
	delete.Importance = widget.DangerImportance

	cancel := widget.NewButton("Cancel", func() {
		popup.Hide()
	})

	popup = widget.NewModalPopUp(container.NewCenter(container.NewVBox(container.New(utils.NewFixedSizeLayoutExpand(fyne.NewSize(350, 60)), container.NewCenter(text)), nameBox.Container(), ipBox.Container(), portBox.Container(), cancel, button, delete)), w.Canvas())

	popup.Show()
}
