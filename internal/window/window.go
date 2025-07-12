package window

import (
	"bytes"
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"
)

//go:embed resources/NotoSansJP-Regular.ttf
var ttf []byte

type Window struct {
	fontFace *text.GoTextFace

	text    string
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
	op := &text.DrawOptions{}
	op.GeoM.Translate(10, float64(10+b.scrollY))
	op.LineSpacing = 48
	text.Draw(screen, b.text, b.fontFace, op)
}
func (b *Window) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func NewWindow(body string) *Window {
	f, err := text.NewGoTextFaceSource(bytes.NewReader(ttf))
	if err != nil {
		panic(err)
	}
	face := &text.GoTextFace{
		Source:    f,
		Direction: text.DirectionLeftToRight,
		Size:      16,
		Language:  language.Japanese,
	}
	return &Window{
		fontFace: face,
		text:     body,
		scrollY:  0,
	}
}

func Open(content string) {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("tenmusu")
	if err := ebiten.RunGame(NewWindow(content)); err != nil {
		panic(err)
	}
	println("Exiting...")
}
