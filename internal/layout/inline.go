package layout

import (
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/parser/css"
	"github.com/pishiko/tenmusu/internal/parser/model"
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
	for _, inline := range l.inlineItems {
		ret = append(ret, inline.PaintTree(ret)...)
	}
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

func (l *InlineContext) GetMinMaxWidth() (float64, float64) {
	minWidth := 0.0
	maxWidth := 0.0
	for i, child := range l.textItems {
		minWidth = math.Max(minWidth, child.Prop().width)
		maxWidth += child.Prop().width
		if i < len(l.children)-1 {
			maxWidth += child.SpaceWidth()
		}
	}
	return minWidth, maxWidth
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
		if len(lastLine.children) == 0 {
			// どうやっても収まらない場合ここに到達 TODO FIX
			return
		}
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
	parent   Layout
	children []*TextLayout

	size   float64
	weight string
}

func (l *InlineLayout) Paint() []Drawable {
	ret := []Drawable{}

	if len(l.children) == 0 {
		return ret
	}
	// bgcolor
	bgcolor, ok := l.node.Style["background-color"]
	if !ok {
		bgcolor = "transparent"
	}
	prevRight := l.children[0].prop.x
	prevTop := l.children[0].prop.y
	if bgcolor != "transparent" {
		for _, child := range l.children {
			if child.Prop().y > prevTop {
				prevRight = child.Prop().x
			}
			right, bottom := child.Prop().x+child.Prop().width, child.Prop().y+child.Prop().height
			ret = append(ret, &RectDrawable{
				top:    child.Prop().y,
				left:   prevRight,
				bottom: bottom,
				right:  right,
				color:  css.RGBA(bgcolor),
			})
			prevRight = right
			prevTop = child.Prop().y
		}
	}
	return ret
}

func (l *InlineLayout) PaintTree(drawables []Drawable) []Drawable {
	return l.Paint()
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
	l.size = l.parent.Prop().size
	if fs, ok := node.Style["font-size"]; ok {
		fspx, _ := strings.CutSuffix(fs, "px")
		fspxInt, _ := strconv.Atoi(fspx)
		l.size = float64(fspxInt)
	}

	for _, word := range split(node.Value) {
		if word == "" {
			continue // Skip empty words
		}
		txt := NewTextLayout(node, word)
		l.children = append(l.children, txt)
	}
}

func split(s string) []string {
	var result []string
	current := ""
	for _, r := range s {
		if unicode.IsSpace(r) {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else if !isASCIIRune(r) {
			if current != "" {
				result = append(result, current)
				current = ""
			}
			result = append(result, string(r))
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func isASCIIRune(r rune) bool {
	return r <= unicode.MaxASCII
}
