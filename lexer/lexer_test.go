package lexer

import (
	"reflect"
	"testing"

	"github.com/saicheems/simplelang/token"
)

type singleTokenTestPair struct {
	test   string
	expect token.Token
}

type multiTokenTestPair struct {
	test   string
	expect []token.Token
}

var singleTokenTests = []singleTokenTestPair{
	{"", *token.EOF},
	{" ", *token.EOF},
	{"\t", *token.EOF},
	{"\n", *token.EOF},
	{"\n\t\n\t", *token.EOF},
	{"    \n\t\n   ", *token.EOF},

	{"//asdfsasdf", *token.EOF},
	{"//asdfsasdf\n", *token.EOF},
	{"//asdfsasdf+", *token.EOF},
	{"//asdf\n@", *token.UnexpectedChar},
	{"/**/", *token.EOF},
	{"/*asdf*/", *token.EOF},
	{"/*asdf\nasdfasdf\nasdf*/", *token.EOF},
	{"/**************ASDF**A******/", *token.EOF},
	{"/******\n\t\n\t*****\n\t\n**    *ASDF**A******/\n\t\n", *token.EOF},

	{"+", token.Token{Tag: token.Plus}},
	{"-", token.Token{Tag: token.Minus}},
	{"134", token.Token{Tag: token.Integer, Val: 134}},
	{"134 ", token.Token{Tag: token.Integer, Val: 134}},
	{" 00001 ", token.Token{Tag: token.Integer, Val: 1}},

	{".", token.Token{Tag: token.Period}},
	{",", token.Token{Tag: token.Comma}},
	{";", token.Token{Tag: token.Semicolon}},
	{"=", token.Token{Tag: token.Equals}},
	{"#", token.Token{Tag: token.NotEquals}},
	{"<", token.Token{Tag: token.LessThan}},
	{">", token.Token{Tag: token.GreaterThan}},
	{"*", token.Token{Tag: token.Times}},

	{"/", token.Token{Tag: token.Divide}},
	{"//**/", *token.EOF},
	{"/**//", token.Token{Tag: token.Divide}},

	{"+", token.Token{Tag: token.Plus}},
	{"-", token.Token{Tag: token.Minus}},
	{"{", token.Token{Tag: token.LeftCurlyBrace}},
	{"}", token.Token{Tag: token.RightCurlyBrace}},
	{"(", token.Token{Tag: token.LeftParen}},
	{")", token.Token{Tag: token.RightParen}},
	{":=", token.Token{Tag: token.Assignment}},
	{"::=", *token.UnexpectedChar},
	{":", *token.UnexpectedChar},
	{"://asdf", *token.UnexpectedChar},
	{"<=", token.Token{Tag: token.LessThanEqualTo}},
	{">=", token.Token{Tag: token.GreaterThanEqualTo}},

	{"Ident", token.Token{Tag: token.Identifier, Lex: "Ident"}},
	{"Ident0123", token.Token{Tag: token.Identifier, Lex: "Ident0123"}},
	{"0Ident0123", token.Token{Tag: token.Integer}},
	{"PROCEDURE", token.Token{Tag: token.Procedure}},
	{"CALL", token.Token{Tag: token.Call}},
	{"BEGIN", token.Token{Tag: token.Begin}},
	{"END", token.Token{Tag: token.End}},
	{"IF", token.Token{Tag: token.If}},
	{"THEN", token.Token{Tag: token.Then}},
	{"WHILE", token.Token{Tag: token.While}},
	{"DO", token.Token{Tag: token.Do}},
	{"ODD", token.Token{Tag: token.Odd}},
}

var multiTokenTests = []multiTokenTestPair{
	{"+-", []token.Token{token.Token{Tag: token.Plus}, token.Token{Tag: token.Minus}, *token.EOF}},
	{"BEGIN\n" +
		"END.", []token.Token{token.Token{Tag: token.Begin}, token.Token{Tag: token.End, Ln: 1},
		token.Token{Tag: token.Period, Ln: 1}, *token.EOF}},
	{"BEGIN\n" +
		"CONST\n" +
		"\tm = 7;\n" +
		"PROCEDURE multiply;\n" +
		"VAR a;\n" +
		"BEGIN\n" +
		"\ta := x;\n" +
		"END;\n" +
		"END.", []token.Token{token.Token{Tag: token.Begin},
		token.Token{Tag: token.Const, Ln: 1}, token.Token{Tag: token.Identifier, Lex: "m", Ln: 2},
		token.Token{Tag: token.Equals, Ln: 2}, token.Token{Tag: token.Integer, Val: 7, Ln: 2},
		token.Token{Tag: token.Semicolon, Ln: 2}, token.Token{Tag: token.Procedure, Ln: 3},
		token.Token{Tag: token.Identifier, Lex: "multiply", Ln: 3}, token.Token{Tag: token.Semicolon, Ln: 3},
		token.Token{Tag: token.Var, Ln: 4}, token.Token{Tag: token.Identifier, Lex: "a", Ln: 4},
		token.Token{Tag: token.Semicolon, Ln: 4}, token.Token{Tag: token.Begin, Ln: 5},
		token.Token{Tag: token.Identifier, Lex: "a", Ln: 6}, token.Token{Tag: token.Assignment, Ln: 6},
		token.Token{Tag: token.Identifier, Lex: "x", Ln: 6}, token.Token{Tag: token.Semicolon, Ln: 6},
		token.Token{Tag: token.End, Ln: 7}, token.Token{Tag: token.Semicolon, Ln: 7},
		token.Token{Tag: token.End, Ln: 8}, token.Token{Tag: token.Period, Ln: 8}, *token.EOF}},
	{"CONST a = 3;.", []token.Token{token.Token{Tag: token.Const}, token.Token{Tag: token.Identifier, Lex: "a"},
		token.Token{Tag: token.Equals}, token.Token{Tag: token.Integer, Val: 3}, token.Token{Tag: token.Semicolon},
		token.Token{Tag: token.Period}, *token.EOF}},
	{"x := ^323;", []token.Token{token.Token{Tag: token.Identifier, Lex: "x"}, token.Token{Tag: token.Assignment},
		*token.UnexpectedChar, token.Token{Tag: token.Integer, Val: 323}, token.Token{Tag: token.Semicolon}, *token.EOF}},
	{"x := ^asdf;", []token.Token{token.Token{Tag: token.Identifier, Lex: "x"}, token.Token{Tag: token.Assignment},
		*token.UnexpectedChar, token.Token{Tag: token.Identifier, Lex: "asdf"}, token.Token{Tag: token.Semicolon}, *token.EOF}},
	{"x:=a+b;", []token.Token{token.Token{Tag: token.Identifier, Lex: "x"}, token.Token{Tag: token.Assignment},
		token.Token{Tag: token.Identifier, Lex: "a"}, token.Token{Tag: token.Plus},
		token.Token{Tag: token.Identifier, Lex: "b"}, token.Token{Tag: token.Semicolon}, *token.EOF}},
	{"x:=a/b;", []token.Token{token.Token{Tag: token.Identifier, Lex: "x"}, token.Token{Tag: token.Assignment},
		token.Token{Tag: token.Identifier, Lex: "a"}, token.Token{Tag: token.Divide},
		token.Token{Tag: token.Identifier, Lex: "b"}, token.Token{Tag: token.Semicolon}, *token.EOF}},
}

func TestScan(t *testing.T) {
	for _, pair := range singleTokenTests {
		l := NewFromString(pair.test)
		tok := l.Scan()

		if *tok != pair.expect {
			t.Error(
				"\nFor\n------\n", pair.test,
				"\n------\nExpected\n------\n", pair.expect,
				"\n------\nGot\n------\n", tok,
			)
		}
	}
	for _, pair := range multiTokenTests {
		l := NewFromString(pair.test)
		out := []token.Token{}
		for {
			tok := l.Scan()
			out = append(out, *tok)
			if tok == token.EOF {
				break
			}
		}
		if !reflect.DeepEqual(out, pair.expect) {
			t.Error(
				"\nFor\n------\n"+pair.test,
				"\n------\nExpected\n------\n", pair.expect,
				"\n------\nGot\n------\n", out,
			)
		}
	}
}
