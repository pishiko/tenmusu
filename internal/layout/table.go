package layout

import (
	"math"

	"github.com/pishiko/tenmusu/internal/parser/model"
)

type TableLayout struct {
	prop     LayoutProperty
	parent   Layout
	previous Layout
	node     *model.Node

	body *TableRowGroupLayout
}

func (l *TableLayout) Layout() {
	l.prop.x = l.parent.Prop().x
	if l.previous != nil {
		l.prop.y = l.previous.Prop().y + l.previous.Prop().height
	} else {
		l.prop.y = l.parent.Prop().y
	}
	l.prop.width = l.parent.Prop().width

	for _, child := range l.node.Children {
		if child.Type == model.Element {
			switch child.Value {
			case "tbody":
				l.body = &TableRowGroupLayout{
					node:   child,
					parent: l,
				}
			case "tr":
				// Implicit tbody
				if l.body == nil {
					l.body = &TableRowGroupLayout{
						node: &model.Node{
							Type:     model.Element,
							Children: l.node.Children,
							Value:    "tbody",
						},
						parent: l,
					}
				}
			default:
				// TODO support thead, tfoot or silently tbody
				panic("unsupported table child: " + child.Value)
			}
		}
	}
	if l.body != nil {
		l.body.Layout()
	}
	l.prop.height = l.body.prop.height
}

func (l *TableLayout) Prop() LayoutProperty {
	return l.prop
}

func (l *TableLayout) PaintTree(drawables []Drawable) []Drawable {
	return l.body.PaintTree(drawables)
}

func (l *TableLayout) GetMinMaxWidth() (float64, float64) {
	panic("not implemented")
	return -1, -1
}

type TableRowGroupLayout struct {
	node   *model.Node
	parent *TableLayout
	rows   []*TableRowLayout
	prop   LayoutProperty
}

func (l *TableRowGroupLayout) Layout() {
	l.prop.x = l.parent.Prop().x
	l.prop.y = l.parent.Prop().y

	previous := (*TableRowLayout)(nil)
	for _, child := range l.node.Children {
		if child.Type == model.Element && child.Value == "tr" {
			row := &TableRowLayout{
				node:     child,
				parent:   l,
				previous: previous,
			}
			l.rows = append(l.rows, row)
			previous = row
		}
	}

	maxss, minss := [][]float64{}, [][]float64{}
	for _, row := range l.rows {
		mins, maxs := row.Init()

		minss = append(minss, mins)
		maxss = append(maxss, maxs)
	}
	maxWidths := maxSlices(maxss...)
	l.prop.width = sumSlices(maxWidths)

	// TODO ちゃんと計算する、Blockにも対応したい
	for _, row := range l.rows {
		row.Layout(maxWidths)
	}
	height := 0.0
	for _, row := range l.rows {
		height += row.prop.height
	}
	l.prop.height = height
}

func (l *TableRowGroupLayout) PaintTree(drawables []Drawable) []Drawable {
	for _, row := range l.rows {
		drawables = row.PaintTree(drawables)
	}
	return drawables
}

func sumSlices(slice []float64) float64 {
	sum := 0.0
	for _, v := range slice {
		sum += v
	}
	return sum
}
func maxSlices(slices ...[]float64) []float64 {
	// 最長スライスの長さを取得
	maxLen := 0
	for _, s := range slices {
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}

	// 結果スライスを作成（初期値は NaN）
	result := make([]float64, maxLen)
	for i := range result {
		result[i] = math.NaN()
	}

	for i := 0; i < maxLen; i++ {
		for _, s := range slices {
			if i < len(s) {
				if math.IsNaN(result[i]) || s[i] > result[i] {
					result[i] = s[i]
				}
			}
		}
	}

	return result
}

type TableRowLayout struct {
	node     *model.Node
	parent   *TableRowGroupLayout
	previous *TableRowLayout
	cells    []*TableCellLayout
	prop     LayoutProperty
}

func (l *TableRowLayout) Init() (minWidths []float64, maxWidths []float64) {
	l.prop.x = l.parent.prop.x
	if l.previous != nil {
		l.prop.y = l.previous.prop.y + l.previous.prop.height
	} else {
		l.prop.y = l.parent.prop.y
	}

	previous := (*TableCellLayout)(nil)
	for _, child := range l.node.Children {
		if child.Type == model.Element {
			switch child.Value {
			case "td", "th":
				cell := &TableCellLayout{
					node:     child,
					parent:   l,
					previous: previous,
				}
				l.cells = append(l.cells, cell)
				previous = cell
			default:
				// TODO support col, colgroup or silently td/th
				panic("unsupported table row child: " + child.Value)
			}
		}
	}
	maxWidths = make([]float64, len(l.cells))
	minWidths = make([]float64, len(l.cells))

	for i, cell := range l.cells {
		cell.Init()
		minW, maxW := cell.GetMinMaxWidth()
		maxWidths[i] = maxW
		minWidths[i] = minW
	}

	return minWidths, maxWidths
}

func (l *TableRowLayout) Layout(widths []float64) {
	l.prop.x = l.parent.prop.x
	if l.previous != nil {
		l.prop.y = l.previous.prop.y + l.previous.prop.height
	} else {
		l.prop.y = l.parent.prop.y
	}
	l.prop.width = l.parent.prop.width
	for i, cell := range l.cells {
		cell.prop.width = widths[i]
		cell.Layout()
	}
	height := 0.0
	for _, cell := range l.cells {
		if cell.prop.height > height {
			height = cell.prop.height
		}
	}
	l.prop.height = height
}

func (l *TableRowLayout) PaintTree(drawables []Drawable) []Drawable {
	for _, cell := range l.cells {
		drawables = cell.PaintTree(drawables)
	}
	return drawables
}

type TableCellLayout struct {
	node     *model.Node
	parent   *TableRowLayout
	previous *TableCellLayout
	prop     LayoutProperty
	children []Layout
}

func (l *TableCellLayout) Init() {
	if l.previous != nil {
		l.prop.x = l.previous.prop.x + l.previous.prop.width
	} else {
		l.prop.x = l.parent.prop.x
	}
	l.prop.y = l.parent.prop.y

	l.children = createLayoutFromNodes(l.node.Children, l)
	for _, l := range l.children {
		l.Layout()
	}
}

func (l *TableCellLayout) Layout() {
	l.prop.y = l.parent.prop.y
	if l.previous != nil {
		l.prop.x = l.previous.prop.x + l.previous.prop.width
	} else {
		l.prop.x = l.parent.prop.x
	}
	l.children = createLayoutFromNodes(l.node.Children, l)
	for _, child := range l.children {
		child.Layout()
	}
	height := 0.0
	for _, child := range l.children {
		height += child.Prop().height
	}
	l.prop.height = height
}

func (l TableCellLayout) Prop() LayoutProperty {
	return l.prop
}

func (l *TableCellLayout) PaintTree(drawables []Drawable) []Drawable {
	for _, child := range l.children {
		drawables = child.PaintTree(drawables)
	}
	return drawables
}

func (l *TableCellLayout) GetMinMaxWidth() (float64, float64) {
	minWidth := 0.0
	maxWidth := 0.0
	for _, child := range l.children {
		min, max := child.GetMinMaxWidth()
		minWidth = math.Max(minWidth, min)
		maxWidth = math.Max(maxWidth, max)
	}
	return minWidth, maxWidth
}
