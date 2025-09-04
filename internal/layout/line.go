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
	children []Layout
	prop     LayoutProperty

	word   string
	font   *text.GoTextFace
	weight string
	style  string
}

func (l *TextLayout) Layout() {
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

	if l.previous != nil {
		preFont := l.previous.font
		space, _ := text.Measure(" ", preFont, preFont.Metrics().HLineGap)
		l.prop.x = l.previous.Prop().x + l.previous.Prop().width + float64(space)
	}

}
func (l TextLayout) Prop() LayoutProperty {
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
	for _, child := range l.children {
		drawables = child.PaintTree(drawables)
	}
	return drawables
}

type LineLayout struct {
	node     *model.Node
	parent   Layout
	children []*TextLayout

	prop LayoutProperty
}

func (l *LineLayout) Layout() {

	l.prop.width = l.parent.Prop().width
	l.prop.x = l.parent.Prop().x

	l.prop.y = l.parent.Prop().y

	for _, child := range l.children {
		child.Layout()
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
}
func (l LineLayout) Prop() LayoutProperty {
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
