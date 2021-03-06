package parser

import (
	"testing"

	"github.com/saicheems/simplelang/lexer"
)

type testPair struct {
	test   string
	expect bool
}

var tests = []testPair{
	{"BEGIN x := 3; END.", true},
	{"BEGIN CALL x; END.", true},
	{"BEGIN x := 3; END", false},
	{"BEGIN ; END.", false},
	{"\n\n\n\nBEGIN ; END.", false},
	{"BEGIN END.", false},
	{"BEGIN WHILE x = 3 DO END.", false},
	{"BEGIN WHILE x = 3 DO BEGIN x := 3; END; END.", true},
	{"BEGIN hello := ^asdf; END.", false},
	{"BEGIN hello := asdf^; END.", false},
	{"BEGIN hello := ^asdf^; END.", false},
	{"BEGIN ^hello^ ^:=^ ^asdf^;^ ^END^.^", false},
	{"BEGIN\n\tWHILE x = 3 DO\n\t\tBEGIN\n\t\t\tx := 3;\n\t\tEND\nEND.", false},
	{"VAR x, y;.", false},
	{"VAR x, squ; BEGIN x := 3; END.", true},
	{"VAR x, squ; PROCEDURE square; BEGIN x := 3; END; BEGIN x := 3; END.", true},
	{"VAR x, squ; PROCEDURE square; BEGIN x := -3+8; END; BEGIN x := 3; END.", true},
	{"CONST a = 1, b = 2; VAR x, y, z; PROCEDURE proc; CONST a = 3, b = 4; VAR x, y, z;" +
		"CALL proc; CALL proc.",
		true},
	{"VAR x, y;" +
		"BEGIN\n" +
		"\tIF x + 3 = y - 5 THEN\n" +
		"\tx := 3;\n" +
		"\nEND.", true},
	{"VAR x, y;" +
		"BEGIN\n" +
		"\tIF x + 3 # y - 5 THEN\n" +
		"\tx := 3;\n" +
		"\nEND.", true},
	{"VAR x, y;" +
		"BEGIN\n" +
		"\tIF x + 3 > y - 5 THEN\n" +
		"\tx := 3;\n" +
		"\nEND.", true},
	{"VAR x, y;" +
		"BEGIN\n" +
		"\tIF x + 3 < y - 5 THEN\n" +
		"\tx := 3;\n" +
		"\nEND.", true},
	{"VAR x, y;" +
		"BEGIN\n" +
		"\tIF x + 3 >= y - 5 THEN\n" +
		"\tx := 3;\n" +
		"\nEND.", true},
	{"VAR x, y;" +
		"BEGIN\n" +
		"\tIF x + 3 <= y - 5 THEN\n" +
		"\tx := 3;\n" +
		"\nEND.", true},
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

func TestScan(t *testing.T) {
	for _, pair := range tests {
		l := lexer.NewFromString(pair.test)
		p := New(l)
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
