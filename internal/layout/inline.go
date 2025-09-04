package layout

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/parser/css"
	"github.com/pishiko/tenmusu/internal/parser/model"
	"golang.org/x/text/language"
)

type InlineLayout struct {
	node     *model.Node
	parent   Layout
	previous Layout
	children []Layout

	prop LayoutProperty

	cursorX   float64
	weight    string
	size      float64
	drawables []TextDrawable
}

func (l InlineLayout) Prop() LayoutProperty {
	return l.prop
}

func (l *InlineLayout) Paint() []Drawable {
	ret := []Drawable{}
	// bgcolor
	bgcolor, ok := l.node.Style["background-color"]
	if !ok {
		bgcolor = "transparent"
	}
	if bgcolor != "transparent" {
		x2, y2 := l.prop.x+l.prop.width, l.prop.y+l.prop.height
		ret = append(ret, &RectDrawable{
			top:    l.prop.y,
			left:   l.prop.x,
			bottom: y2,
			right:  x2,
			color:  css.RGBA(bgcolor),
		})
	}

	for _, d := range l.drawables {
		ret = append(ret, &d)
	}

	return ret
}

func (l *InlineLayout) Layout() {
	l.prop.x = l.parent.Prop().x
	l.prop.width = l.parent.Prop().width
	if l.previous != nil {
		l.prop.y = l.previous.Prop().y + l.previous.Prop().height
	} else {
		l.prop.y = l.parent.Prop().y
	}

	l.newLine()
	l.recurse(l.node)

	for _, child := range l.children {
		child.Layout()
	}

	// Height
	height := 0.0
	for _, child := range l.children {
		height += child.Prop().height
	}
	l.prop.height = height
}

func (l *InlineLayout) PaintTree(drawables []Drawable) []Drawable {
	drawables = append(drawables, l.Paint()...)
	for _, child := range l.children {
		drawables = child.PaintTree(drawables)
	}
	return drawables
}

func (l *InlineLayout) recurse(node *model.Node) {
	switch node.Type {
	case model.Element:
		l.openTag(node.Value)
		for _, child := range node.Children {
			l.recurse(child)
		}
		l.closeTag(node.Value)
	case model.Text:
		l.word(node)
	}
}

func (l *InlineLayout) openTag(tag string) {
	switch tag {
	case "br":
		// l.flush()
	}
}

func (l *InlineLayout) closeTag(tag string) {
	switch tag {
	}
}

func (l *InlineLayout) word(node *model.Node) {
	if fs, ok := node.Style["font-size"]; ok {
		fspx, _ := strings.CutSuffix(fs, "px")
		fspxInt, _ := strconv.Atoi(fspx)
		l.size = float64(fspxInt)
	}

	for _, word := range strings.FieldsFunc(node.Value, unicode.IsSpace) {
		if word == "" {
			continue // Skip empty words
		}
		source := fontSource.normal
		if l.weight == "bold" {
			source = fontSource.bold
		}
		f := &text.GoTextFace{
			Source:    source,
			Direction: text.DirectionLeftToRight,
			Size:      l.size,
			Language:  language.Japanese,
		}
		w, _ := text.Measure(word, f, f.Metrics().HLineGap)

		// 右端まで言ったら改行
		if l.cursorX+w > l.prop.width {
			l.newLine()
		}

		line := l.children[len(l.children)-1].(*LineLayout)
		var previousWord *TextLayout
		if len(line.children) > 0 {
			previousWord = line.children[len(line.children)-1]
		}
		txt := &TextLayout{
			node:     node,
			word:     word,
			parent:   l,
			previous: previousWord,
		}
		line.children = append(line.children, txt)

		// Update x position for the next character
		l.cursorX += w
		spaceWidth, _ := text.Measure(" ", f, f.Metrics().HLineGap)
		l.cursorX += spaceWidth
	}
}

func (l *InlineLayout) newLine() {
	l.cursorX = 0
	newLine := &LineLayout{
		node:   l.node,
		parent: l,
	}
	l.children = append(l.children, newLine)
}
