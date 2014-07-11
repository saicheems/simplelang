// Package analyser implements some basic semantic analysis on the parse tree.
package analyser

import (
	"fmt"

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
func (a *Analyser) loadSymbolTables(ast *parser.Node) {
	sym := symtable.New()
	cons := ast.Children[0] // Constants
	vars := ast.Children[1] // Vars
	proc := ast.Children[2] // Procedures

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
	ast.Sym = sym
}

func (a *Analyser) recurseProgramCheck(ast *parser.Node) {
	a.recurseBlockCheck(ast.Children[0], make([]*symtable.SymbolTable, 0))
}

func (a *Analyser) recurseConstCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	// Don't really care about the naming of constants... They come before the vars names
	// anyway.
}

func (a *Analyser) recurseVarCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	// If the immediate parent symbol table has constants of the same name, then there's an
	// ambiguity issue.
	for _, node := range ast.Children {
		if a.findSymbolInTables(node.Tok.Lex, symtable.Constant, syms) {
			a.appendError(node.Tok)
		}
	}
}

func (a *Analyser) recurseProcedureCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	for _, node := range ast.Children {
		bloc := node.Children[1]
		a.recurseBlockCheck(bloc, syms)
	}
}

func (a *Analyser) recurseBlockCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	syms = append(syms, ast.Sym)
	a.recurseConstCheck(ast.Children[0], syms)
	a.recurseVarCheck(ast.Children[1], syms)
	a.recurseProcedureCheck(ast.Children[2], syms)
	a.recurseStatementCheck(ast.Children[3], syms)
}

func (a *Analyser) recurseStatementCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	for _, node := range ast.Children {
		if node.Type == parser.TypeAssignment {
			a.assignmentCheck(node, syms)
		} else if node.Type == parser.TypeCall {
			a.callCheck(node, syms)
		} else if node.Type == parser.TypeBegin {
			a.recurseStatementCheck(node, syms)
		} else if node.Type == parser.TypeIfThen {
			a.ifThenCheck(node, syms)
		} else if node.Type == parser.TypeWhileDo {
			a.whileDoCheck(node, syms)
		} else {
			// This shouldn't happen ever...
		}
	}
}

func (a *Analyser) assignmentCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	iden := ast.Children[0]
	expr := ast.Children[1]
	a.recurseExpressionCheck(expr, syms)
	if !a.findSymbolInTables(iden.Tok.Lex, symtable.Integer, syms) {
		a.appendError(iden.Tok)
	}
}

func (a *Analyser) callCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	iden := ast.Children[0]
	if !a.findSymbolInTables(iden.Tok.Lex, symtable.Integer, syms) {
		a.appendError(iden.Tok)
	}
	a.appendError(iden.Tok)
}

func (a *Analyser) ifThenCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	a.recurseExpressionCheck(ast.Children[0], syms)
	a.recurseStatementCheck(ast.Children[1], syms)
}

func (a *Analyser) whileDoCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	a.recurseExpressionCheck(ast.Children[0], syms)
	a.recurseStatementCheck(ast.Children[1], syms)
}

func (a *Analyser) recurseExpressionCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	if ast.Type == parser.TypeTerminal {
		// Only look through the symbol table if it's an idenfitier!
		if ast.Tok.Tag == token.Identifier {
			if !a.findSymbolInTables(ast.Tok.Lex, symtable.Integer, syms) &&
				!a.findSymbolInTables(ast.Tok.Lex, symtable.Constant, syms) {
				a.appendError(ast.Tok)
			}
		}
		return
	}
	left := ast.Children[0]
	a.recurseExpressionCheck(left, syms)
	right := ast.Children[1]
	a.recurseExpressionCheck(right, syms)
}

func (a *Analyser) recurseConditionCheck(ast *parser.Node, syms []*symtable.SymbolTable) {
	if ast.Type == parser.TypeCond {
		a.recurseExpressionCheck(ast.Children[0], syms)
		a.recurseExpressionCheck(ast.Children[1], syms)
	} else if ast.Type == parser.TypeOdd {
		a.recurseExpressionCheck(ast.Children[0], syms)
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
