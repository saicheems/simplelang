// Package lexer implements a lexical analyzer for sailang.
package lexer

import (
	"bufio"
	"os"
	"strings"

	"github.com/saicheems/token"
)

type Lexer struct {
	rd   *bufio.Reader
	sym  *token.SymbolTable
	peek byte // Peek byte.
	ln   int  // Current line number in input stream.
}

func NewLexer(f *os.File, s *token.SymbolTable) *Lexer {
	l := new(Lexer)
	l.rd = bufio.NewReader(f)
	l.sym = s
	l.loadKeywords()
	return l
}

// Used to create a new lexer for testing.
func NewLexerFromReader(r *strings.Reader) *Lexer {
	l := new(Lexer)
	l.rd = bufio.NewReader(r)
	l.sym = new(token.SymbolTable)
	l.loadKeywords()
	return l
}

// Scan returns the next valid token from the input stream. If a lexing error
// occurs, it returns a Token of type Error. If the input stream is completed
// then an io.EOF error is returned. Otherwise error is nil.
func (l *Lexer) Scan() (*token.Token, error) {
	return new(token.Token), nil
}

// Loads reserved keywords into the symbol table. Should be called on init.
func (l *Lexer) loadKeywords() {

}