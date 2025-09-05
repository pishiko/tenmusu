package layout

import (
	"github.com/pishiko/tenmusu/internal/parser/css"
	"github.com/pishiko/tenmusu/internal/parser/model"
)

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

	inlineContext := (*InlineContext)(nil)
	previous := (Layout)(nil)
	for _, child := range l.node.Children {
		switch getLayoutMode(child) {
		case Block:
			if inlineContext != nil {
				inlineContext = nil
			}
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
		case Inline:
			if inlineContext == nil {
				inlineContext = &InlineContext{
					parent:      l,
					previous:    previous,
					children:    []*LineLayout{},
					inlineItems: []*InlineLayout{},
				}
				l.children = append(l.children, inlineContext)
				previous = inlineContext
			}
			inline := &InlineLayout{
				parent:   l,
				node:     child,
				children: []*TextLayout{},
			}
			inlineContext.inlineItems = append(inlineContext.inlineItems, inline)
		}
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

func (l *BlockLayout) layoutMode() LayoutMode {
	return getLayoutMode(l.node)
}
