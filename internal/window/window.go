package window

import (
	_ "embed"
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/html"
)

type Window struct {
	node    html.Node
	scrollY int
}

func (b *Window) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		b.scrollY += 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		b.scrollY -= 5
	}
	return nil
}
func (b *Window) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	ebitenutil.DebugPrint(screen, "FPS: "+fmt.Sprintf("%.2f", ebiten.ActualFPS()))

	layout := NewDocumentLayout(b.node, float64(b.scrollY), screen.Bounds())
	layout.layout()
	for _, drawable := range layout.drawables {

		if drawable.y+drawable.h < 0 || drawable.y > float64(screen.Bounds().Dy()) {
			continue
		}

		// Draw the character
		op := &text.DrawOptions{}
		color := ebiten.ColorScale{}
		color.SetR(0)
		color.SetG(0)
		color.SetB(0)
		color.SetA(1)
		op.ColorScale = color

		if drawable.style == "italic" {
			const deg = -15
			ang := deg * math.Pi / 180
			op.GeoM.Skew(ang, 0)
		}
		op.GeoM.Translate(drawable.x, drawable.y)
		text.Draw(screen, drawable.word, drawable.font, op)
	}

}
func (b *Window) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func NewWindow(node html.Node) *Window {
	return &Window{
		node:    node,
		scrollY: 0,
	}
}

func Open(node html.Node) {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("tenmusu")
	if err := ebiten.RunGame(NewWindow(node)); err != nil {
		panic(err)
	}
	println("Exiting...")
}
