package analyser

import (
	"testing"

	"github.com/saicheems/lexer"
	"github.com/saicheems/parser"
)

type testPair struct {
	test   string
	expect bool
}

var tests = []testPair{
	/*	{"BEGIN x:=3; END.", false},
		{"VAR x;BEGIN x:=3;END.", true},
		{"PROCEDURE hello;BEGIN x:= 3;END;BEGIN x := 3;END.", false},
		{"VAR x;PROCEDURE hello;BEGIN x:= 3;END;BEGIN x:=3;END.", true},
		{"CONST x:=3;PROCEDURE hello;BEGIN x:= 3;END;BEGIN x:=3;END.", false},
		{"VAR x;PROCEDURE hello;BEGIN x:= 3;END;BEGIN hello:=3;END.", false},*/
	{"CONST x=3,y=4;\n" +
		"VAR a,b,c;\n" +
		"PROCEDURE sum;\n" +
		"\tVAR a,b;\n" +
		"\tBEGIN\n" +
		"\t\ta:=x;\n" +
		"\t\tb:=y;\n" +
		"\t\tc:=a+b;\n" +
		"\tEND;\n" +
		"CALL sum.\n", true},
}

func TestAnalyse(t *testing.T) {
	for _, pair := range tests {
		l := lexer.NewFromString(pair.test)
		p := parser.New(l)
		a := New(p)
		pass := a.Analyse()

		if pass != pair.expect {
			t.Error(
				"\nFor\n------\n"+pair.test,
				"\n------\nExpected\n------\n", pair.expect,
				"\n------\nGot\n------\n", pass,
			)
		}
	}
}
