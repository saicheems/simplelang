// Package parser implements the parsing stage of the compilation including the
// parser and the abstract syntax tree.
package parser

import (
	"github.com/saicheems/lexer"
	"github.com/saicheems/token"
)

type Parser struct {
	lex  *lexer.Lexer
	top  *token.SymbolTable
	look *token.Token
}

func New(l *lexer.Lexer, s *token.SymbolTable) *Parser {
	p := new(Parser)
	p.lex = l
	p.top = s
	return p
}

// Parse begins reading the token stream recursively and generates an abstract
// syntax tree (AST). It returns a pointer to the root node of an abstract
// syntax tree. If there is an error in the parse, the return value will be nil.
func (p *Parser) Parse() *AstNode {
	p.move()
	program := newAstNodeProgram()
	block := p.parseBlock()
	if block == nil {
		return nil
	}
	if !p.match(token.TagPeriod) {
		return nil
	}
	return program
}

// Parses a block and returns an AST node.
func (p *Parser) parseBlock() *AstNode {
	a := newAstNode(TypeBlock)
	consts := p.parseConsts()
	if consts == nil {
		return nil
	}
	if !p.parseVars() {
		return nil
	}
	if !p.parseProcedure() {
		return nil
	}
	if !p.parseStatement() {
		return nil
	}
	return a
}

// Parses consts and returns an AST node.
func (p *Parser) parseConsts() *AstNode {
	v := make([]*AstNode, 0)
	if p.match(token.TagConst) {
		for {
			asgn := p.parseAssignment()
			if asgn == nil {
				return nil
			}
			v = append(v, asgn)
			if !p.match(token.TagComma) {
				break
			}
		}
		if !p.match(token.TagSemicolon) {
			return nil
		}
	}
	a := newAstNodeConst(v)
	return a
}

func (p *Parser) parseVars() bool {
	if p.match(token.TagVar) {
		if !p.match(token.TagIdentifier) {
			return false
		}
		for {
			if !p.match(token.TagComma) {
				break
			}
			if !p.match(token.TagIdentifier) {
				return false
			}
		}
		if !p.match(token.TagSemicolon) {
			return false
		}
	}
	return true
}

func (p *Parser) parseProcedure() bool {
	for {
		if !p.match(token.TagProcedure) {
			break
		}
		if !p.match(token.TagIdentifier) {
			return false
		}
		if !p.match(token.TagSemicolon) {
			return false
		}
		if p.parseBlock() == nil {
			return false
		}
		if !p.match(token.TagSemicolon) {
			return false
		}
	}
	return true
}

func (p *Parser) parseStatement() bool {
	if p.match(token.TagIdentifier) {
		if !p.match(token.TagAssignment) {
			return false
		}
		if !p.parseExpression() {
			return false
		}
	} else if p.match(token.TagCall) {
		if !p.match(token.TagIdentifier) {
			return false
		}
	} else if p.match(token.TagBegin) {
		if !p.parseStatement() {
			return false
		}
		if !p.match(token.TagSemicolon) {
			return false
		}
		for {
			if !p.parseStatement() {
				break
			}
			if !p.match(token.TagSemicolon) {
				return false
			}
		}
		if !p.match(token.TagEnd) {
			return false
		}
	} else if p.match(token.TagIf) {
		if !p.parseCondition() {
			return false
		}
		if !p.match(token.TagThen) {
			return false
		}
		if !p.parseStatement() {
			return false
		}
	} else {
		if !p.match(token.TagWhile) {
			// Expected statement.
			return false
		}
		if !p.parseCondition() {
			return false
		}
		if !p.match(token.TagDo) {
			return false
		}
		if !p.parseStatement() {
			return false
		}
	}
	return true
}

func (p *Parser) parseCondition() bool {
	if p.match(token.TagOdd) {
		if !p.parseExpression() {
			return false
		}
	} else {
		if !p.parseExpression() {
			// Expected condition.
			return false
		}
		if !(p.match(token.TagEquals) || p.match(token.TagNotEquals) ||
			p.match(token.TagLessThan) || p.match(token.TagLessThanEqualTo) ||
			p.match(token.TagGreaterThan) || p.match(token.TagGreaterThanEqualTo)) {
			return false
		}
		if !p.parseExpression() {
			return false
		}
	}
	return true
}

func (p *Parser) parseExpression() bool {
	p.match(token.TagPlus)
	p.match(token.TagMinus)

	if !p.parseTerm() {
		return false
	}
	for {
		if !(p.match(token.TagPlus) || p.match(token.TagMinus)) {
			break
		}
		if !p.parseTerm() {
			return false
		}
	}
	return true
}

func (p *Parser) parseTerm() bool {
	if p.parseFactor() {
		for {
			if !(p.match(token.TagTimes) || p.match(token.TagDivide)) {
				break
			}
			if !p.parseFactor() {
				return false
			}
		}
		return true
	}
	return false
}

func (p *Parser) parseFactor() bool {
	if p.match(token.TagIdentifier) {
	} else if p.match(token.TagInteger) {
	} else {
		if !p.match(token.TagLeftParen) {
			return false
		}
		if !p.parseExpression() {
			return false
		}
		if !p.match(token.TagRightParen) {
			return false
		}
	}
	return true
}

func (p *Parser) parseAssignment() *AstNode {
	a := newAstNodeTerminal(p.look)
	if !p.match(token.TagIdentifier) {
		return nil
	}
	if !p.match(token.TagEquals) {
		return nil
	}
	b := newAstNodeTerminal(p.look)
	if !p.match(token.TagInteger) {
		return nil
	}
	asgn := newAstNodeAssigment(a, b)
	return asgn
}

func (p *Parser) move() {
	tok := p.lex.Scan()
	p.look = tok
}

func (p *Parser) match(t int) bool {
	if p.look.Tag == t {
		p.move()
		return true
	}
	return false
}
