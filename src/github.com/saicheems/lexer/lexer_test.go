package lexer

import (
	"reflect"
	"testing"

	"github.com/saicheems/token"
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

	{".", token.Token{Tag: token.TagPeriod}},
	{",", token.Token{Tag: token.TagComma}},
	{";", token.Token{Tag: token.TagSemicolon}},
	{"=", token.Token{Tag: token.TagEquals}},
	{"#", token.Token{Tag: token.TagNotEquals}},
	{"<", token.Token{Tag: token.TagLessThan}},
	{">", token.Token{Tag: token.TagGreaterThan}},
	{"*", token.Token{Tag: token.TagTimes}},

	{"/", token.Token{Tag: token.TagDivide}},
	{"//**/", token.EOF},
	{"/**//", token.Token{Tag: token.TagDivide}},

	{"?", token.Token{Tag: token.TagQuestion}},
	{"!", token.Token{Tag: token.TagExclamation}},
	{"+", token.Token{Tag: token.TagPlus}},
	{"-", token.Token{Tag: token.TagMinus}},
	{"{", token.Token{Tag: token.TagLeftCurlyBrace}},
	{"}", token.Token{Tag: token.TagRightCurlyBrace}},
	{"(", token.Token{Tag: token.TagLeftParen}},
	{")", token.Token{Tag: token.TagRightParen}},
	{":=", token.Token{Tag: token.TagAssignment}},
	{"::=", token.UnexpectedChar},
	{":", token.UnexpectedChar},
	{"://asdf", token.UnexpectedChar},
	{"<=", token.Token{Tag: token.TagLessThanEqualTo}},
	{">=", token.Token{Tag: token.TagGreaterThanEqualTo}},

	{"Ident", token.Token{Tag: token.TagIdentifier, Lex: "Ident"}},
	{"Ident0123", token.Token{Tag: token.TagIdentifier, Lex: "Ident0123"}},
	{"0Ident0123", token.Token{Tag: token.TagInteger}},
	{"PROCEDURE", token.Token{Tag: token.TagProcedure}},
	{"CALL", token.Token{Tag: token.TagCall}},
	{"BEGIN", token.Token{Tag: token.TagBegin}},
	{"END", token.Token{Tag: token.TagEnd}},
	{"IF", token.Token{Tag: token.TagIf}},
	{"THEN", token.Token{Tag: token.TagThen}},
	{"WHILE", token.Token{Tag: token.TagWhile}},
	{"DO", token.Token{Tag: token.TagDo}},
	{"ODD", token.Token{Tag: token.TagOdd}},
}

var multiTokenTests = []multiTokenTestPair{
	{"+-", []token.Token{token.Token{Tag: token.TagPlus}, token.Token{Tag: token.TagMinus}, token.EOF}},
	{"BEGIN\n" +
		"END.", []token.Token{token.Token{Tag: token.TagBegin}, token.Token{Tag: token.TagEnd}, token.Token{Tag: token.TagPeriod}, token.EOF}},
}

func TestScan(t *testing.T) {
	for _, pair := range singleTokenTests {
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
	for _, pair := range multiTokenTests {
		l, _ := NewFromString(pair.test)
		out := []token.Token{}
		for {
			tok := l.Scan()
			out = append(out, *tok)
			if tok == &token.EOF {
				break
			}
		}
		if !reflect.DeepEqual(out, pair.expect) {
			t.Error(
				"For", pair.test,
				"expected", pair.expect,
				"got", out,
			)
		}
	}
}
