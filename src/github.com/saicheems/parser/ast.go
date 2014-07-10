package parser

import (
	"github.com/saicheems/symtable"
	"github.com/saicheems/token"
)

const (
	TypeProgram         = 0
	TypeBlock           = 1
	TypeConst           = 2
	TypeVar             = 3
	TypeProcedureParent = 4
	TypeProcedure       = 5
	TypeCall            = 6
	TypeBegin           = 7
	TypeIfThen          = 8
	TypeWhileDo         = 9
	TypeOdd             = 10
	TypeCond            = 11
	TypeMath            = 12
	TypeAssignment      = 13
	TypeTerminal        = 14
)

// Represents a single node of the abstract syntax tree.
type Node struct {
	Type     int
	Op       int
	Tok      *token.Token
	Sym      *symtable.SymbolTable
	Children []*Node
}

// Creates a new node given an input type.
func newNode(t int) *Node {
	node := new(Node)
	node.Type = t
	node.Children = make([]*Node, 0)
	return node
}

// Appends a node to the children of the parent node.
func (n *Node) appendNode(node ...*Node) {
	n.Children = append(n.Children, node...)
}

// Returns a new program node. Argument is a block node.
func newProgramNode(block *Node) *Node {
	node := newNode(TypeProgram)
	node.appendNode(block)
	return node
}

// Returns a new block node. Arguments are const, var, procedure and stmt nodes.
func newBlockNode(cons *Node, vars *Node, proc *Node, stmt *Node) *Node {
	node := newNode(TypeBlock)
	node.appendNode(cons, vars, proc, stmt)
	return node
}

// Returns a new const node.
func newConstNode() *Node {
	node := newNode(TypeConst)
	return node
}

// Returns a new var node.
func newVarNode() *Node {
	node := newNode(TypeVar)
	return node
}

// Returns a new procedure parent node.
func newProcedureParentNode() *Node {
	node := newNode(TypeProcedureParent)
	return node
}

// Returns a new procedure node. Arguments are a terminal node and a block
// node.
func newProcedureNode(iden *Node, bloc *Node) *Node {
	node := newNode(TypeProcedure)
	node.appendNode(iden, bloc)
	return node
}

// Returns a new call node. Argument is a terminal node.
func newCallNode(iden *Node) *Node {
	node := newNode(TypeCall)
	node.appendNode(iden)
	return node
}

// Returns a new begin node.
func newBeginNode() *Node {
	node := newNode(TypeBegin)
	return node
}

// Returns a new if then node. Arguments are a condition and a statement.
func newIfThenNode(cond *Node, stmt *Node) *Node {
	node := newNode(TypeIfThen)
	node.appendNode(cond, stmt)
	return node
}

// Returns a new while do node. Arguments are a condition and a statement.
func newWhileDoNode(cond *Node, stmt *Node) *Node {
	node := newNode(TypeWhileDo)
	node.appendNode(cond, stmt)
	return node
}

// Returns a new odd node. Argument is an expression.
func newOddNode(expr *Node) *Node {
	node := newNode(TypeOdd)
	node.appendNode(expr)
	return node
}

// Returns a new math node. Arguments are an operation, a left hand and right
// hand side.
func newMathNode(op int, left *Node, right *Node) *Node {
	node := newNode(TypeMath)
	node.Op = op
	node.appendNode(left, right)
	return node
}

// Returns a new condition node. Arguments are an op type, the left side expr
// and the right side expr.
func newCondNode(op int, left *Node, right *Node) *Node {
	node := newNode(TypeCond)
	node.Op = op
	node.appendNode(left, right)
	return node
}

// Returns a new assignment node. Arguments are the left and right side of the
// assignment.
func newAssignmentNode(left *Node, right *Node) *Node {
	node := newNode(TypeAssignment)
	node.appendNode(left, right)
	return node
}

func newTerminalNode(tok *token.Token) *Node {
	node := newNode(TypeTerminal)
	node.Tok = tok
	return node
}
