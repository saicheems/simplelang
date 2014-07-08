// Package token implements the token type. It also includes a symbol table to
// track identifiers through the compilation stages.
package token

import (
	"errors"
	"io"
)

const (
	TagPeriod             = '.'
	TagComma              = ','
	TagSemicolon          = ';'
	TagEquals             = '='
	TagNotEquals          = '#'
	TagLessThan           = '<'
	TagGreaterThan        = '>'
	TagTimes              = '*'
	TagDivide             = '/'
	TagQuestion           = '?'
	TagExclamation        = '!'
	TagPlus               = '+'
	TagMinus              = '-'
	TagLeftCurlyBrace     = '{'
	TagRightCurlyBrace    = '}'
	TagLeftParen          = '('
	TagRightParen         = ')'
	TagInteger            = 256
	TagConst              = 257
	TagVar                = 258
	TagIdentifier         = 259
	TagProcedure          = 260
	TagAssignment         = 261
	TagCall               = 262
	TagBegin              = 263
	TagEnd                = 264
	TagIf                 = 265
	TagThen               = 266
	TagWhile              = 267
	TagDo                 = 268
	TagOdd                = 269
	TagLessThanEqualTo    = 270
	TagGreaterThanEqualTo = 271
)

var EOF = Token{Err: io.EOF}
var UnexpectedChar = Token{Err: errors.New("Unexpected character")}

// Token type; contains all the information necessary to represent lexical
// elements.
type Token struct {
	Tag int    // Tag.
	Val int    // Value.
	Ln  int    // Line number.
	Lex string // Lexeme.
	Err error  // Error.
}
