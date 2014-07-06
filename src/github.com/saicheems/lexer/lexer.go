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
func NewLexerFromString(s string) *Lexer {
	l := new(Lexer)
	l.rd = bufio.NewReader(strings.NewReader(s))
	l.sym = new(token.SymbolTable)
	l.loadKeywords()
	return l
}

// Scan returns the next valid token from the input stream. If a lexing error
// occurs, it returns a Token of type Error. If the input stream is completed
// then an io.EOF error is returned. Otherwise error is nil.
func (l *Lexer) Scan() (*token.Token, error) {
	tok := new(token.Token)

	err := l.scanWhitespace()
	if err != nil {
		return tok, err
	}
	return tok, nil
}

func (l *Lexer) scanWhitespace() error {
	for {
		err := l.readChar()
		if err != nil {
			// If we hit an EOF, finish.
			return err
		}
		if l.peek == ' ' || l.peek == '\t' {
			// Continue to eat any whitespace.
			continue
		} else if l.peek == '\n' {
			// If we encounter a newline, increment the line count.
			l.ln++
		} else {
			break
		}
	}
	return nil
}

// Loads reserved keywords into the symbol table. Should be called on init.
func (l *Lexer) loadKeywords() {

}

func (l *Lexer) readChar() error {
	c, err := l.rd.ReadByte()
	l.peek = c
	return err
}

func (l *Lexer) unreadChar() error {
	// Error should never be encountered.
	return l.rd.UnreadByte()
}
