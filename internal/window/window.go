package window

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/html"
	"golang.org/x/text/language"
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

	// pos
	y := 0 + b.scrollY // Initial Y position with scroll offset
	x := 0             // Initial X position

	//style
	weight := "normal"
	style := "roman"

	for _, token := range b.tokens {
		switch token.Type {
		case html.Tag:
			if token.Value == "i" {
				style = "italic"
			} else if token.Value == "/i" {
				style = "roman"
			} else if token.Value == "b" {
				weight = "bold"
			} else if token.Value == "/b" {
				weight = "normal"
			}
		case html.Text:
			for _, word := range strings.Split(token.Value, " ") {
				if word == "" {
					continue // Skip empty words
				}
				f := &text.GoTextFace{
					Source:    b.fontSource,
					Direction: text.DirectionLeftToRight,
					Size:      16,
					Language:  language.Japanese,
				}
				w, h := text.Measure(word, f, f.Metrics().HLineGap)

				// 右端まで言ったら改行
				if x+int(w) > screen.Bounds().Dx() {
					x = 0
					y += int(h) //todo fix
				}
				// y < 0 はスキップ
				if y+int(h) < 0 {
					x += int(w)
					spaceWidth, _ := text.Measure(" ", f, f.Metrics().HLineGap)
					x += int(spaceWidth)
					continue
				}

				// y > 画面下以降は終了
				if y > screen.Bounds().Dy() {
					break
				}

				// Draw the character
				op := &text.DrawOptions{}

				if style == "italic" {
					// todo italic
					op.GeoM.Skew(-0.5, 0)
				}
				if weight == "bold" {
					// todo bold
				}
				op.GeoM.Translate(float64(x), float64(y))

				text.Draw(screen, word, f, op)

				// Update x position for the next character
				x += int(w)
				spaceWidth, _ := text.Measure(" ", f, f.Metrics().HLineGap)
				x += int(spaceWidth)
			}
		}

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
