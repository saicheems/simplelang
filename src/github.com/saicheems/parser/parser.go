// Package parser implements the parsing stage of the compilation including the
// parser and the abstract syntax tree.
package parser

import (
	"fmt"

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
func (p *Parser) Parse() *Node {
	block := p.parseBlock()
	// Expect a period.
	p.expect(token.TagPeriod)
	if len(p.err) > 0 {
		fmt.Println(p.err[0])
		return nil
	}
	return newProgramNode(block)
}

func (p *Parser) parseBlock() *Node {
	cons := p.parseConst()
	vars := p.parseVar()
	proc := p.parseProcedure()
	stmt := p.parseStatement()
	return newBlockNode(cons, vars, proc, stmt)
}

// Parses consts and returns an AST node. Returns nil if there are no consts.
func (p *Parser) parseConst() *Node {
	cons := newConstNode()
	if !p.accept(token.TagConst) {
		return cons
	}
	for {
		iden := p.getTerminalNodeFromLookahead()
		p.expect(token.TagIdentifier)
		p.expect(token.TagEquals)
		inte := p.getTerminalNodeFromLookahead()
		p.expect(token.TagInteger)
		cons.appendNode(newAssignmentNode(iden, inte))
		if !p.accept(token.TagComma) {
			break
		}
	}
	p.expect(token.TagSemicolon)
	return cons
}

// Parses vars and returns an AST node. Returns nil if there are no vars.
func (p *Parser) parseVar() *Node {
	vars := newVarNode()
	if !p.accept(token.TagVar) {
		return vars
	}
	for {
		iden := p.getTerminalNodeFromLookahead()
		p.expect(token.TagIdentifier)
		vars.appendNode(iden)
		if !p.accept(token.TagComma) {
			break
		}
	}
	p.expect(token.TagSemicolon)
	return vars
}

// Parses procedures and returns an AST node. Returns nil if there are no procedures.
func (p *Parser) parseProcedure() *Node {
	proc := newProcedureParentNode()
	if !p.accept(token.TagProcedure) {
		return proc
	}
	for {
		iden := p.getTerminalNodeFromLookahead()
		p.expect(token.TagIdentifier)
		p.expect(token.TagSemicolon)
		bloc := p.parseBlock()
		p.expect(token.TagSemicolon)
		proc.appendNode(newProcedureNode(iden, bloc))
		if !p.accept(token.TagProcedure) {
			break
		}
	}
	return proc
}

func (p *Parser) parseStatement() *Node {
	iden := p.getTerminalNodeFromLookahead()
	if p.accept(token.TagIdentifier) {
		p.expect(token.TagAssignment)
		expr := p.parseExpression()
		return newAssignmentNode(iden, expr)
	} else if p.accept(token.TagCall) {
		iden := p.getTerminalNodeFromLookahead()
		p.expect(token.TagIdentifier)
		return newCallNode(iden)
	} else if p.accept(token.TagBegin) {
		begin := newBeginNode()
		stmt := p.parseStatement()
		begin.appendNode(stmt)
		p.expect(token.TagSemicolon)
		for {
			if !p.compareLookahead(token.TagIdentifier, token.TagCall, token.TagBegin,
				token.TagIf, token.TagWhile) {
				break
			}
			stmt := p.parseStatement()
			begin.appendNode(stmt)
			p.expect(token.TagSemicolon)
		}
		p.expect(token.TagEnd)
		return begin
	} else if p.accept(token.TagIf) {
		cond := p.parseCondition()
		p.expect(token.TagThen)
		stmt := p.parseStatement()
		return newIfThenNode(cond, stmt)
	} else if p.accept(token.TagWhile) {
		cond := p.parseCondition()
		p.expect(token.TagDo)
		stmt := p.parseStatement()
		return newWhileDoNode(cond, stmt)
	} else {
		p.appendError()
		return nil
	}
}

func (p *Parser) parseCondition() *Node {
	if p.accept(token.TagOdd) {
		expr := p.parseExpression()
		return newOddNode(expr)
	} else {
		left := p.parseExpression()
		equalOp := token.TagGreaterThanEqualTo
		if p.accept(token.TagEquals) {
			equalOp = token.TagEquals
		} else if p.accept(token.TagNotEquals) {
			equalOp = token.TagNotEquals
		} else if p.accept(token.TagLessThan) {
			equalOp = token.TagLessThan
		} else if p.accept(token.TagGreaterThan) {
			equalOp = token.TagGreaterThan
		} else if p.accept(token.TagLessThanEqualTo) {
			equalOp = token.TagLessThanEqualTo
		} else if p.expect(token.TagGreaterThanEqualTo) {
			equalOp = token.TagGreaterThanEqualTo
		}
		right := p.parseExpression()
		return newCondNode(equalOp, left, right)
	}
}

func (p *Parser) parseExpression() *Node {
	op := int(token.TagPlus)
	p.accept(token.TagPlus)
	if p.accept(token.TagMinus) {
		op = token.TagMinus
	}
	term := newMathNode(op, newTerminalNode(&token.Token{Tag: token.TagInteger, Val: 0}), p.parseTerm())
	for {
		if p.accept(token.TagPlus) {
			op = token.TagPlus
		} else if p.accept(token.TagMinus) {
			op = token.TagPlus
		} else {
			break
		}
		second := p.parseTerm()
		term = newMathNode(op, term, second)
	}
	return term
}

func (p *Parser) parseTerm() *Node {
	op := int(token.TagTimes)
	fact := newMathNode(op, newTerminalNode(&token.Token{Tag: token.TagInteger, Val: 1}), p.parseFactor())
	for {
		if p.accept(token.TagTimes) {
			op = token.TagTimes
		} else if p.accept(token.TagDivide) {
			op = token.TagDivide
		} else {
			break
		}
		second := p.parseFactor()
		fact = newMathNode(op, fact, second)
	}
	return fact
}

func (p *Parser) parseFactor() *Node {
	iden := p.getTerminalNodeFromLookahead()
	if p.accept(token.TagIdentifier) || p.accept(token.TagInteger) {
		return iden
	} else {
		p.expect(token.TagLeftParen)
		expr := p.parseExpression()
		p.expect(token.TagRightParen)
		return expr
	}
}

// Returns an AST node containing the lookahead token if it is of type integer
// or identifier.
func (p *Parser) getTerminalNodeFromLookahead() *Node {
	// Only return a node if the lookahead token is an actual terminal.
	if p.look.Tag == token.TagIdentifier || p.look.Tag == token.TagInteger {
		return newTerminalNode(p.look)
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
