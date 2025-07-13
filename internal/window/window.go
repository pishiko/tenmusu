package window

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	ebitenutil.DebugPrint(screen, "FPS: "+fmt.Sprintf("%.2f", ebiten.ActualFPS()))

	y := 0 + b.scrollY // Initial Y position with scroll offset
	x := 0             // Initial X position
	for _, c := range b.text {
		w, h := text.Measure(string(c), b.fontFace, 48)

		// 右端まで言ったら改行
		if x+int(w) > screen.Bounds().Dx() {
			x = 0
			y += int(h)
		}
		// y < 0 はスキップ
		if y+int(h) < 0 {
			x += int(w) // Move to the next character
			continue
		}

		// y > 画面下以降は終了
		if y > screen.Bounds().Dy() {
			break
		}

		// Draw the character
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		text.Draw(screen, string(c), b.fontFace, op)

		// Update x position for the next character
		x += int(w)
	}
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
