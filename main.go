// Package main links all stages of the compilation together given an input file to compiled and
// produces the output language file.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/saicheems/simplelang/analyser"
	"github.com/saicheems/simplelang/codegen"
	"github.com/saicheems/simplelang/lexer"
	"github.com/saicheems/simplelang/parser"
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
	code := c.String()
	f, err = os.Create("out.s")
	if err != nil {
		fmt.Println("Error creating output file.")
	}
	io.WriteString(f, code)
	f.Close()
}
