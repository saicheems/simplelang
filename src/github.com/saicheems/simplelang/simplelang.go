// Package main links all stages of the compilation together given an input file to compiled and
// produces the output language file.
package main

import (
	"fmt"
	"os"

	"github.com/saicheems/analyser"
	"github.com/saicheems/codegen"
	"github.com/saicheems/lexer"
	"github.com/saicheems/parser"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("One argument is required: a path to a file to be compiled.")
		return
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Error opening file.")
		return
	}
	l := lexer.New(f)
	p := parser.New(l)
	a := analyser.New(p)
	c := codegen.New(a)
	c.Generate()
}
