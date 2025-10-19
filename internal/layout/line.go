package layout

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/parser/css"
	"github.com/pishiko/tenmusu/internal/parser/model"
	"golang.org/x/text/language"
)

type TextLayout struct {
	node     *model.Node
	parent   Layout
	previous *TextLayout
	prop     Rect

	word   string
	font   *text.GoTextFace
	weight string
	style  string
}

func NewTextLayout(node *model.Node, word string) *TextLayout {
	t := &TextLayout{
		node: node,
		word: word,
	}
	t.calcWH()
	return t
}

func (l *TextLayout) Layout(x float64) {
	l.prop.x = x
}
func (l TextLayout) Prop() Rect {
	return l.prop
}
func (l *TextLayout) Paint() []Drawable {
	color := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if c, ok := l.node.Style["color"]; ok {
		color = css.RGBA(c)
	}
	return []Drawable{&TextDrawable{
		word:   l.word,
		font:   l.font,
		x:      l.prop.x,
		y:      l.prop.y,
		style:  l.style,
		weight: l.weight,
		w:      l.prop.width,
		h:      l.prop.height,
		color:  color,
	}}
}

func (l *TextLayout) PaintTree(drawables []Drawable) []Drawable {
	drawables = append(drawables, l.Paint()...)
	return drawables
}

func (l *TextLayout) calcWH() {
	l.weight, _ = l.node.Style["font-weight"]
	l.style, _ = l.node.Style["font-style"]

	fs, _ := l.node.Style["font-size"]
	fspx, _ := strings.CutSuffix(fs, "px")
	fspxInt, _ := strconv.Atoi(fspx)
	size := float64(fspxInt)

	source := fontSource.normal
	if l.weight == "bold" {
		source = fontSource.bold
	}
	l.font = &text.GoTextFace{
		Source:    source,
		Direction: text.DirectionLeftToRight,
		Size:      size,
		Language:  language.Japanese,
	}
	w, h := text.Measure(l.word, l.font, l.font.Metrics().HLineGap)
	l.prop.width = float64(w)
	l.prop.height = float64(h)
}

func (l *TextLayout) SpaceWidth() float64 {
	rs := []rune(l.word)
	if !isASCIIRune(rs[len(rs)-1]) {
		return 0
	}
	space, _ := text.Measure(" ", l.font, l.font.Metrics().HLineGap)
	return space
}

type LineLayout struct {
	parent   Layout
	previous Layout
	children []*TextLayout

	prop Rect
}

func (l *LineLayout) Layout(x float64, y float64, width float64) Rect {

	l.prop.width = width
	l.prop.x = x
	l.prop.y = y

	for i, child := range l.children {
		cx := l.prop.x
		if i > 0 {
			cx = l.children[i-1].prop.x + l.children[i-1].prop.width + l.children[i-1].SpaceWidth()
		}
		child.Layout(cx)
	}

	maxAscent := 0.0
	maxDescent := 0.0
	maxGap := 0.0
	for _, txt := range l.children {
		metrics := txt.font.Metrics()

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
	baseline := l.prop.y + maxAscent

	for _, txt := range l.children {
		txt.prop.y = baseline - txt.font.Metrics().HAscent
	}

	l.prop.height = (maxAscent + maxDescent) + maxGap
	return Rect{
		x:      l.prop.x,
		y:      l.prop.y,
		width:  l.prop.width,
		height: l.prop.height,
	}
}
func (l LineLayout) Prop() Rect {
	return l.prop
}
func (l *LineLayout) Paint() []Drawable {
	return []Drawable{}
}

func (l *LineLayout) PaintTree(drawables []Drawable) []Drawable {
	drawables = append(drawables, l.Paint()...)
	for _, child := range l.children {
		drawables = child.PaintTree(drawables)
	}
	return drawables
}

func (l *LineLayout) GetMinMaxWidth() (float64, float64) {
	panic("not implemented")
	return -1, -1
}
