package window

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/html"
)

//go:embed resources/NotoSansJP-Regular.ttf
var ttf []byte

type Window struct {
	fontSource *text.GoTextFaceSource

	tokens  []html.Token
	scrollY int
}

func (b *Window) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		b.scrollY += 10
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		b.scrollY -= 10
	}
	return nil
}
func (b *Window) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "FPS: "+fmt.Sprintf("%.2f", ebiten.ActualFPS()))

	displays := Layout(b.tokens, float64(b.scrollY), b.fontSource, screen.Bounds())
	for _, display := range displays {
		text.Draw(screen, display.word, display.font, display.op)
	}

}
func (b *Window) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func NewWindow(tokens []html.Token) *Window {
	f, err := text.NewGoTextFaceSource(bytes.NewReader(ttf))
	if err != nil {
		panic(err)
	}
	return &Window{
		fontSource: f,
		tokens:     tokens,
		scrollY:    0,
	}
}

func Open(tokens []html.Token) {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("tenmusu")
	if err := ebiten.RunGame(NewWindow(tokens)); err != nil {
		panic(err)
	}
	println("Exiting...")
}
