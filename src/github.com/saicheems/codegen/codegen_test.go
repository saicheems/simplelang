package codegen

import (
	"testing"

	"github.com/saicheems/analyser"
	"github.com/saicheems/lexer"
	"github.com/saicheems/parser"
)

type testPair struct {
	test   string
	expect string
}

var tests = []testPair{
	{"VAR x;BEGIN x:=3;END.", "main:\n"},
}

func TestAnalyse(t *testing.T) {
	for _, pair := range tests {
		l := lexer.NewFromString(pair.test)
		p := parser.New(l)
		a := analyser.New(p)
		c := NewToString(a)
		c.Generate()
		pass := c.ToString()

		if pass != pair.expect {
			t.Error(
				"\nFor\n------\n"+pair.test,
				"\n------\nExpected\n------\n", pair.expect,
				"\n------\nGot\n------\n", pass,
			)
		}
	}
}
