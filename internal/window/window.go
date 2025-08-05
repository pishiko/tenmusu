package window

import (
	_ "embed"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/pishiko/tenmusu/internal/parser/model"
)

type Window struct {
	node    *model.Node
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

	layout := NewDocumentLayout(b.node, screen.Bounds())
	layout.layout()
	for _, drawable := range layout.drawables {
		drawable.Draw(screen, float64(b.scrollY))
	}

}
func (b *Window) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func NewWindow(node *model.Node) *Window {
	return &Window{
		node:    node,
		scrollY: 0,
	}
}

func Open(node *model.Node) {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("tenmusu")
	if err := ebiten.RunGame(NewWindow(node)); err != nil {
		panic(err)
	}
	println("Exiting...")
}
