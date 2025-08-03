package css

import "github.com/pishiko/tenmusu/internal/parser/model"

type Selector interface {
	Matches(node *model.Node) bool
	Priority() int
}

type TagSelector struct {
	tag string
}

func (ts *TagSelector) Matches(n *model.Node) bool {
	return n.Type == model.Element && n.Value == ts.tag
}

func (ts *TagSelector) Priority() int {
	return 1
}

type DescendantSelector struct {
	ancestor   Selector
	descendant Selector
}

func (ds *DescendantSelector) Matches(node *model.Node) bool {
	if !ds.ancestor.Matches(node) {
		return false
	}
	for node.Parent != nil {
		if ds.ancestor.Matches(node.Parent) {
			return true
		}
		node = node.Parent
	}
	return false
}

func (ds *DescendantSelector) Priority() int {
	return ds.ancestor.Priority() + ds.descendant.Priority()
}
