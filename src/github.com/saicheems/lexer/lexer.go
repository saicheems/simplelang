// Package lexer implements a lexical analyzer for sailang.
package lexer

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/saicheems/token"
)

// Unexpected character lexing error.
var UnexpectedChar = errors.New("Unexpected character")

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
func NewLexerFromString(s string) (*Lexer, *token.SymbolTable) {
	l := new(Lexer)
	l.rd = bufio.NewReader(strings.NewReader(s))
	l.sym = new(token.SymbolTable)
	l.loadKeywords()
	return l, l.sym
}

// Scan returns the next valid token from the input stream. If a lexing error
// occurs, it returns a Token of type Error. If the input stream is completed
// then an io.EOF error is returned. Otherwise error is nil.
func (l *Lexer) Scan() (*token.Token, error) {
	tok := new(token.Token)

	err := l.readCharAndWhitespace()
	if err != nil {
		return tok, err
	}
	err = l.scanComments()
	if err != nil {
		return tok, err
	}

	if l.peek == '+' {
		tok.Tag = token.TagPlus
		return tok, nil
	} else if l.peek == '-' {
		tok.Tag = token.TagMinus
		return tok, nil
	} else if l.peek == '{' {
		tok.Tag = token.TagLeftCurlyBrace
		return tok, nil
	} else if l.peek == '}' {
		tok.Tag = token.TagRightCurlyBrace
		return tok, nil
	}

	if isDigit(l.peek) {
		v := 0
		for {
			v = 10*v + convertCharDigitToInt(l.peek)
			err = l.readChar()
			if err != nil {
				break
			}
			if !isDigit(l.peek) {
				l.unreadChar()
				break
			}
		}
		tok.Tag = token.TagInteger
		tok.Val = v
		return tok, err
	}
	return tok, UnexpectedChar
}

func (l *Lexer) scanComments() error {
	if l.peek == '/' {
		match, err := l.readCharAndMatch('*')
		if err != nil {
			return err
		}
		if match {
			for {
				match, err := l.readCharAndMatch('*')
				if err != nil {
					return err
				}
				if match {
					match, err := l.readCharAndMatch('/')
					if err != nil {
						return err
					}
					if match {
						// Skip ahead to the next
						// non-whitespace peek char.
						err := l.readCharAndWhitespace()
						if err != nil {
							return err
						}
						break
					} else {
						l.unreadChar()
					}
				}
			}
		} else {
			l.unreadChar()
			match, err := l.readCharAndMatch('/')
			if err != nil {
				return err
			}
			if match {
				for {
					err := l.readChar()
					if err != nil {
						return err
					}
					if l.peek == '\n' {
						err := l.readCharAndWhitespace()
						if err != nil {
							return err
						}
						break
					}
				}
			}
		}
	}
	return nil
}

// Loads reserved keywords into the symbol table. Should be called on init.
func (l *Lexer) loadKeywords() {

}

func (l *Lexer) readChar() error {
	c, err := l.rd.ReadByte()
	if err != nil {
		return err
	}
	l.peek = c
	return nil
}

func (l *Lexer) readCharAndWhitespace() error {
	// If we see whitespace, let's go ahead and eat it here.
	// TODO: Make sure this doesn't cause any problems.
	for {
		c, err := l.rd.ReadByte()
		if err != nil {
			return err
		}
		if c == '\n' {
			// Increment the lexer's newline count if we see them.
			l.ln++
		} else if c == ' ' || c == '\t' {
			continue
		} else {
			l.peek = c
			break
		}
	}
	return nil
}

// Calls readChar and matches the input character to the peek character. If
// they match, the function returns true. Otherwise it returns false. The
// function returns false if there's an error along with the error.
// The only error possible should be io.EOF.
func (l *Lexer) readCharAndMatch(c byte) (bool, error) {
	err := l.readChar()
	if err != nil {
		return false, err
	}
	if l.peek != c {
		return false, nil
	}
	l.peek = ' '
	return true, nil
}

func (l *Lexer) unreadChar() error {
	// Error should never be encountered.
	return l.rd.UnreadByte()
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func convertCharDigitToInt(c byte) int {
	// TODO: Do any checks here?
	return int(c - '0')
}
