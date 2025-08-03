package model

type NodeType int

const (
	Text NodeType = iota
	Element
)

type Node struct {
	Type     NodeType
	Value    string
	Children []Node
	Parent   *Node
	Attrs    map[string]string
	Style    map[string]string
}
