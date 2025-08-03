package window

import (
	"image"
	"strconv"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/parser/css"
	"github.com/pishiko/tenmusu/internal/parser/model"
	"golang.org/x/text/language"
)

type LayoutMode int

const (
	Block LayoutMode = iota
	Inline
)

type DocumentLayout struct {
	node       model.Node
	screenRect image.Rectangle
	children   []*BlockLayout
	drawables  []Drawable
}

func NewDocumentLayout(node model.Node, screenRect image.Rectangle) *DocumentLayout {
	return &DocumentLayout{
		node:       node,
		screenRect: screenRect,
		drawables:  []Drawable{},
	}
}

func (l *DocumentLayout) layout() {
	parent := &BlockLayout{
		node:     nil,
		parent:   nil,
		previous: nil,
		x:        8.0,
		y:        8.0,
		width:    float64(l.screenRect.Dx()) - 16.0,
		height:   float64(l.screenRect.Dy()) - 16.0,
	}

	child := NewBlockLayout(&l.node, parent, nil)
	l.children = append(l.children, child)
	child.layout()
	l.drawables = []Drawable{}
	for _, child := range l.children {
		l.drawables = paintTree(child, l.drawables)
	}
}

func paintTree(layout *BlockLayout, drawables []Drawable) []Drawable {
	drawables = append(drawables, layout.paint()...)
	for _, child := range layout.children {
		drawables = paintTree(child, drawables)
	}
	return drawables
}

type BlockLayout struct {
	node     *model.Node
	parent   *BlockLayout
	previous *BlockLayout
	children []*BlockLayout

	x      float64
	y      float64
	width  float64
	height float64

	cursorX   float64
	cursorY   float64
	weight    string
	style     string
	size      float64
	line      []TextDrawable
	drawables []TextDrawable
}

func NewBlockLayout(node *model.Node, parent *BlockLayout, previous *BlockLayout) *BlockLayout {
	return &BlockLayout{
		node:     node,
		parent:   parent,
		previous: previous,
		children: []*BlockLayout{},
	}
}

func (l *BlockLayout) paint() []Drawable {
	ret := []Drawable{}
	// bgcolor
	bgcolor, ok := l.node.Style["background-color"]
	if !ok {
		bgcolor = "transparent"
	}
	if bgcolor != "transparent" {
		x2, y2 := l.x+l.width, l.y+l.height
		ret = append(ret, &RectDrawable{
			top:    l.y,
			left:   l.x,
			bottom: y2,
			right:  x2,
			color:  css.RGBA(bgcolor),
		})
	}

	if l.layoutMode() == Inline {
		for _, d := range l.drawables {
			ret = append(ret, &d)
		}
	}
	return ret
}

func (l *BlockLayout) layout() {
	l.x = l.parent.x
	l.width = l.parent.width
	if l.previous != nil {
		l.y = l.previous.y + l.previous.height
	} else {
		l.y = l.parent.y
	}

	switch l.layoutMode() {
	case Block:
		previous := (*BlockLayout)(nil)
		for _, child := range l.node.Children {
			next := NewBlockLayout(&child, l, previous)
			l.children = append(l.children, next)
			previous = next
		}
	case Inline:
		l.cursorX = 0.0
		l.cursorY = 0.0
		l.weight = "normal"
		l.style = "roman"
		l.size = 16.0
		l.line = []TextDrawable{}
		l.recurse(*l.node)
		l.flush()
	}
	for _, child := range l.children {
		child.layout()
	}

	// Height
	switch l.layoutMode() {
	case Block:
		height := 0.0
		for _, child := range l.children {
			height += child.height
		}
		l.height = height
	case Inline:
		l.height = l.cursorY
	}
}

func (l *BlockLayout) layoutMode() LayoutMode {
	switch l.node.Type {
	case model.Text:
		return Inline
	case model.Element:
		switch l.node.Value {
		case "html", "body", "article", "section", "nav", "aside",
			"h1", "h2", "h3", "h4", "h5", "h6", "hgroup", "header",
			"footer", "address", "p", "hr", "pre", "blockquote",
			"ol", "ul", "menu", "li", "dl", "dt", "dd", "figure",
			"figcaption", "main", "div", "table", "form", "fieldset",
			"legend", "details", "summary":
			return Block
		default:
			if len(l.node.Children) > 0 {
				return Inline
			}
		}
	}
	return Block
}

func (l *BlockLayout) recurse(node model.Node) {
	switch node.Type {
	case model.Element:
		l.openTag(node.Value)
		for _, child := range node.Children {
			l.recurse(child)
		}
		l.closeTag(node.Value)
	case model.Text:
		l.text(node)
	}
}

func (l *BlockLayout) openTag(tag string) {
	switch tag {
	case "em":
		l.style = "italic"
	case "strong":
		l.weight = "bold"
	case "br":
		l.flush()
	}
}

func (l *BlockLayout) closeTag(tag string) {
	switch tag {
	case "em":
		l.style = "roman"
	case "strong":
		l.weight = "normal"
	case "p":
		l.flush()
		l.cursorY += 16.0 // Add some space for paragraph
	}
}

func (l *BlockLayout) text(node model.Node) {

	if weight, ok := node.Style["font-weight"]; ok {
		l.weight = weight
	}
	if style, ok := node.Style["font-style"]; ok {
		l.style = style
	}
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
		w, h := text.Measure(word, f, f.Metrics().HLineGap)

		// 右端まで言ったら改行
		if l.cursorX+w > l.width {
			l.flush()
		}

		l.line = append(l.line, TextDrawable{
			word:   word,
			font:   f,
			x:      l.cursorX,
			y:      l.cursorY,
			style:  l.style,
			weight: l.weight,
			w:      w,
			h:      h,
		})

		// Update x position for the next character
		l.cursorX += w
		spaceWidth, _ := text.Measure(" ", f, f.Metrics().HLineGap)
		l.cursorX += spaceWidth
	}
}

func (l *BlockLayout) flush() {
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
		baseline := l.cursorY + maxAscent
		y := baseline - d.font.Metrics().HAscent
		l.drawables = append(l.drawables,
			TextDrawable{
				word:   d.word,
				font:   d.font,
				x:      l.x + d.x,
				y:      l.y + y,
				style:  d.style,
				weight: d.weight,
				w:      d.w,
				h:      d.h,
			},
		)
	}
	l.cursorY += (maxAscent + maxDescent) + maxGap
	l.line = []TextDrawable{}
	l.cursorX = 0
}
