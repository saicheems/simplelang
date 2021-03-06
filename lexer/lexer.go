// Package lexer implements a lexical analyzer for sailang.
package lexer

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	"github.com/saicheems/simplelang/token"
)

// Lexer implements the lexical scanning phase of the compilation.
type Lexer struct {
	rd   *bufio.Reader
	res  map[string]int // Map of reserved keywords.
	peek byte           // Peek byte.
	ln   int            // Current line number in input stream.
}

// New returns a new Lexer given a File. The file is opened and a bufio.Reader is created to read
// input characters.
func New(f *os.File) *Lexer {
	l := new(Lexer)
	l.rd = bufio.NewReader(f)
	l.res = make(map[string]int)
	l.loadKeywords()
	return l
}

// NewFromString returns a new Lexer given a string.
func NewFromString(s string) *Lexer {
	l := new(Lexer)
	l.rd = bufio.NewReader(strings.NewReader(s))
	l.res = make(map[string]int)
	l.loadKeywords()
	return l
}

// Scan returns the next valid token from the input stream. If a lexing error occurs, it returns an
// Token with the tag Error. If the input stream is completed then token.EOF is returned. Otherwise
// token.UnexpectedChar is returned..
func (l *Lexer) Scan() *token.Token {
	if l.readCharAndWhitespace() != nil {
		return token.EOF
	}
	if l.scanComments() != nil {
		return token.EOF
	}
	tok := token.New(l.ln)
	if l.peek == '.' {
		tok.Tag = token.Period
		return tok
	} else if l.peek == ',' {
		tok.Tag = token.Comma
		return tok
	} else if l.peek == ';' {
		tok.Tag = token.Semicolon
		return tok
	} else if l.peek == '=' {
		tok.Tag = token.Equals
		return tok
	} else if l.peek == '#' {
		tok.Tag = token.NotEquals
		return tok
	} else if l.peek == '<' {
		tok.Tag = token.LessThan
		// We won't do anything about an error here.
		m, _ := l.readCharAndMatch('=')
		if m {
			tok.Tag = token.LessThanEqualTo
			return tok
		} else {
			l.unreadChar()
		}
		return tok
	} else if l.peek == '>' {
		tok.Tag = token.GreaterThan
		// We won't do anything about an error here.
		m, _ := l.readCharAndMatch('=')
		if m {
			tok.Tag = token.GreaterThanEqualTo
			return tok
		} else {
			l.unreadChar()
		}
		return tok
	} else if l.peek == '*' {
		tok.Tag = token.Times
		return tok
	} else if l.peek == '/' {
		tok.Tag = token.Divide
		return tok
	} else if l.peek == '+' {
		tok.Tag = token.Plus
		return tok
	} else if l.peek == '-' {
		tok.Tag = token.Minus
		return tok
	} else if l.peek == '{' {
		tok.Tag = token.LeftCurlyBrace
		return tok
	} else if l.peek == '}' {
		tok.Tag = token.RightCurlyBrace
		return tok
	} else if l.peek == '(' {
		tok.Tag = token.LeftParen
		return tok
	} else if l.peek == ')' {
		tok.Tag = token.RightParen
		return tok
	} else if l.peek == '!' {
		tok.Tag = token.Exclamation
		return tok
	} else if l.peek == ':' {
		// We won't do anything about an error here.
		m, _ := l.readCharAndMatch('=')
		if m {
			tok.Tag = token.Assignment
			return tok
		} else {
			l.unreadChar()
		}
	}
	if isAlpha(l.peek) {
		var strBuf bytes.Buffer
		for {
			strBuf.WriteByte(l.peek)
			err := l.readChar()
			if err != nil {
				break
			}
			if !(isAlpha(l.peek) || isDigit(l.peek)) {
				l.unreadChar()
				break
			}
		}
		lexeme := strBuf.String()
		tok.Tag = token.Identifier
		if l.res[lexeme] != 0 {
			tok.Tag = l.res[lexeme]
		}
		// We won't set the lexeme of the token if it's a keyword.
		if tok.Tag == token.Identifier {
			tok.Lex = lexeme
		}
		return tok
	}
	if isDigit(l.peek) {
		v := 0
		for {
			v = 10*v + convertCharDigitToInt(l.peek)
			err := l.readChar()
			if err != nil {
				break
			}
			if !isDigit(l.peek) {
				l.unreadChar()
				break
			}
		}
		tok.Tag = token.Integer
		tok.Val = v
		return tok
	}
	return token.UnexpectedChar
}

// scanComments checks for block comments or line comments and eats input until they are terminated.
// It returns an io.EOF error if EOF is encountered. Otherwise it returns nil. Otherwise it returns
// nil. Otherwise it returns nil. Otherwise it returns nil.
func (l *Lexer) scanComments() error {
	if l.peek == '/' {
		match, err := l.readCharAndMatch('*')
		if err != nil {
			// We'll return nil in this case so we can pick up the divide token...
			return nil
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
			} else {
				l.unreadChar()
				// We need to reset the state so division op can be read.
				l.peek = '/'
			}
		}
	}
	return nil
}

// loadKeywords loads reserved keywords into the reserved keyword table. Should be called on init.
func (l *Lexer) loadKeywords() {
	l.res["CONST"] = token.Const
	l.res["VAR"] = token.Var
	l.res["PROCEDURE"] = token.Procedure
	l.res["CALL"] = token.Call
	l.res["BEGIN"] = token.Begin
	l.res["END"] = token.End
	l.res["IF"] = token.If
	l.res["THEN"] = token.Then
	l.res["WHILE"] = token.While
	l.res["DO"] = token.Do
	l.res["ODD"] = token.Odd
}

// readChar reads a single character from the input stream and sets peek. It returns the error
// io.EOF if EOF is encountered. Otherwise it returns nil.
func (l *Lexer) readChar() error {
	c, err := l.rd.ReadByte()
	if err != nil {
		return err
	}
	l.peek = c
	return nil
}

// readChar disregards all whitespace before the first non-whitespace character in the input stream.
// It stops at the first non-whitespace character and and sets peek. It returns the error io.EOF if
// EOF is encountered. Otherwise it returns nil.
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

// readCharAndMatch calls readChar and matches the input character to the peek character. If they
// match, the function returns true. Otherwise it returns false. The function returns false if
// there's an error. The error returned will be either io.EOF or nil.
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

// unreadChar unreads the last character read from the input stream. It does not modify peek.
func (l *Lexer) unreadChar() error {
	// Error should never be encountered.
	return l.rd.UnreadByte()
}

// isAlpha returns true if the input byte is an ASCII alphabetic character (a-z, A-Z). Otherwise it
// returns false.
func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isDigit returns true if the input byte is an ASCII digit (0-9). Otherwise it returns false.
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// convertDigitToInt returns the integer version of the input byte if the input byte is a digit
// (0-9). Otherwise it returns -1.
func convertCharDigitToInt(c byte) int {
	if isDigit(c) {
		return int(c - '0')
	}
	return -1
}
