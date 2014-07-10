// Package lexer implements a lexical analyzer for sailang.
package lexer

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	"github.com/saicheems/token"
)

type Lexer struct {
	rd   *bufio.Reader
	res  map[string]int // Map of reserved keywords.
	peek byte           // Peek byte.
	ln   int            // Current line number in input stream.
}

func New(f *os.File) *Lexer {
	l := new(Lexer)
	l.rd = bufio.NewReader(f)
	l.res = make(map[string]int)
	l.loadKeywords()
	return l
}

// Used to create a new lexer for testing.
func NewFromString(s string) *Lexer {
	l := new(Lexer)
	l.rd = bufio.NewReader(strings.NewReader(s))
	l.res = make(map[string]int)
	l.loadKeywords()
	return l
}

// Scan returns the next valid token from the input stream. If a lexing error
// occurs, it returns a Token of type Error. If the input stream is completed
// then an io.EOF error is returned. Otherwise error is nil.
func (l *Lexer) Scan() *token.Token {
	if l.readCharAndWhitespace() != nil {
		return &token.EOF
	}
	if l.scanComments() != nil {
		return &token.EOF
	}
	tok := &token.Token{Ln: l.ln}
	if l.peek == '.' {
		tok.Tag = token.TagPeriod
		return tok
	} else if l.peek == ',' {
		tok.Tag = token.TagComma
		return tok
	} else if l.peek == ';' {
		tok.Tag = token.TagSemicolon
		return tok
	} else if l.peek == '=' {
		tok.Tag = token.TagEquals
		return tok
	} else if l.peek == '#' {
		tok.Tag = token.TagNotEquals
		return tok
	} else if l.peek == '<' {
		tok.Tag = token.TagLessThan
		// We won't do anything about an error here.
		m, _ := l.readCharAndMatch('=')
		if m {
			tok.Tag = token.TagLessThanEqualTo
			return tok
		} else {
			l.unreadChar()
		}
		return tok
	} else if l.peek == '>' {
		tok.Tag = token.TagGreaterThan
		// We won't do anything about an error here.
		m, _ := l.readCharAndMatch('=')
		if m {
			tok.Tag = token.TagGreaterThanEqualTo
			return tok
		} else {
			l.unreadChar()
		}
		return tok
	} else if l.peek == '*' {
		tok.Tag = token.TagTimes
		return tok
	} else if l.peek == '/' {
		tok.Tag = token.TagDivide
		return tok
	} else if l.peek == '?' {
		tok.Tag = token.TagQuestion
		return tok
	} else if l.peek == '!' {
		tok.Tag = token.TagExclamation
		return tok
	} else if l.peek == '+' {
		tok.Tag = token.TagPlus
		return tok
	} else if l.peek == '-' {
		tok.Tag = token.TagMinus
		return tok
	} else if l.peek == '{' {
		tok.Tag = token.TagLeftCurlyBrace
		return tok
	} else if l.peek == '}' {
		tok.Tag = token.TagRightCurlyBrace
		return tok
	} else if l.peek == '(' {
		tok.Tag = token.TagLeftParen
		return tok
	} else if l.peek == ')' {
		tok.Tag = token.TagRightParen
		return tok
	} else if l.peek == ':' {
		// We won't do anything about an error here.
		m, _ := l.readCharAndMatch('=')
		if m {
			tok.Tag = token.TagAssignment
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
		tok.Tag = token.TagIdentifier
		if l.res[lexeme] != 0 {
			tok.Tag = l.res[lexeme]
		}
		// We won't set the lexeme of the token if it's a keyword.
		if tok.Tag == token.TagIdentifier {
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
		tok.Tag = token.TagInteger
		tok.Val = v
		return tok
	}
	return &token.UnexpectedChar
}

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
			}
		}
	}
	return nil
}

// Loads reserved keywords into the symbol table. Should be called on init.
func (l *Lexer) loadKeywords() {
	l.res["CONST"] = token.TagConst
	l.res["VAR"] = token.TagVar
	l.res["PROCEDURE"] = token.TagProcedure
	l.res["CALL"] = token.TagCall
	l.res["BEGIN"] = token.TagBegin
	l.res["END"] = token.TagEnd
	l.res["IF"] = token.TagIf
	l.res["THEN"] = token.TagThen
	l.res["WHILE"] = token.TagWhile
	l.res["DO"] = token.TagDo
	l.res["ODD"] = token.TagOdd
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

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func convertCharDigitToInt(c byte) int {
	// TODO: Do any checks here?
	return int(c - '0')
}
