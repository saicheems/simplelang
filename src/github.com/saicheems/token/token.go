// Package token implements the token type. It also includes a symbol table to
// track identifiers through the compilation stages.
package token

const (
	TagPlus    = '+' // Tag for plus symbol
	TagMinus   = '-'
	TagInteger = 256
)

// Token type; contains all the information necessary to represent lexical
// elements.
type Token struct {
	Tag int    // Tag.
	Val int    // Value.
	Ln  int    // Line number.
	Lex string // Lexeme.
	Err string // Error string.
}
