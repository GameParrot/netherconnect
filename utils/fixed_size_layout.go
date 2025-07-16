package utils

import "fyne.io/fyne/v2"

type FixedSizeLayout struct {
	size fyne.Size
}

func NewFixedSizeLayout(size fyne.Size) *FixedSizeLayout {
	return &FixedSizeLayout{size: size}
}

func (m FixedSizeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	topLeft := fyne.NewPos(0, 0)
	for _, child := range objects {
		child.Resize(size)
		child.Move(topLeft)
	}
}

func (m FixedSizeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return m.size
}
