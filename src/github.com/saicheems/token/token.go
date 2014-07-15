// Package token implements the token type.
package token

import (
	"errors"
	"io"
)

const (
	Period             = iota // .
	Comma                     // ,
	Semicolon                 // ;
	Equals                    // =
	NotEquals                 // #
	LessThan                  // <
	GreaterThan               // >
	LessThanEqualTo           // <=
	GreaterThanEqualTo        // >=
	Plus                      // +
	Minus                     // -
	Times                     // *
	Divide                    // /
	LeftCurlyBrace            // {
	RightCurlyBrace           // }
	LeftParen                 // (
	RightParen                // )
	Exclamation               // !
	Assignment                // :=
	Integer                   // ex. 42
	Identifier                // ex. abc, abc123, ABC123
	Begin                     // BEGIN
	Call                      // CALL
	Const                     // CONST
	Do                        // DO
	End                       // END
	If                        // IF
	Odd                       // ODD
	Procedure                 // PROCEDURE
	Then                      // THEN
	Var                       // VAR
	While                     // WHILE
	Error                     // Special type for EOF and UnexpectedChar.
)

// EOF is a pointer to a Token with the Err field set to io.EOF. It is used to represent the end of
// a token stream.
var EOF = &Token{Tag: Error, Err: io.EOF}

// UnexpectedChar is a pointer to a Token with the Err field set to "Unexpected character". It is
// used to represent a input character that does not fit into any of the tags defined by the
// package.
var UnexpectedChar = &Token{Tag: Error, Err: errors.New("Unexpected character")}

// Token implements a lexical token. It contains all the information needed by the compiler to
// represent a lexical unit.
type Token struct {
	Tag int    // Tag. One of the constants defined in this package.
	Val int    // Value.
	Ln  int    // Line number.
	Lex string // Lexeme.
	Err error  // Error.
}

// New returns a new Token with the line number field set to the argument.
func New(ln int) *Token {
	return &Token{Ln: ln}
}
