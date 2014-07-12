// Package analyser implements some basic semantic analysis on the parse tree.
package analyser

import (
	"fmt"

	"github.com/saicheems/ast"
	"github.com/saicheems/parser"
	"github.com/saicheems/symtable"
	"github.com/saicheems/token"
)

type Analyser struct {
	par *parser.Parser
	err []error
}

func New(par *parser.Parser) *Analyser {
	a := new(Analyser)
	a.par = par
	a.err = make([]error, 0)
	return a
}

func (a *Analyser) Analyse() bool {
	ast := a.par.Parse()
	if ast == nil {
		return false
	}
	a.loadSymbolTables(ast.Children[0])
	a.recurseProgramCheck(ast)
	if len(a.err) > 0 {
		for _, err := range a.err {
			fmt.Println(err)
		}
		return false
	}
	return true
}

// Loads all of the symbol tables. In this simple language all symbols should be defined in the
// header of the program, so it is an easy pass.
func (a *Analyser) loadSymbolTables(node *ast.Node) {
	sym := symtable.New()
	cons := node.Children[0] // Constants
	vars := node.Children[1] // Vars
	proc := node.Children[2] // Procedures

	for _, node := range cons.Children {
		iden := node.Children[0]
		sym.Put(symtable.Constant, iden.Tok.Lex)
	}
	for _, node := range vars.Children {
		sym.Put(symtable.Integer, node.Tok.Lex)
	}
	for _, node := range proc.Children {
		iden := node.Children[0]
		bloc := node.Children[1]
		sym.Put(symtable.Procedure, iden.Tok.Lex)
		// Recursively load on inner procedures.
		a.loadSymbolTables(bloc)

	}
	node.Sym = sym
}

func (a *Analyser) recurseProgramCheck(node *ast.Node) {
	a.recurseBlockCheck(node.Children[0], make([]*symtable.SymbolTable, 0))
}

func (a *Analyser) recurseConstCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	// Don't really care about the naming of constants... They come before the vars names
	// anyway.
}

func (a *Analyser) recurseVarCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	// If the immediate parent symbol table has constants of the same name, then there's an
	// ambiguity issue.
	for _, node := range node.Children {
		if a.findSymbolInTables(node.Tok.Lex, symtable.Constant, syms) {
			a.appendError(node.Tok)
		}
	}
}

func (a *Analyser) recurseProcedureCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	for _, node := range node.Children {
		id := node.Children[0]
		if !a.findSymbolInTables(id.Tok.Lex, symtable.Procedure, syms) {
			a.appendError(id.Tok)
		}
		bloc := node.Children[1]
		a.recurseBlockCheck(bloc, syms)
	}
}

func (a *Analyser) recurseBlockCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	syms = append(syms, node.Sym)
	a.recurseConstCheck(node.Children[0], syms)
	a.recurseVarCheck(node.Children[1], syms)
	a.recurseProcedureCheck(node.Children[2], syms)
	a.recurseStatementCheck(node.Children[3], syms)
}

func (a *Analyser) recurseStatementCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	if node.Tag == ast.Assignment {
		a.assignmentCheck(node, syms)
	} else if node.Tag == ast.Call {
		a.callCheck(node, syms)
	} else if node.Tag == ast.Begin {
		for _, node := range node.Children {
			a.recurseStatementCheck(node, syms)
		}
	} else if node.Tag == ast.IfThen {
		a.ifThenCheck(node, syms)
	} else if node.Tag == ast.WhileDo {
		a.whileDoCheck(node, syms)
	} else {
		// This shouldn't happen ever...
		a.appendError(node.Tok)
	}
}

func (a *Analyser) assignmentCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	iden := node.Children[0]
	expr := node.Children[1]
	a.recurseExpressionCheck(expr, syms)
	if !a.findSymbolInTables(iden.Tok.Lex, symtable.Integer, syms) {
		a.appendError(iden.Tok)
	}
}

func (a *Analyser) callCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	iden := node.Children[0]
	if !a.findSymbolInTables(iden.Tok.Lex, symtable.Procedure, syms) {
		a.appendError(iden.Tok)
	}
}

func (a *Analyser) ifThenCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	a.recurseExpressionCheck(node.Children[0], syms)
	a.recurseStatementCheck(node.Children[1], syms)
}

func (a *Analyser) whileDoCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	a.recurseExpressionCheck(node.Children[0], syms)
	a.recurseStatementCheck(node.Children[1], syms)
}

func (a *Analyser) recurseExpressionCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	if node.Tag == ast.Terminal {
		// Only look through the symbol table if it's an idenfitier!
		if node.Tok.Tag == token.Identifier {
			if !a.findSymbolInTables(node.Tok.Lex, symtable.Integer, syms) &&
				!a.findSymbolInTables(node.Tok.Lex, symtable.Constant, syms) {
				a.appendError(node.Tok)
			}
		}
		return
	}
	left := node.Children[0]
	a.recurseExpressionCheck(left, syms)
	right := node.Children[1]
	a.recurseExpressionCheck(right, syms)
}

func (a *Analyser) recurseConditionCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	if node.Tag == ast.Cond {
		a.recurseExpressionCheck(node.Children[0], syms)
		a.recurseExpressionCheck(node.Children[1], syms)
	} else if node.Tag == ast.Odd {
		a.recurseExpressionCheck(node.Children[0], syms)
	}
}

func (a *Analyser) findSymbolInTables(lex string, symbol int, syms []*symtable.SymbolTable) bool {
	// Go backwards so we search the closest table first. I don't think this matters, but it's
	// better for clarity.
	for i := len(syms) - 1; i >= 0; i-- {
		if syms[i].Get(symbol, lex) {
			return true
		}
	}
	return false
}

func (a *Analyser) appendError(tok *token.Token) {
	a.err = append(a.err, fmt.Errorf("Semantic error near line %d.", tok.Ln))
}
