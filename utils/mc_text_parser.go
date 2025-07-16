package utils

import (
	"image/color"
	"strings"
)

type TextStyle uint8

func (t TextStyle) SetBold() TextStyle {
	return t | (1 << 1)
}

func (t TextStyle) Bold() bool {
	return t&(1<<1) != 0
}

func (t TextStyle) SetItalic() TextStyle {
	return t | (1 << 2)
}

func (t TextStyle) Italic() bool {
	return t&(1<<2) != 0
}

func (t TextStyle) SetObfuscated() TextStyle {
	return t | (1 << 3)
}

func (t TextStyle) Obfuscated() bool {
	return t&(1<<3) != 0
}

type TextEntry struct {
	Color color.RGBA
	Style TextStyle
	Text  string
}

func ParseText(s string) (t []TextEntry) {
	var style TextStyle
	lastColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for i, s := range strings.Split(s, "§") {
		if len(s) == 0 {
			continue
		}
		c := s[0]
		e := TextEntry{Color: lastColor}
		if i == 0 {
			e.Text = s
			e.Color = lastColor
			e.Style = style
			t = append(t, e)
			continue
		} else {
			e.Text = s[1:]
		}
		switch c {
		case '0':
			e.Color = color.RGBA{A: 255}
		case '1':
			e.Color = color.RGBA{B: 170, A: 255}
		case '2':
			e.Color = color.RGBA{G: 170, A: 255}
		case '3':
			e.Color = color.RGBA{G: 170, B: 170, A: 255}
		case '4':
			e.Color = color.RGBA{R: 170, A: 255}
		case '5':
			e.Color = color.RGBA{R: 170, B: 170, A: 255}
		case '6':
			e.Color = color.RGBA{R: 255, G: 170, A: 255}
		case '7':
			e.Color = color.RGBA{R: 198, G: 198, B: 198, A: 255}
		case '8':
			e.Color = color.RGBA{R: 85, G: 85, B: 85, A: 255}
		case '9':
			e.Color = color.RGBA{R: 85, G: 85, B: 255, A: 255}
		case 'a':
			e.Color = color.RGBA{R: 85, G: 255, B: 85, A: 255}
		case 'b':
			e.Color = color.RGBA{R: 85, G: 255, B: 255, A: 255}
		case 'c':
			e.Color = color.RGBA{R: 255, G: 85, B: 85, A: 255}
		case 'd':
			e.Color = color.RGBA{R: 255, G: 85, B: 255, A: 255}
		case 'e':
			e.Color = color.RGBA{R: 255, G: 255, B: 85, A: 255}
		case 'f':
			e.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		case 'g':
			e.Color = color.RGBA{R: 221, G: 214, B: 5, A: 255}
		case 'h':
			e.Color = color.RGBA{R: 227, G: 212, B: 209, A: 255}
		case 'i':
			e.Color = color.RGBA{R: 206, G: 202, B: 202, A: 255}
		case 'j':
			e.Color = color.RGBA{R: 68, G: 58, B: 59, A: 255}
		case 'm':
			e.Color = color.RGBA{R: 51, G: 22, B: 7, A: 255}
		case 'n':
			e.Color = color.RGBA{R: 180, G: 104, B: 77, A: 255}
		case 'p':
			e.Color = color.RGBA{R: 222, G: 177, B: 45, A: 255}
		case 'q':
			e.Color = color.RGBA{R: 17, G: 159, B: 54, A: 255}
		case 's':
			e.Color = color.RGBA{R: 44, G: 186, B: 168, A: 255}
		case 't':
			e.Color = color.RGBA{R: 33, G: 73, B: 123, A: 255}
		case 'u':
			e.Color = color.RGBA{R: 154, G: 92, B: 198, A: 255}
		case 'v':
			e.Color = color.RGBA{R: 235, G: 114, B: 20, A: 255}
		case 'k':
			style = style.SetObfuscated()
		case 'l':
			style = style.SetBold()
		case 'o':
			style = style.SetItalic()
		case 'r':
			style = 0
			e.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		}
		lastColor = e.Color
		e.Style = style
		t = append(t, e)
	}
	return t
}
