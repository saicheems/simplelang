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

	{"//asdfsasdf", token.Token{}, io.EOF},
	{"//asdfsasdf\n", token.Token{}, io.EOF},
	{"//asdfsasdf+", token.Token{}, io.EOF},
	{"//asdf\n@", token.Token{}, UnexpectedChar},
	{"/**/", token.Token{}, io.EOF},
	{"/*asdf*/", token.Token{}, io.EOF},
	{"/*asdf\nasdfasdf\nasdf*/", token.Token{}, io.EOF},
	{"/**************ASDF**A******/", token.Token{}, io.EOF},
	{"/******\n\t\n\t*****\n\t\n**    *ASDF**A******/\n\t\n", token.Token{}, io.EOF},

	{"+", token.Token{Tag: token.TagPlus}, nil},
	{"-", token.Token{Tag: token.TagMinus}, nil},
	{"134", token.Token{Tag: token.TagInteger, Val: 134}, io.EOF},
	{"134 ", token.Token{Tag: token.TagInteger, Val: 134}, nil},
	{" 00001 ", token.Token{Tag: token.TagInteger, Val: 1}, nil},
}

func TestScan(t *testing.T) {
	for _, pair := range tests {
		l, _ := NewLexerFromString(pair.test)
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
