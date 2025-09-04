package layout

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Drawable interface {
	Draw(screen *ebiten.Image, scrollY float64)
}

type TextDrawable struct {
	word   string
	font   *text.GoTextFace
	x, y   float64
	style  string
	weight string
	w, h   float64
	color  color.RGBA
}

func (d *TextDrawable) Draw(screen *ebiten.Image, scrollY float64) {
	d.y += scrollY
	if d.y+d.h < 0 || d.y > float64(screen.Bounds().Dy()) {
		return
	}

	// Draw the character
	op := &text.DrawOptions{}

	r, g, b, a := d.color.RGBA()
	colorScale := ebiten.ColorScale{}
	colorScale.SetR(float32(r) / 65535.0)
	colorScale.SetG(float32(g) / 65535.0)
	colorScale.SetB(float32(b) / 65535.0)
	colorScale.SetA(float32(a) / 65535.0)
	op.ColorScale = colorScale

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

func (d *RectDrawable) Draw(screen *ebiten.Image, scrollY float64) {
	d.top += scrollY
	d.bottom += scrollY
	if d.top > float64(screen.Bounds().Dy()) || d.bottom < 0 ||
		d.left > float64(screen.Bounds().Dx()) || d.right < 0 {
		return
	}

	width := d.right - d.left
	height := d.bottom - d.top
	if width <= 0 || height <= 0 {
		return
	}
	rectImage := ebiten.NewImage(int(width), int(height))
	rectImage.Fill(d.color)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(d.left, d.top)
	screen.DrawImage(rectImage, op)
}
