package window

import (
	"image"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/html"
	"golang.org/x/text/language"
)

type Display struct {
	op   *text.DrawOptions
	word string
	font *text.GoTextFace
}

func Layout(tokens []html.Token, scrollY float64, fontSource *text.GoTextFaceSource, screenRect image.Rectangle) []Display {
	displays := []Display{}

	// pos
	y := 0 + scrollY
	x := 0

	//style
	weight := "normal"
	style := "roman"

	for _, token := range tokens {
		switch token.Type {
		case html.Tag:
			switch token.Value {
			case "i":
				style = "italic"
			case "/i":
				style = "roman"
			case "b":
				weight = "bold"
			case "/b":
				weight = "normal"
			}
		case html.Text:
			for _, word := range strings.Split(token.Value, " ") {
				if word == "" {
					continue // Skip empty words
				}
				f := &text.GoTextFace{
					Source:    fontSource,
					Direction: text.DirectionLeftToRight,
					Size:      16,
					Language:  language.Japanese,
				}
				w, h := text.Measure(word, f, f.Metrics().HLineGap)

				// 右端まで言ったら改行
				if x+int(w) > screenRect.Dx() {
					x = 0
					y += h //todo fix
				}
				// y < 0 はスキップ
				if y+h < 0 {
					x += int(w)
					spaceWidth, _ := text.Measure(" ", f, f.Metrics().HLineGap)
					x += int(spaceWidth)
					continue
				}

				// y > 画面下以降は終了
				if y > float64(screenRect.Dy()) {
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

				displays = append(displays, Display{
					op:   op,
					word: word,
					font: f,
				})

				// Update x position for the next character
				x += int(w)
				spaceWidth, _ := text.Measure(" ", f, f.Metrics().HLineGap)
				x += int(spaceWidth)
			}
		}

	}
	return displays
}
