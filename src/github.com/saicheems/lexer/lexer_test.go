package lexer

import (
	"testing"

	"github.com/saicheems/token"
)

type testPair struct {
	test   string
	expect token.Token
}

var tests = []testPair{
	{"", token.EOF},
	{" ", token.EOF},
	{"\t", token.EOF},
	{"\n", token.EOF},
	{"\n\t\n\t", token.EOF},
	{"    \n\t\n   ", token.EOF},

	{"//asdfsasdf", token.EOF},
	{"//asdfsasdf\n", token.EOF},
	{"//asdfsasdf+", token.EOF},
	{"//asdf\n@", token.UnexpectedChar},
	{"/**/", token.EOF},
	{"/*asdf*/", token.EOF},
	{"/*asdf\nasdfasdf\nasdf*/", token.EOF},
	{"/**************ASDF**A******/", token.EOF},
	{"/******\n\t\n\t*****\n\t\n**    *ASDF**A******/\n\t\n", token.EOF},

	{"+", token.Token{Tag: token.TagPlus}},
	{"-", token.Token{Tag: token.TagMinus}},
	{"134", token.Token{Tag: token.TagInteger, Val: 134}},
	{"134 ", token.Token{Tag: token.TagInteger, Val: 134}},
	{" 00001 ", token.Token{Tag: token.TagInteger, Val: 1}},
}

func TestScan(t *testing.T) {
	for _, pair := range tests {
		l, _ := NewFromString(pair.test)
		tok := l.Scan()

		if *tok != pair.expect {
			t.Error(
				"For", pair.test,
				"expected", pair.expect,
				"got", tok,
			)
		}
	}
}
