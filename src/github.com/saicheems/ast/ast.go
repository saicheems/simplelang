// Package ast implements the abstract syntax tree node type.
package ast

import (
	"github.com/saicheems/symtable"
	"github.com/saicheems/token"
)

const (
	Program         = iota // The parent node.
	Block                  // Contains a set of statements.
	Const                  // ex. CONST a = 3, b = 4;
	Var                    // ex. VAR a, b;
	ProcedureParent        // Contains a set of procedure nodes.
	Procedure              // ex. PROCEDURE a; BLOCK
	Call                   // ex. CALL a;
	Begin                  // ex. BEGIN stmt END;
	IfThen                 // ex. IF cond THEN stmt;
	WhileDo                // ex. WHILE cond DO stmt;
	Odd                    // ex. ODD expr;
	Cond                   // ex. a == b; x # y;
	Math                   // Forms mathematical expressions.
	Assignment             // ex. a := 3;
	Terminal               // Contains a identifier token or an integer token.
	Print                  // ex. !X prints X.
)

// Represents a single node of the abstract syntax tree.
type Node struct {
	Tag      int                   // One of the constants defined by the package.
	Op       int                   // An operation: +, -, *, /, =, #, ...
	Tok      *token.Token          // Token for terminal nodes.
	Sym      *symtable.SymbolTable // Symbol table that encloses scope of this Node's children.
	Children []*Node               // Contains all children of this node.
}

// newNode returns a new Node. The argument is a tag set to one of the constants defined by the
// package..
func NewNode(t int) *Node {
	node := new(Node)
	node.Tag = t
	node.Children = make([]*Node, 0) // Initialize the children as an empty array.
	return node
}

// appendNode Appends a Node to the children of the Node it is called on.
func (n *Node) AppendNode(node ...*Node) {
	n.Children = append(n.Children, node...)
}

// NewProgramNode returns a new program Node given a block Node.
func NewProgramNode(block *Node) *Node {
	node := NewNode(Program)
	node.AppendNode(block)
	return node
}

// NewBlockNode returns a new block Node given const, var, procedure and statement Nodes.
func NewBlockNode(cons *Node, vars *Node, proc *Node, stmt *Node) *Node {
	node := NewNode(Block)
	node.AppendNode(cons, vars, proc, stmt)
	return node
}

// NewConstNode returns a new const Node.
func NewConstNode() *Node {
	node := NewNode(Const)
	return node
}

// NewVarNode returns a new var Node.
func NewVarNode() *Node {
	node := NewNode(Var)
	return node
}

// NewProcedureParentNode returns a new procedure parent Node. The procedure parent Node should
// enclose a set of procedure Nodes.
func NewProcedureParentNode() *Node {
	node := NewNode(ProcedureParent)
	return node
}

// NewProcedureNode Returns a new procedure Node given a terminal Node and a block Node.
func NewProcedureNode(iden *Node, bloc *Node) *Node {
	node := NewNode(Procedure)
	node.AppendNode(iden, bloc)
	return node
}

// NewCallNode returns a new call Node given a terminal Node.
func NewCallNode(iden *Node) *Node {
	node := NewNode(Call)
	node.AppendNode(iden)
	return node
}

// NewBeginNode returns a new begin Node.
func NewBeginNode() *Node {
	node := NewNode(Begin)
	return node
}

// NewIfThenNode returns a new if then Node given a condition Node and a statement Node.
func NewIfThenNode(cond *Node, stmt *Node) *Node {
	node := NewNode(IfThen)
	node.AppendNode(cond, stmt)
	return node
}

// NewWhileDoNode returns a new while do Node given a condition Node and a statement Node.
func NewWhileDoNode(cond *Node, stmt *Node) *Node {
	node := NewNode(WhileDo)
	node.AppendNode(cond, stmt)
	return node
}

// NewOddNode returns a new odd Node given an expression Node.
func NewOddNode(expr *Node) *Node {
	node := NewNode(Odd)
	node.AppendNode(expr)
	return node
}

// NewMathNode returns a new math Node given an operation, a left hand expression Node and a right
// hand expression Node.
func NewMathNode(op int, left *Node, right *Node) *Node {
	node := NewNode(Math)
	node.Op = op
	node.AppendNode(left, right)
	return node
}

// NewCondNode returns a new condition Node given an operation, a left hand epression Node and a
// right hand expression Node.
func NewCondNode(op int, left *Node, right *Node) *Node {
	node := NewNode(Cond)
	node.Op = op
	node.AppendNode(left, right)
	return node
}

// NewAssignmentNode returns a new assignment Node given a left hand terminal Node and a right hand
// expression Node.
func NewAssignmentNode(left *Node, right *Node) *Node {
	node := NewNode(Assignment)
	node.AppendNode(left, right)
	return node
}

// NewPrintNode returns a new print Node given an expression to print.
func NewPrintNode(expr *Node) *Node {
	node := NewNode(Print)
	node.AppendNode(expr)
	return node
}

// NewTerminalNode returns a new terminal Node given a terminal Token (Identifier or Integer).
func NewTerminalNode(tok *token.Token) *Node {
	node := NewNode(Terminal)
	node.Tok = tok
	return node
}
