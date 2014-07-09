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
	{"BEGIN x := 3; END.", true},
	{"BEGIN END.", false},
	{"VAR x, y;.", false},
	{"VAR x, squ; BEGIN x := 3; END.", true},
	{"VAR x, squ; PROCEDURE square; BEGIN x := 3; END; BEGIN x := 3; END.", true},
	{"VAR x, squ; PROCEDURE square; BEGIN x := -3+8; END; BEGIN x := 3; END.", true},
	{"VAR x, squ;\n" +
		"PROCEDURE square;\n" +
		"BEGIN\n" +
		"squ:= x * x;\n" +
		"END;\n" +
		"BEGIN\n" +
		"x := 1;\n" +
		"WHILE x <= 10 DO\n" +
		"\tBEGIN\n" +
		"\tCALL square;\n" +
		"\tx := x + 1;\n" +
		"\tEND;\n" +
		"END.\n", true},
	{"CONST m = 3, n = 6;\n" +
		"VAR x, squ;\n" +
		"PROCEDURE square;\n" +
		"BEGIN\n" +
		"squ:= x * x;\n" +
		"END;\n" +
		"BEGIN\n" +
		"x := 1;\n" +
		"WHILE x <= 10 DO\n" +
		"\tBEGIN\n" +
		"\tCALL square;\n" +
		"\tx := x + 1;\n" +
		"\tEND;\n" +
		"END.\n", true},
}

func TestScan(t *testing.T) {
	for _, pair := range tests {
		l, s := lexer.NewFromString(pair.test)
		p := New(l, s)
		pass := p.Parse()

		if (pass != nil) != pair.expect {
			t.Error(
				"\nFor\n------\n"+pair.test,
				"\n------\nExpected\n------\n", pair.expect,
				"\n------\nGot\n------\n", pass,
			)
		}
	}
}
