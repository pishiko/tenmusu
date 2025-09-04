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

	previous := (Layout)(nil)
	for _, child := range l.node.Children {
		var next Layout
		if previous != nil {
			if isBlockLayout(child) {
				next = &BlockLayout{
					node:     child,
					parent:   l,
					previous: previous,
					children: []Layout{},
				}
			} else {
				next = &InlineLayout{
					node:     child,
					parent:   l,
					previous: previous,
					children: []Layout{},
				}
			}
		} else {
			if isBlockLayout(child) {
				next = &BlockLayout{
					node:     child,
					parent:   l,
					children: []Layout{},
				}
			} else {
				next = &InlineLayout{
					node:     child,
					parent:   l,
					children: []Layout{},
				}
			}
		}
		l.children = append(l.children, next)
		previous = next
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

func (l *BlockLayout) recurse(node *model.Node) {
	l.openTag(node.Value)
	for _, child := range node.Children {
		l.recurse(child)
	}
	l.closeTag(node.Value)

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
