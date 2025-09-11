package layout

import (
	"math"
	"strconv"

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
	return l.body.GetMinMaxWidth()
}

type TableRowGroupLayout struct {
	node   *model.Node
	parent *TableLayout
	rows   []*TableRowLayout
	prop   LayoutProperty

	initialized bool
}

func (l *TableRowGroupLayout) Init() {
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
	l.initialized = true
}

func (l *TableRowGroupLayout) Layout() {
	l.prop.x = l.parent.Prop().x
	l.prop.y = l.parent.Prop().y

	if !l.initialized {
		l.Init()
	}

	maxss, minss := [][]float64{}, [][]float64{}
	for _, row := range l.rows {
		mins, maxs := row.GetMinMaxWidth()

		minss = append(minss, mins)
		maxss = append(maxss, maxs)
	}
	maxWidths := maxSlices(maxss...)
	minWidths := maxSlices(minss...)
	maxSum := sumSlices(maxWidths)
	minSum := sumSlices(minWidths)
	l.prop.width = sumSlices(maxWidths)

	rowWidths := []float64{}

	l.prop.width = l.parent.Prop().width
	if l.prop.width <= minSum {
		// テーブル幅が最小幅以下なら最小幅に合わせる
		rowWidths = minWidths
	} else if l.prop.width >= maxSum {
		// テーブル幅が最大幅以上なら最大幅+余白を均等割り
		rowWidths = maxWidths
		extra := (l.prop.width - maxSum) / float64(len(maxWidths))
		for i := range rowWidths {
			rowWidths[i] += extra
		}
	} else {
		// 最小幅を確保して、残りを伸びやすさで分配
		rowWidths = minWidths
		flexibilities := make([]float64, len(maxWidths))
		for i := range flexibilities {
			flexibilities[i] = (maxWidths[i] - minWidths[i])
		}
		flexSum := sumSlices(flexibilities)
		extra := l.prop.width - minSum
		for i := range rowWidths {
			if flexSum > 0 {
				rowWidths[i] += extra * (flexibilities[i] / flexSum)
			}
		}
	}

	for _, row := range l.rows {
		row.Layout(rowWidths)
	}
	height := 0.0
	for _, row := range l.rows {
		height += row.prop.height
	}
	l.prop.height = height
}

func (l *TableRowGroupLayout) GetMinMaxWidth() (float64, float64) {
	if !l.initialized {
		l.Init()
	}
	maxss, minss := [][]float64{}, [][]float64{}
	for _, row := range l.rows {
		mins, maxs := row.GetMinMaxWidth()

		minss = append(minss, mins)
		maxss = append(maxss, maxs)
	}
	return sumSlices(maxSlices(minss...)), sumSlices(maxSlices(maxss...))
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

	initialized bool
}

func (l *TableRowLayout) Init() {
	l.prop.x = l.parent.prop.x
	if l.previous != nil {
		l.prop.y = l.previous.prop.y + l.previous.prop.height
	} else {
		l.prop.y = l.parent.prop.y
	}

	previous := (*TableCellLayout)(nil)
	i := 0
	for _, child := range l.node.Children {
		if child.Type == model.Element {
			switch child.Value {
			case "td", "th":
				if l.previous != nil {
					// rowspan考慮
					for {
						if i >= len(l.previous.cells) {
							break
						}
						aboveCell := l.previous.cells[i]
						if aboveCell.rowSpan <= 1 {
							break
						}
						// 上のセルが rowspan>1 している場合、ダミーセルを挿入して位置を合わせる
						joinedRoot := aboveCell
						for joinedRoot.joinedRowRoot != nil {
							joinedRoot = joinedRoot.joinedRowRoot
						}
						dummy := &TableCellLayout{
							node: &model.Node{
								Type:     model.Element,
								Value:    "td",
								Children: []*model.Node{},
							},
							parent:        l,
							previous:      previous,
							joinedRowRoot: joinedRoot,
							rowSpan:       aboveCell.rowSpan - 1,
						}
						l.cells = append(l.cells, dummy)
						previous = dummy
						i++
					}
				}
				// colspan考慮

				rowSpan := 1
				if a, ok := child.Attrs["rowspan"]; ok {
					n, err := strconv.Atoi(a)
					if err == nil {
						rowSpan = n
					}
				}
				colSpan := 1
				if a, ok := child.Attrs["colspan"]; ok {
					n, err := strconv.Atoi(a)
					if err == nil {
						colSpan = n
					}
				}

				cell := &TableCellLayout{
					node:     child,
					parent:   l,
					previous: previous,
					rowSpan:  rowSpan,
					colSpan:  colSpan,
				}
				l.cells = append(l.cells, cell)

				previous = cell
				for c := colSpan - 1; c >= 1; c-- {
					// colspan>1 の場合、ダミーセルを挿入して位置を合わせる
					dummy := &TableCellLayout{
						node: &model.Node{
							Type:     model.Element,
							Value:    "td",
							Children: []*model.Node{},
						},
						parent:        l,
						previous:      previous,
						joinedColRoot: cell,
						colSpan:       c,
					}
					l.cells[len(l.cells)-1].colNext = dummy
					l.cells = append(l.cells, dummy)
					previous = dummy
					i++
				}
			default:
				// TODO support col, colgroup or silently td/th
				panic("unsupported table row child: " + child.Value)
			}
		}
		i++
	}
	l.initialized = true
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
	}
	for _, cell := range l.cells {
		cell.Layout()
	}
	height := 0.0
	for _, cell := range l.cells {
		currentHeight := cell.prop.height
		// rowspan考慮
		// TODO FIX 均等に分配しているが正確ではない
		// 自身がrowspan>1の場合
		if cell.joinedRowRoot != nil {
			currentHeight = cell.joinedRowRoot.prop.height / float64(cell.joinedRowRoot.rowSpan)
		} else if cell.rowSpan > 1 {
			currentHeight = cell.prop.height / float64(cell.rowSpan)
		}
		if currentHeight > height {
			height = currentHeight
		}
	}
	l.prop.height = height
}

func (l *TableRowLayout) GetMinMaxWidth() (minWidths []float64, maxWidths []float64) {
	if !l.initialized {
		l.Init()
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

	joinedRowRoot *TableCellLayout
	joinedColRoot *TableCellLayout
	colNext       *TableCellLayout

	rowSpan int
	colSpan int
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
	for next := l.colNext; next != nil; next = next.colNext {
		l.prop.width += next.prop.width
	}
	// TODO FIX
	l.Init()
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
	// colspanのdummyのセルの場合均等に返す
	// TODO FIX 正確には均等ではない
	if l.joinedColRoot != nil {
		minWidth, maxWidth := l.joinedColRoot.GetMinMaxWidth()
		return minWidth / float64(l.joinedColRoot.colSpan), maxWidth / float64(l.joinedColRoot.colSpan)
	}
	minWidth := 0.0
	maxWidth := 0.0
	for _, child := range l.children {
		min, max := child.GetMinMaxWidth()
		minWidth = math.Max(minWidth, min)
		maxWidth = math.Max(maxWidth, max)
	}
	// colspan > 0の場合均等に返す
	// TODO FIX 正確には均等ではない
	if l.colSpan > 1 {
		return minWidth / float64(l.colSpan), maxWidth / float64(l.colSpan)
	}
	return minWidth, maxWidth
}
