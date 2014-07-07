// Package token implements the token type. It also includes a symbol table to
// track identifiers through the compilation stages.
package token

import (
	"errors"
	"io"
)

const (
	TagPlus            = '+' // Tag for plus symbol
	TagMinus           = '-'
	TagLeftCurlyBrace  = '{'
	TagRightCurlyBrace = '}'
	TagInteger         = 256
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
