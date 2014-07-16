// Package parser implements the parsing stage of the compilation including the
// parser and the abstract syntax tree.
package parser

import (
	"fmt"

	"github.com/saicheems/ast"
	"github.com/saicheems/lexer"
	"github.com/saicheems/token"
)

// Parser implements the parsing stage of the compilation.
type Parser struct {
	lex  *lexer.Lexer
	peek *token.Token // Next Token in the Token stream.
	err  []error      // Set errors if we have a parse failure.
}

// New returns a new Parser.
func New(l *lexer.Lexer) *Parser {
	p := new(Parser)
	p.lex = l
	// Initialize the error slice.
	p.err = make([]error, 0)
	p.move()
	return p
}

// Parse returns the head node of the abstract syntax tree. If there is an error in the parse it
// will return nil.
func (p *Parser) Parse() *ast.Node {
	block := p.parseBlock()
	// Expect a period to finish the program.
	p.expect(token.Period)
	// Print the first error if there are any and return nil (I'm not confident in the quality
	// of the errors produced yet - error recovery needs to be implemented.
	// TODO: Implement error recovery.
	if len(p.err) > 0 {
		fmt.Println(p.err[0])
		return nil
	}
	return ast.NewProgramNode(block)
}

// parseBlock parses blocks and returns a block Node.
func (p *Parser) parseBlock() *ast.Node {
	cons := p.parseConst()
	vars := p.parseVar()
	proc := p.parseProcedure()
	stmt := p.parseStatement()
	return ast.NewBlockNode(cons, vars, proc, stmt)
}

// parseConst parses consts and returns a const Node.
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

// parseVar parses vars and returns a var Node.
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

// parseProcedure parses procedures and returns a procedure Node.
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

// parseStatement parses all types of statement and returns the particular statement Node. Returns
// nil if no statement can be parsed.
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
		for {
			stmt := p.parseStatement()
			begin.AppendNode(stmt)
			p.expect(token.Semicolon)
			// If the next token can't begin a statement, stop looking for them.
			if !p.compareLookahead(token.Identifier, token.Call, token.Begin,
				token.If, token.While, token.Exclamation) {
				break
			}
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
	} else if p.accept(token.Exclamation) {
		expr := p.parseExpression()
		return ast.NewPrintNode(expr)
	} else {
		// If this function is called we expect to parse a statement.
		p.appendError()
		return nil
	}
}

// parseCondition parses conditions and returns a condition Node.
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

// parseExpression parses expressions and returns a math Node.
func (p *Parser) parseExpression() *ast.Node {
	op := int(token.Plus)
	var term *ast.Node

	if p.accept(token.Minus) {
		op = token.Minus
		term = ast.NewMathNode(op,
			ast.NewTerminalNode(&token.Token{Tag: token.Integer, Val: 0}),
			p.parseTerm())
	} else {
		p.accept(token.Plus)
		term = p.parseTerm()
	}
	for {
		if p.accept(token.Plus) {
			op = token.Plus
		} else if p.accept(token.Minus) {
			op = token.Minus
		} else {
			break
		}
		second := p.parseTerm()
		term = ast.NewMathNode(op, term, second)
	}
	return term
}

// parseTerm parses terms and returns a math Node.
func (p *Parser) parseTerm() *ast.Node {
	op := int(token.Times)
	fact := p.parseFactor()
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

// parseFactor parses factors and returns either a math Node or a terminal Node.
func (p *Parser) parseFactor() *ast.Node {
	iden := p.getTerminalNodeFromLookahead()
	if p.accept(token.Identifier) || p.accept(token.Integer) {
		return iden
	} else if p.accept(token.LeftParen) {
		expr := p.parseExpression()
		p.expect(token.RightParen)
		return expr
	} else {
		return nil
	}
}

// getTerminalNodeFromLookahead returns a Node containing the peek token if it is of type Integer or
// Identifier.
func (p *Parser) getTerminalNodeFromLookahead() *ast.Node {
	// Only return a node if the peekahead token is an actual terminal.
	if p.peek.Tag == token.Identifier || p.peek.Tag == token.Integer {
		return ast.NewTerminalNode(p.peek)
	}
	return nil
}

// move Moves the token stream forward by one token and sets the peek token.
func (p *Parser) move() {
	tok := p.lex.Scan()
	p.peek = tok
}

// compareLookahead takes in any number of tags and returns a bool representing whether or not any
// of those tags match the tag of the peek Token.
func (p *Parser) compareLookahead(t ...int) bool {
	for i := 0; i < len(t); i++ {
		if t[i] == p.peek.Tag {
			return true
		}
	}
	return false
}

// accept takes a tag and compares it to the tag of the peek token. If the two match, it moves the
// token stream forward and returns true. Otherwise it returns false.
func (p *Parser) accept(t int) bool {
	if p.peek.Tag == t {
		p.move()
		return true
	}
	return false
}

// expect takes a tag and compares it to the tag of the peek token. If the two match, it returns
// true. Otherwise it returns false and calls appendError. Regardless, it will move the token stream
// forward.
func (p *Parser) expect(t int) bool {
	acc := p.accept(t)
	if !acc {
		p.move()
		p.appendError()
	}
	return acc
}

// appendError adds a new "syntax" error to the err list.
func (p *Parser) appendError() {
	p.err = append(p.err, fmt.Errorf("Syntax error near line %d.", p.peek.Ln))
}
