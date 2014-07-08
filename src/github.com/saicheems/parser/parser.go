// Package parser implements the parsing stage of the compilation.
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
}

func New(l *lexer.Lexer, s *token.SymbolTable) *Parser {
	p := new(Parser)
	p.lex = l
	p.top = s
	return p
}

func (p *Parser) Parse() bool {
	p.move()
	fmt.Println(p.look)
	if !p.parseBlock() {
		return false
	}
	if !p.match('.') {
		return false
	}
	return true
}

func (p *Parser) parseBlock() bool {
	if !p.parseConsts() {
		return false
	}
	if !p.parseVars() {
		return false
	}
	if !p.parseProcedure() {
		return false
	}
	return true
}

func (p *Parser) parseConsts() bool {
	if p.match(token.TagConst) {
		if !p.match(token.TagIdentifier) {
			return false
		}
		if !p.match(token.TagEquals) {
			return false
		}
		if !p.match(token.TagInteger) {
			return false
		}
		for {
			if !p.match(',') {
				break
			}
			if !p.match(token.TagIdentifier) {
				return false
			}
			if !p.match(token.TagEquals) {
				return false
			}
			if !p.match(token.TagInteger) {
				return false
			}
		}
		if !p.match(token.TagSemicolon) {
			return false
		}
	}
	return true
}

func (p *Parser) parseVars() bool {
	if p.match(token.TagVar) {
		if !p.match(token.TagIdentifier) {
			return false
		}
		for {
			if !p.match(',') {
				break
			}
			if !p.match(token.TagIdentifier) {
				return false
			}
		}
		if !p.match(';') {
			return false
		}
	}
	return true
}

func (p *Parser) parseProcedure() bool {
	// TODO: Implement.
	return true
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
