package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowAlert(canvas fyne.Canvas, label string, button string, closeFunc func()) {
	var popup *widget.PopUp
	popup = widget.NewModalPopUp(container.NewVBox(
		widget.NewLabel(label),
		widget.NewButton(button, func() {
			popup.Hide()
			closeFunc()
		}),
	), canvas)
	popup.Show()
}
