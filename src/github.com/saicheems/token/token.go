// Package token implements the token type. It also includes a symbol table to
// track identifiers through the compilation stages.
package token

type Token struct {
	tag int    // Tag.
	val int    // Value.
	ln  int    // Line number.
	lex string // Lexeme.
	err string // Error string.
}
