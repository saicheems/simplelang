package parser

import (
	"testing"

	"github.com/saicheems/lexer"
)

type testPair struct {
	test   string
	expect bool
}

var tests = []testPair{
	{"", false},
	{"{}", true},
	{"{\n\n", false},
	{"{\n\n}", true},
	{"{\n\n} ", true},
}

func TestScan(t *testing.T) {
	for _, pair := range tests {
		l, s := lexer.NewLexerFromString(pair.test)
		p := NewParser(l, s)
		pass, err := p.Parse()

		if pass != pair.expect {
			t.Error(
				"For", pair.test,
				"expected", pair.expect,
				"got", pass, err,
			)
		}
	}
}
