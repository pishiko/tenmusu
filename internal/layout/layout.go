package layout

import (
	"image"
	"image/color"
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
	// debugPrint(child, 0)
	// os.Exit(0)
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

type BlockLayout struct {
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

func (l BlockLayout) Prop() LayoutProperty {
	return l.prop
}

func (l *BlockLayout) Paint() []Drawable {
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

	if l.layoutMode() == Inline {
		for _, d := range l.drawables {
			ret = append(ret, &d)
		}
	}
	return ret
}

func (l *BlockLayout) Layout() {
	l.prop.x = l.parent.Prop().x
	l.prop.width = l.parent.Prop().width
	if l.previous != nil {
		l.prop.y = l.previous.Prop().y + l.previous.Prop().height
	} else {
		l.prop.y = l.parent.Prop().y
	}

	switch l.layoutMode() {
	case Block:
		previous := (*BlockLayout)(nil)
		for _, child := range l.node.Children {
			var next *BlockLayout
			if previous != nil {
				next = &BlockLayout{
					node:     child,
					parent:   l,
					previous: previous,
					children: []Layout{},
				}
			} else {
				next = &BlockLayout{
					node:     child,
					parent:   l,
					children: []Layout{},
				}
			}
			l.children = append(l.children, next)
			previous = next
		}
	case Inline:
		l.newLine()
		l.recurse(l.node)
	}
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

func (l *BlockLayout) PaintTree(drawables []Drawable) []Drawable {
	drawables = append(drawables, l.Paint()...)
	for _, child := range l.children {
		drawables = child.PaintTree(drawables)
	}
	return drawables
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

func (l *BlockLayout) layoutMode() LayoutMode {
	return getLayoutMode(l.node)
}

func (l *BlockLayout) recurse(node *model.Node) {
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

func (l *BlockLayout) openTag(tag string) {
	switch tag {
	case "br":
		// l.flush()
	}
}

func (l *BlockLayout) closeTag(tag string) {
	switch tag {
	}
}

func (l *BlockLayout) word(node *model.Node) {
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

func (l *BlockLayout) newLine() {
	l.cursorX = 0
	var newLine *LineLayout
	if len(l.children) > 0 {
		lastLine := l.children[len(l.children)-1]
		newLine = &LineLayout{
			node:     l.node,
			parent:   l,
			previous: lastLine,
		}
	} else {
		newLine = &LineLayout{
			node:   l.node,
			parent: l,
		}
	}
	l.children = append(l.children, newLine)
}

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
	previous Layout
	children []*TextLayout

	prop LayoutProperty
}

func (l *LineLayout) Layout() {

	l.prop.width = l.parent.Prop().width
	l.prop.x = l.parent.Prop().x

	if l.previous != nil {
		l.prop.y = l.previous.Prop().y + l.previous.Prop().height
	} else {
		l.prop.y = l.parent.Prop().y
	}

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

func debugPrint(layout Layout, indent int) {
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
	}
}
