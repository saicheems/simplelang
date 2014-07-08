// Package parser implements the parsing stage of the compilation.
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

func (p *Parser) Parse() bool {
	p.move()
	if !p.parseBlock() {
		return false
	}
	if !p.match(token.TagPeriod) {
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
	if !p.parseStatement() {
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
			if !p.match(token.TagComma) {
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
		if !p.parseBlock() {
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
