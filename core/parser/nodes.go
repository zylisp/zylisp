package parser

import (
	"fmt"
	"go/token"
)

// Constants for the parser package
const (
	NodeIdent NodeType = iota
	NodeString
	NodeNumber
	NodeCall
	NodeVector
)

// Node holds data for parsed AST nodes
type Node interface {
	Type() NodeType
	// Position() Pos
	String() string
	Copy() Node
}

// NodeType defines a dedicated type for node types
type NodeType int

// Type is a convenience method for getting the NodeType
func (t NodeType) Type() NodeType {
	return t
}

// IdentNode defines the identity node for use in identity operations
type IdentNode struct {
	// Pos
	NodeType
	Ident string
}

// Copy creates a copy of an IdentNode instance
func (node *IdentNode) Copy() Node {
	return NewIdentNode(node.Ident)
}

// String returns a string represnetation of IdentNode
func (node *IdentNode) String() string {
	if node.Ident == "nil" {
		return "()"
	}

	return node.Ident
}

// StringNode object
type StringNode struct {
	// Pos
	NodeType
	Value string
}

// Copy creates a copy of a string node
func (node *StringNode) Copy() Node {
	return newStringNode(node.Value)
}

// String returns the string represnetation of a string node
func (node *StringNode) String() string {
	return node.Value
}

// NumberNode object
type NumberNode struct {
	// Pos
	NodeType
	Value      string
	NumberType token.Token
}

// Copy creates a copy of a number node
func (node *NumberNode) Copy() Node {
	return &NumberNode{NodeType: node.Type(), Value: node.Value, NumberType: node.NumberType}
}

// String returns the string value for a number node
func (node *NumberNode) String() string {
	return node.Value
}

// VectorNode object
type VectorNode struct {
	// Pos
	NodeType
	Nodes []Node
}

// Copy creates a copy of a vector node
func (node *VectorNode) Copy() Node {
	vect := &VectorNode{NodeType: node.Type(), Nodes: make([]Node, len(node.Nodes))}
	for i, v := range node.Nodes {
		vect.Nodes[i] = v.Copy()
	}
	return vect
}

// String returns the string value of a vector node
func (node *VectorNode) String() string {
	return fmt.Sprint(node.Nodes)
}

// CallNode object
type CallNode struct {
	// Pos
	NodeType
	Callee Node
	Args   []Node
}

// Copy creates a copy of a call node
func (node *CallNode) Copy() Node {
	call := &CallNode{NodeType: node.Type(), Callee: node.Callee.Copy(), Args: make([]Node, len(node.Args))}
	for i, v := range node.Args {
		call.Args[i] = v.Copy()
	}
	return call
}

// String returns the string value for a call node
func (node *CallNode) String() string {
	args := fmt.Sprint(node.Args)
	return fmt.Sprintf("(%s %s)", node.Callee, args[1:len(args)-1])
}

// Node constrcutors

// NewIdentNode creates a new identity node
func NewIdentNode(name string) *IdentNode {
	return &IdentNode{NodeType: NodeIdent, Ident: name}
}

func newStringNode(val string) *StringNode {
	return &StringNode{NodeType: NodeString, Value: val}
}

func newIntNode(val string) *NumberNode {
	return &NumberNode{NodeType: NodeNumber, Value: val, NumberType: token.INT}
}

func newFloatNode(val string) *NumberNode {
	return &NumberNode{NodeType: NodeNumber, Value: val, NumberType: token.FLOAT}
}

func newComplexNode(val string) *NumberNode {
	return &NumberNode{NodeType: NodeNumber, Value: val, NumberType: token.IMAG}
}

// We return Node here, because it could be that it's nil
func newCallNode(args []Node) Node {
	if len(args) > 0 {
		return &CallNode{NodeType: NodeCall, Callee: args[0], Args: args[1:]}
	}
	return nilNode
}

func newVectNode(content []Node) *VectorNode {
	return &VectorNode{NodeType: NodeVector, Nodes: content}
}