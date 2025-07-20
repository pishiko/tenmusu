package window

import (
	"image"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/html"
	"golang.org/x/text/language"
)

type Drawable struct {
	word   string
	font   *text.GoTextFace
	x, y   float64
	style  string
	weight string
	w, h   float64
}

type Layout struct {
	tokens     []html.Token
	scrollY    float64
	fontSource FontSource
	screenRect image.Rectangle

	xCursor    float64
	yCursor    float64
	weight     string
	style      string
	size       float64
	line       []Drawable
	_drawables []Drawable
}

func NewLayout(tokens []html.Token, scrollY float64, fontSource FontSource, screenRect image.Rectangle) *Layout {
	return &Layout{
		tokens:     tokens,
		fontSource: fontSource,
		screenRect: screenRect,
		xCursor:    0.0,
		yCursor:    0.0 + scrollY,
		weight:     "normal",
		style:      "roman",
		size:       16.0,
		line:       []Drawable{},
		_drawables: []Drawable{},
	}
}

func (l *Layout) Drawables() []Drawable {
	for _, token := range l.tokens {
		switch token.Type {
		case html.Tag:
			switch token.Value {
			case "i":
				l.style = "italic"
			case "/i":
				l.style = "roman"
			case "em":
				l.style = "italic"
			case "/em":
				l.style = "roman"
			case "b":
				l.weight = "bold"
			case "/b":
				l.weight = "normal"
			case "strong":
				l.weight = "bold"
			case "/strong":
				l.weight = "normal"
			case "big":
				l.size += 4.0
			case "/big":
				l.size -= 4.0
			case "small":
				l.size -= 2.0
			case "/small":
				l.size += 2.0
			case "br":
				l.flush()
			case "/p":
				l.flush()
				l.yCursor += 16.0 // Add some space for paragraph
			}
		case html.Text:
			for _, word := range strings.Split(token.Value, " ") {
				if word == "" {
					continue // Skip empty words
				}
				source := l.fontSource.normal
				if l.weight == "bold" {
					source = l.fontSource.bold
				}
				f := &text.GoTextFace{
					Source:    source,
					Direction: text.DirectionLeftToRight,
					Size:      l.size,
					Language:  language.Japanese,
				}
				w, h := text.Measure(word, f, f.Metrics().HLineGap)

				// 右端まで言ったら改行
				if l.xCursor+w > float64(l.screenRect.Dx()) {
					l.flush()
				}

				l.line = append(l.line, Drawable{
					word:   word,
					font:   f,
					x:      l.xCursor,
					y:      l.yCursor,
					style:  l.style,
					weight: l.weight,
					w:      w,
					h:      h,
				})

				// Update x position for the next character
				l.xCursor += w
				spaceWidth, _ := text.Measure(" ", f, f.Metrics().HLineGap)
				l.xCursor += spaceWidth
			}
		}

	}
	return l._drawables
}

func (l *Layout) flush() {
	maxAscent := 0.0
	maxDescent := 0.0
	maxGap := 0.0
	for _, d := range l.line {
		metrics := d.font.Metrics()

		if metrics.HAscent > maxAscent {
			maxAscent = metrics.HAscent
		}
		if metrics.HDescent > maxDescent {
			maxDescent = metrics.HDescent
		}
		if metrics.HLineGap > maxGap {
			maxGap = metrics.HLineGap
		}
	}
	for _, d := range l.line {
		baseline := l.yCursor + maxAscent
		y := baseline - d.font.Metrics().HAscent
		l._drawables = append(l._drawables,
			Drawable{
				word:   d.word,
				font:   d.font,
				x:      d.x,
				y:      y,
				style:  d.style,
				weight: d.weight,
				w:      d.w,
				h:      d.h,
			},
		)
	}
	l.yCursor += (maxAscent + maxDescent) + maxGap
	l.line = []Drawable{}
	l.xCursor = 0
}
