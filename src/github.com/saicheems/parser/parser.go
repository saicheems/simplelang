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
	top  *token.SymbolTable
	look *token.Token
	err  []error // Set error if we have a parse failure.
}

func New(l *lexer.Lexer, s *token.SymbolTable) *Parser {
	p := new(Parser)
	p.lex = l
	p.top = s
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
	p.match(token.TagPeriod)
	if len(p.err) != 0 {
		return nil
	}
	return newProgramNode(block)
}

func (p *Parser) parseBlock() *Node {
	cons := p.parseConst()
	vars := p.parseVar()
	proc := p.parseProcedure()
	stmt := p.parseStatement()
	if stmt == nil {
		p.err = append(p.err, fmt.Errorf("Syntax error near line %d.\n", p.look.Ln))
	}
	return newBlockNode(cons, vars, proc, stmt)
}

// Parses consts and returns an AST node. Returns nil if there are no consts.
func (p *Parser) parseConst() *Node {
	if !p.accept(token.TagConst) {
		return nil
	}
	cons := newConstNode()
	for {
		p.match(token.TagIdentifier)
		iden := p.getTerminalNodeFromLookahead()
		p.match(token.TagEquals)
		p.match(token.TagInteger)
		inte := p.getTerminalNodeFromLookahead()
		cons.appendNode(newAssignmentNode(iden, inte))
		if !p.accept(token.TagComma) {
			break
		}
	}
	p.match(token.TagSemicolon)
	return cons
}

// Parses vars and returns an AST node. Returns nil if there are no vars.
func (p *Parser) parseVar() *Node {
	if !p.accept(token.TagVar) {
		return nil
	}
	vars := newVarNode()
	for {
		iden := p.getTerminalNodeFromLookahead()
		p.match(token.TagIdentifier)
		vars.appendNode(iden)
		if !p.accept(token.TagComma) {
			break
		}
	}
	p.match(token.TagSemicolon)
	return vars
}

// Parses procedures and returns an AST node. Returns nil if there are no procedures.
func (p *Parser) parseProcedure() *Node {
	if !p.accept(token.TagProcedure) {
		return nil
	}
	proc := newProcedureParentNode()
	for {
		iden := p.getTerminalNodeFromLookahead()
		p.match(token.TagIdentifier)
		p.match(token.TagSemicolon)
		bloc := p.parseBlock()
		p.match(token.TagSemicolon)
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
		p.match(token.TagAssignment)
		expr := p.parseExpression()
		return newAssignmentNode(iden, expr)
	} else if p.accept(token.TagCall) {
		iden := p.getTerminalNodeFromLookahead()
		p.match(token.TagIdentifier)
		return newCallNode(iden)
	} else if p.accept(token.TagBegin) {
		begin := newBeginNode()
		stmt := p.parseStatement()
		if stmt == nil {
			p.err = append(p.err, fmt.Errorf("Syntax error near line %d.\n", p.look.Ln))
		}
		begin.appendNode(stmt)
		p.match(token.TagSemicolon)
		for {
			stmt := p.parseStatement()
			if stmt == nil {
				break
			}
			begin.appendNode(stmt)
			p.match(token.TagSemicolon)
			break
		}
		p.match(token.TagEnd)
		return begin
	} else if p.accept(token.TagIf) {
		cond := p.parseCondition()
		p.match(token.TagThen)
		stmt := p.parseStatement()
		if stmt == nil {
			p.err = append(p.err, fmt.Errorf("Syntax error near line %d.\n", p.look.Ln))
		}
		return newIfThenNode(cond, stmt)
	} else if p.accept(token.TagWhile) {
		cond := p.parseCondition()
		p.match(token.TagDo)
		stmt := p.parseStatement()
		if stmt == nil {
			p.err = append(p.err, fmt.Errorf("Syntax error near line %d.\n", p.look.Ln))
		}
		return newWhileDoNode(cond, stmt)
	}
	return nil
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
		} else if p.match(token.TagGreaterThanEqualTo) {
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
		p.match(token.TagLeftParen)
		expr := p.parseExpression()
		p.match(token.TagRightParen)
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

// Takes a tag and checks the token stream to see if it matches. Returns true
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
func (p *Parser) match(t int) bool {
	acc := p.accept(t)
	if !acc {
		p.err = append(p.err, fmt.Errorf("Syntax error near line %d.\n", p.look.Ln))
	}
	return acc
}
