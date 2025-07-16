package utils

import "fyne.io/fyne/v2"

type FixedSizeLayoutExpand struct {
	size fyne.Size
}

func NewFixedSizeLayoutExpand(size fyne.Size) *FixedSizeLayout {
	return &FixedSizeLayout{size: size}
}

func (m FixedSizeLayoutExpand) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	topLeft := fyne.NewPos(0, 0)
	for _, child := range objects {
		child.Resize(size)
		child.Move(topLeft)
	}
}

func (m FixedSizeLayoutExpand) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		minSize = minSize.Max(child.MinSize())
	}

	return minSize.Max(m.size)
}
