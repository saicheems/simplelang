// Package parser implements the parsing stage of the compilation including the
// parser and the abstract syntax tree.
package parser

import (
	"fmt"

	"github.com/saicheems/ast"
	"github.com/saicheems/lexer"
	"github.com/saicheems/token"
)

type Parser struct {
	lex  *lexer.Lexer
	look *token.Token
	err  []error // Set error if we have a parse failure.
}

func New(l *lexer.Lexer) *Parser {
	p := new(Parser)
	p.lex = l
	// Initialize the error slice.
	p.err = make([]error, 0)
	p.move()
	return p
}

// Parse begins reading the token stream recursively and generates an abstract
// syntax tree (AST). It returns a pointer to the root node of an abstract
// syntax tree. If there is an error in the parse, the return value will be nil.
func (p *Parser) Parse() *ast.Node {
	block := p.parseBlock()
	// Expect a period.
	p.expect(token.Period)
	if len(p.err) > 0 {
		fmt.Println(p.err[0])
		return nil
	}
	return ast.NewProgramNode(block)
}

func (p *Parser) parseBlock() *ast.Node {
	cons := p.parseConst()
	vars := p.parseVar()
	proc := p.parseProcedure()
	stmt := p.parseStatement()
	return ast.NewBlockNode(cons, vars, proc, stmt)
}

// Parses consts and returns an AST node. Returns nil if there are no consts.
func (p *Parser) parseConst() *ast.Node {
	cons := ast.NewConstNode()
	if !p.accept(token.Const) {
		return cons
	}
	for {
		iden := p.getTerminalNodeFromLookahead()
		p.expect(token.Identifier)
		p.expect(token.Equals)
		inte := p.getTerminalNodeFromLookahead()
		p.expect(token.Integer)
		cons.AppendNode(ast.NewAssignmentNode(iden, inte))
		if !p.accept(token.Comma) {
			break
		}
	}
	p.expect(token.Semicolon)
	return cons
}

// Parses vars and returns an AST node. Returns nil if there are no vars.
func (p *Parser) parseVar() *ast.Node {
	vars := ast.NewVarNode()
	if !p.accept(token.Var) {
		return vars
	}
	for {
		iden := p.getTerminalNodeFromLookahead()
		p.expect(token.Identifier)
		vars.AppendNode(iden)
		if !p.accept(token.Comma) {
			break
		}
	}
	p.expect(token.Semicolon)
	return vars
}

// Parses procedures and returns an AST node. Returns nil if there are no procedures.
func (p *Parser) parseProcedure() *ast.Node {
	proc := ast.NewProcedureParentNode()
	if !p.accept(token.Procedure) {
		return proc
	}
	for {
		iden := p.getTerminalNodeFromLookahead()
		p.expect(token.Identifier)
		p.expect(token.Semicolon)
		bloc := p.parseBlock()
		p.expect(token.Semicolon)
		proc.AppendNode(ast.NewProcedureNode(iden, bloc))
		if !p.accept(token.Procedure) {
			break
		}
	}
	return proc
}

func (p *Parser) parseStatement() *ast.Node {
	iden := p.getTerminalNodeFromLookahead()
	if p.accept(token.Identifier) {
		p.expect(token.Assignment)
		expr := p.parseExpression()
		return ast.NewAssignmentNode(iden, expr)
	} else if p.accept(token.Call) {
		iden := p.getTerminalNodeFromLookahead()
		p.expect(token.Identifier)
		return ast.NewCallNode(iden)
	} else if p.accept(token.Begin) {
		begin := ast.NewBeginNode()
		stmt := p.parseStatement()
		begin.AppendNode(stmt)
		p.expect(token.Semicolon)
		for {
			if !p.compareLookahead(token.Identifier, token.Call, token.Begin,
				token.If, token.While) {
				break
			}
			stmt := p.parseStatement()
			begin.AppendNode(stmt)
			p.expect(token.Semicolon)
		}
		p.expect(token.End)
		return begin
	} else if p.accept(token.If) {
		cond := p.parseCondition()
		p.expect(token.Then)
		stmt := p.parseStatement()
		return ast.NewIfThenNode(cond, stmt)
	} else if p.accept(token.While) {
		cond := p.parseCondition()
		p.expect(token.Do)
		stmt := p.parseStatement()
		return ast.NewWhileDoNode(cond, stmt)
	} else {
		p.appendError()
		return nil
	}
}

func (p *Parser) parseCondition() *ast.Node {
	if p.accept(token.Odd) {
		expr := p.parseExpression()
		return ast.NewOddNode(expr)
	} else {
		left := p.parseExpression()
		equalOp := token.GreaterThanEqualTo
		if p.accept(token.Equals) {
			equalOp = token.Equals
		} else if p.accept(token.NotEquals) {
			equalOp = token.NotEquals
		} else if p.accept(token.LessThan) {
			equalOp = token.LessThan
		} else if p.accept(token.GreaterThan) {
			equalOp = token.GreaterThan
		} else if p.accept(token.LessThanEqualTo) {
			equalOp = token.LessThanEqualTo
		} else if p.expect(token.GreaterThanEqualTo) {
			equalOp = token.GreaterThanEqualTo
		}
		right := p.parseExpression()
		return ast.NewCondNode(equalOp, left, right)
	}
}

func (p *Parser) parseExpression() *ast.Node {
	op := int(token.Plus)
	p.accept(token.Plus)
	if p.accept(token.Minus) {
		op = token.Minus
	}
	term := ast.NewMathNode(op, ast.NewTerminalNode(&token.Token{Tag: token.Integer, Val: 0}), p.parseTerm())
	for {
		if p.accept(token.Plus) {
			op = token.Plus
		} else if p.accept(token.Minus) {
			op = token.Plus
		} else {
			break
		}
		second := p.parseTerm()
		term = ast.NewMathNode(op, term, second)
	}
	return term
}

func (p *Parser) parseTerm() *ast.Node {
	op := int(token.Times)
	fact := ast.NewMathNode(op, ast.NewTerminalNode(&token.Token{Tag: token.Integer, Val: 1}), p.parseFactor())
	for {
		if p.accept(token.Times) {
			op = token.Times
		} else if p.accept(token.Divide) {
			op = token.Divide
		} else {
			break
		}
		second := p.parseFactor()
		fact = ast.NewMathNode(op, fact, second)
	}
	return fact
}

func (p *Parser) parseFactor() *ast.Node {
	iden := p.getTerminalNodeFromLookahead()
	if p.accept(token.Identifier) || p.accept(token.Integer) {
		return iden
	} else {
		p.expect(token.LeftParen)
		expr := p.parseExpression()
		p.expect(token.RightParen)
		return expr
	}
}

// Returns an AST node containing the lookahead token if it is of type integer
// or identifier.
func (p *Parser) getTerminalNodeFromLookahead() *ast.Node {
	// Only return a node if the lookahead token is an actual terminal.
	if p.look.Tag == token.Identifier || p.look.Tag == token.Integer {
		return ast.NewTerminalNode(p.look)
	}
	return nil
}

// Moves the token stream forward by one token. Sets the lookahead token.
func (p *Parser) move() {
	tok := p.lex.Scan()
	p.look = tok
}

func (p *Parser) compareLookahead(t ...int) bool {
	for i := 0; i < len(t); i++ {
		if t[i] == p.look.Tag {
			return true
		}
	}
	return false
}

// Takes a tag and checks the token stream to see if it expectes. Returns true
// and advancess the input stream if so, otherwise returns false.
func (p *Parser) accept(t int) bool {
	if p.look.Tag == t {
		p.move()
		return true
	}
	return false
}

// Does the same thing as accept but raises an error and appends it to the
// parsers error list.
func (p *Parser) expect(t int) bool {
	acc := p.accept(t)
	if !acc {
		p.move()
		p.appendError()
	}
	return acc
}

func (p *Parser) appendError() {
	p.err = append(p.err, fmt.Errorf("Syntax error near line %d.", p.look.Ln))
}
