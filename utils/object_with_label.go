package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type WidgetWithLabel[T fyne.CanvasObject] struct {
	container *fyne.Container
	obj       T
}

func NewObjectWithLabel[T fyne.CanvasObject](label string, obj T) *WidgetWithLabel[T] {
	return &WidgetWithLabel[T]{container: container.NewBorder(nil, nil, widget.NewLabel(label), nil, obj), obj: obj}
}

func (e *WidgetWithLabel[T]) Container() *fyne.Container {
	return e.container
}

func (e *WidgetWithLabel[T]) Obj() T {
	return e.obj
}
