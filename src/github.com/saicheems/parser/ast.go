package parser

import (
	"github.com/saicheems/token"
)

const (
	TypeProgram    = 0
	TypeBlock      = 1
	TypeConst      = 2
	TypeVar        = 3
	TypeProcedure  = 4
	TypeAssignment = 5
	TypeTerminal   = 6
)

// Represents a single node of the abstract syntax tree.
type AstNode struct {
	Type     int
	Tok      *token.Token
	Children []*AstNode
}

func newAstNode(t int) *AstNode {
	a := new(AstNode)
	a.Type = t
	a.Children = make([]*AstNode, 0)
	return a
}

func newAstNodeProgram() *AstNode {
	return newAstNode(TypeProgram)
}

func newAstNodeBlock(c *AstNode, v *AstNode, p *AstNode) *AstNode {
	a := newAstNode(TypeBlock)
	a.Children = append(a.Children, c, v, p)
	return a
}

func newAstNodeConst(v []*AstNode) *AstNode {
	a := newAstNode(TypeConst)
	a.Children = append(a.Children, v...)
	return a
}

func newAstNodeVar(v []*AstNode) *AstNode {
	a := newAstNode(TypeConst)
	a.Children = append(a.Children, v...)
	return a
}

// Returns a new astNode with TypeProcedure. Takes in an identifier node and a
// block node.
func newAstNodeProcedure(v *AstNode, b *AstNode) *AstNode {
	a := newAstNode(TypeProcedure)
	a.Children = append(a.Children, v, b)
	return a
}

func newAstNodeAssigment(a *AstNode, b *AstNode) *AstNode {
	x := newAstNode(TypeAssignment)
	x.Children = append(x.Children, a, b)
	return x
}

func newAstNodeTerminal(t *token.Token) *AstNode {
	a := newAstNode(TypeTerminal)
	a.Tok = t
	return a
}
