package layout

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/parser/model"
	"golang.org/x/text/language"
)

func NewInlineContext(inlineItems []*InlineLayout, parent Layout, previous Layout) *InlineContext {
	return &InlineContext{
		inlineItems: inlineItems,
		parent:      parent,
		previous:    previous,
	}
}

type InlineContext struct {
	parent   Layout
	previous Layout
	children []*LineLayout

	inlineItems []*InlineLayout
	textItems   []*TextLayout
	prop        LayoutProperty

	cursorX   float64
	drawables []TextDrawable
}

func (l InlineContext) Prop() LayoutProperty {
	return l.prop
}

func (l *InlineContext) Paint() []Drawable {
	ret := []Drawable{}
	// // bgcolor
	// bgcolor, ok := l.node.Style["background-color"]
	// if !ok {
	// 	bgcolor = "transparent"
	// }
	// if bgcolor != "transparent" {
	// 	x2, y2 := l.prop.x+l.prop.width, l.prop.y+l.prop.height
	// 	ret = append(ret, &RectDrawable{
	// 		top:    l.prop.y,
	// 		left:   l.prop.x,
	// 		bottom: y2,
	// 		right:  x2,
	// 		color:  css.RGBA(bgcolor),
	// 	})
	// }

	return ret
}

func (l *InlineContext) Layout() {
	l.prop.x = l.parent.Prop().x
	l.prop.width = l.parent.Prop().width
	if l.previous != nil {
		l.prop.y = l.previous.Prop().y + l.previous.Prop().height
	} else {
		l.prop.y = l.parent.Prop().y
	}

	l.newLine()

	for _, item := range l.inlineItems {
		l.textItems = append(l.textItems, item.Items()...)
	}
	l.word()

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

func (l *InlineContext) PaintTree(drawables []Drawable) []Drawable {
	drawables = append(drawables, l.Paint()...)
	for _, child := range l.children {
		drawables = child.PaintTree(drawables)
	}
	return drawables
}

func (l *InlineContext) word() {
	for _, txt := range l.textItems {
		// 右端まで言ったら改行
		if l.cursorX+txt.prop.width > l.prop.width {
			l.newLine()
		}

		line := l.children[len(l.children)-1]
		var previousWord *TextLayout
		if len(line.children) > 0 {
			previousWord = line.children[len(line.children)-1]
		}
		txt.parent = l
		txt.previous = previousWord
		line.children = append(line.children, txt)

		// Update x position for the next character
		l.cursorX += txt.prop.width
		spaceWidth, _ := text.Measure(" ", txt.font, txt.font.Metrics().HLineGap)
		l.cursorX += spaceWidth
	}
}

func (l *InlineContext) newLine() {
	l.cursorX = 0
	var newLine *LineLayout
	if len(l.children) > 0 {
		lastLine := l.children[len(l.children)-1]
		newLine = &LineLayout{
			parent:   l,
			previous: lastLine,
		}
	} else {
		newLine = &LineLayout{
			parent: l,
		}
	}
	l.children = append(l.children, newLine)
}

type InlineLayout struct {
	node     *model.Node
	prop     LayoutProperty
	parent   *BlockLayout
	children []*TextLayout

	size   float64
	weight string
}

// func (l InlineLayout) Prop() LayoutProperty {
// 	return l.prop
// }

// func (l InlineLayout) Layout() {

// }

func (l *InlineLayout) Paint() []Drawable {
	return []Drawable{}
}

func (l *InlineLayout) PaintTree(drawables []Drawable) []Drawable {
	return drawables
}

func (l *InlineLayout) Items() []*TextLayout {
	l.recurse(l.node)
	return l.children
}

func (l *InlineLayout) recurse(node *model.Node) {
	switch node.Type {
	case model.Element:
		// l.openTag(node.Value)
		for _, child := range node.Children {
			l.recurse(child)
		}
		// l.closeTag(node.Value)
	case model.Text:
		l.word(node)
	}
}

func (l *InlineLayout) word(node *model.Node) {
	l.size = l.parent.size
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

		txt := &TextLayout{
			node: node,
			word: word,
			font: f,
		}
		txt.prop.width = float64(w)
		l.children = append(l.children, txt)
	}
}
