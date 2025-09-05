package layout

import (
	"image"
	"strconv"
	"strings"

	"github.com/pishiko/tenmusu/internal/parser/model"
)

type LayoutMode int

const (
	Block LayoutMode = iota
	Inline
)

type DocumentLayout struct {
	node       *model.Node
	screenRect image.Rectangle
	children   []Layout
	drawables  []Drawable
}

func NewDocumentLayout(node *model.Node, screenRect image.Rectangle) *DocumentLayout {
	return &DocumentLayout{
		node:       node,
		screenRect: screenRect,
		drawables:  []Drawable{},
	}
}

func (l *DocumentLayout) Layout() []Drawable {
	parent := &BlockLayout{
		prop: LayoutProperty{
			x:      8.0,
			y:      8.0,
			width:  float64(l.screenRect.Dx()) - 16.0,
			height: float64(l.screenRect.Dy()) - 16.0,
		},
	}

	child := &BlockLayout{
		node:     l.node,
		parent:   parent,
		children: []Layout{},
	}
	l.children = append(l.children, child)
	child.Layout()
	debugPrint(child, 0)
	l.drawables = []Drawable{}
	for _, child := range l.children {
		l.drawables = child.PaintTree(l.drawables)
	}
	return l.drawables
}

type LayoutProperty struct {
	x      float64
	y      float64
	width  float64
	height float64
}

type Layout interface {
	Layout()
	Paint() []Drawable
	Prop() LayoutProperty
	PaintTree([]Drawable) []Drawable
}

func getLayoutMode(node *model.Node) LayoutMode {
	switch node.Type {
	case model.Text:
		return Inline
	case model.Element:
		switch node.Value {
		case "html", "body", "article", "section", "nav", "aside",
			"h1", "h2", "h3", "h4", "h5", "h6", "hgroup", "header",
			"footer", "address", "p", "hr", "pre", "blockquote",
			"ol", "ul", "menu", "li", "dl", "dt", "dd", "figure",
			"figcaption", "main", "div", "table", "form", "fieldset",
			"legend", "details", "summary":
			return Block
		default:
			if len(node.Children) > 0 {
				return Inline
			}
		}
	}
	return Block
}

var debugPrinted = false

func debugPrint(layout Layout, indent int) {
	if debugPrinted {
		return
	}
	switch v := layout.(type) {
	case *BlockLayout:
		name := "Block"
		if v.layoutMode() == Inline {
			name = "Inline"
		}
		println(strings.Repeat("  ", indent) + name + "Layout: <" + v.node.Value + "> x=" + strconv.Itoa(int(v.prop.x)) +
			" y=" + strconv.Itoa(int(v.prop.y)) +
			" w=" + strconv.Itoa(int(v.prop.width)) +
			" h=" + strconv.Itoa(int(v.prop.height)))
	case *LineLayout:
		println(strings.Repeat("  ", indent) + "LineLayout: x=" + strconv.Itoa(int(v.prop.x)) +
			" y=" + strconv.Itoa(int(v.prop.y)) +
			" w=" + strconv.Itoa(int(v.prop.width)) +
			" h=" + strconv.Itoa(int(v.prop.height)))
	case *TextLayout:
		println(strings.Repeat("  ", indent) + "TextLayout: \"" + v.word + "\" x=" + strconv.Itoa(int(v.prop.x)) +
			" y=" + strconv.Itoa(int(v.prop.y)) +
			" w=" + strconv.Itoa(int(v.prop.width)) +
			" h=" + strconv.Itoa(int(v.prop.height)))
	case *InlineContext:
		println(strings.Repeat("  ", indent) + "[InlineContext]")
	}
	switch v := layout.(type) {
	case *BlockLayout:
		for _, child := range v.children {
			debugPrint(child, indent+1)
		}
	case *LineLayout:
		for _, child := range v.children {
			debugPrint(child, indent+1)
		}
	case *TextLayout:
		for _, child := range v.children {
			debugPrint(child, indent+1)
		}
	case *InlineContext:
		for _, child := range v.children {
			debugPrint(child, indent+1)
		}
	}
	if indent == 0 {
		debugPrinted = true
	}
}
