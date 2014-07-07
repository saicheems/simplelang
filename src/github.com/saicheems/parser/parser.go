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

func NewParser(l *lexer.Lexer, s *token.SymbolTable) *Parser {
	p := new(Parser)
	p.lex = l
	p.top = s
	return p
}

func (p *Parser) Parse() (bool, error) {
	err := p.move()
	if err != nil {
		return false, err
	}
	return p.parseBlock()
}

func (p *Parser) parseBlock() (bool, error) {
	// TODO: What about all of these EOF errors?
	m, err := p.match('{')
	if !m {
		return m, fmt.Errorf("Syntax error near line %d.\n", p.look.Ln)
	}
	// Match statements.
	m, err = p.match('}')
	if !m {
		return m, fmt.Errorf("Syntax error near line %d.\n", p.look.Ln)
	}
	return true, err
}

func (p *Parser) move() error {
	tok, err := p.lex.Scan()
	if err != nil {
		return err
	}
	p.look = tok
	fmt.Println(tok)
	return nil
}

func (p *Parser) match(t int) (bool, error) {
	if p.look.Tag == t {
		err := p.move()
		return true, err
	} else {
		return false, nil
	}
}
