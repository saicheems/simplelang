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
	{".", true},
	{"CONST a = 3;.", true},
	{"CONST a = 3, b = 4;.", true},
	{"CONSt a = 3;.", false},
	{"VAR a;.", true},
	{"CONST a = 3; VAR b, c, d;.", true},
}

func TestScan(t *testing.T) {
	for _, pair := range tests {
		l, s := lexer.NewFromString(pair.test)
		p := New(l, s)
		pass := p.Parse()

		if pass != pair.expect {
			t.Error(
				"\nFor\n------\n"+pair.test,
				"\n------\nExpected\n------\n", pair.expect,
				"\n------\nGot\n------\n", pass,
			)
		}
	}
}
