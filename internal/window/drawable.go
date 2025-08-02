package window

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Drawable interface {
	Draw(screen *ebiten.Image)
}

type TextDrawable struct {
	word   string
	font   *text.GoTextFace
	x, y   float64
	style  string
	weight string
	w, h   float64
}

func (d *TextDrawable) Draw(screen *ebiten.Image) {
	if d.y+d.h < 0 || d.y > float64(screen.Bounds().Dy()) {
		return
	}

	// Draw the character
	op := &text.DrawOptions{}
	color := ebiten.ColorScale{}
	color.SetR(0)
	color.SetG(0)
	color.SetB(0)
	color.SetA(1)
	op.ColorScale = color

	if d.style == "italic" {
		const deg = -15
		ang := deg * math.Pi / 180
		op.GeoM.Skew(ang, 0)
	}
	op.GeoM.Translate(d.x, d.y)
	text.Draw(screen, d.word, d.font, op)
}

type RectDrawable struct {
	top, left, bottom, right float64
	color                    color.Color
}

func (d *RectDrawable) Draw(screen *ebiten.Image) {
	if d.top > float64(screen.Bounds().Dy()) || d.bottom < 0 ||
		d.left > float64(screen.Bounds().Dx()) || d.right < 0 {
		return
	}

	width := d.right - d.left
	height := d.bottom - d.top
	rectImage := ebiten.NewImage(int(width), int(height))
	rectImage.Fill(d.color)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(d.left, d.top)
	screen.DrawImage(rectImage, op)
}
