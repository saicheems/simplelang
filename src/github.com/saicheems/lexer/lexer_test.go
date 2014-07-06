package lexer

import (
	"io"
	"testing"

	"github.com/saicheems/token"
)

type testPair struct {
	test   string
	expect token.Token
	err    error
}

var tests = []testPair{
	{"", token.Token{}, io.EOF},
	{" ", token.Token{}, io.EOF},
	{"\t", token.Token{}, io.EOF},
	{"\n", token.Token{}, io.EOF},
	{"\n\t\n\t", token.Token{}, io.EOF},
	{"    \n\t\n   ", token.Token{}, io.EOF},
}

func TestScan(t *testing.T) {
	for _, pair := range tests {
		l := NewLexerFromString(pair.test)
		tok, err := l.Scan()

		if err != pair.err || *tok != pair.expect {
			t.Error(
				"For", pair.test,
				"expected", pair.expect,
				"got", tok, err,
			)
		}
	}
}
