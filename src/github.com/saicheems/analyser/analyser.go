// Package analyser implements some basic semantic analysis on the parse tree.
package analyser

import (
	"fmt"

	"github.com/saicheems/ast"
	"github.com/saicheems/parser"
	"github.com/saicheems/symtable"
	"github.com/saicheems/token"
)

// Analyser implements the semantic analysis stage of the compilation.
type Analyser struct {
	par *parser.Parser
	err []error
}

// New returns a new Analyser.
func New(par *parser.Parser) *Analyser {
	a := new(Analyser)
	a.par = par
	a.err = make([]error, 0)
	return a
}

// Analyse returns the abstract syntax tree if the semantic analysis is successful. Otherwise it
// returns nil.
func (a *Analyser) Analyse() *ast.Node {
	ast := a.par.Parse()
	if ast == nil {
		return nil
	}
	a.loadSymbolTables(ast.Children[0])
	a.recurseProgramCheck(ast)
	if len(a.err) > 0 {
		for _, err := range a.err {
			fmt.Println(err)
		}
		return nil
	}
	return ast
}

// loadSymbolTables Loads all of the symbol tables. In this simple language all symbols should be
// defined in the header of the program, so it is an easy pass.
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

// recurseProgramCheck recurses on the top node in the AST (the program node).
func (a *Analyser) recurseProgramCheck(node *ast.Node) {
	a.recurseBlockCheck(node.Children[0], make([]*symtable.SymbolTable, 0))
}

// recurseConstCheck recurses on the const node.
func (a *Analyser) recurseConstCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	// Don't really care about the naming of constants... They come before the vars names
	// anyway.
}

// recurseVarCheck recurses on the var node.
func (a *Analyser) recurseVarCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	// If the immediate parent symbol table has constants of the same name, then there's an
	// ambiguity issue.
	for _, node := range node.Children {
		if a.findSymbolInTables(node.Tok.Lex, symtable.Constant, syms) {
			a.appendError(node.Tok)
		}
	}
}

// recurseProcedureCheck recurses on the procedure node.
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

// recurseBlockCheck recurses on a block.
func (a *Analyser) recurseBlockCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	syms = append(syms, node.Sym)
	a.recurseConstCheck(node.Children[0], syms)
	a.recurseVarCheck(node.Children[1], syms)
	a.recurseProcedureCheck(node.Children[2], syms)
	a.recurseStatementCheck(node.Children[3], syms)
}

// recurseStatementCheck recurses on a statement.
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

// assignmentCheck validates an assigment.
func (a *Analyser) assignmentCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	iden := node.Children[0]
	expr := node.Children[1]
	a.recurseExpressionCheck(expr, syms)
	if !a.findSymbolInTables(iden.Tok.Lex, symtable.Integer, syms) {
		a.appendError(iden.Tok)
	}
}

// callCheck validates a call.
func (a *Analyser) callCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	iden := node.Children[0]
	if !a.findSymbolInTables(iden.Tok.Lex, symtable.Procedure, syms) {
		a.appendError(iden.Tok)
	}
}

// ifThenCheck validates an if then statement.
func (a *Analyser) ifThenCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	a.recurseExpressionCheck(node.Children[0], syms)
	a.recurseStatementCheck(node.Children[1], syms)
}

// whileDoCheck validates a while do statement.
func (a *Analyser) whileDoCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	a.recurseExpressionCheck(node.Children[0], syms)
	a.recurseStatementCheck(node.Children[1], syms)
}

// recurseExpressionCheck recurses on an expression.
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

// recurseConditionCheck recurses on a condition.
func (a *Analyser) recurseConditionCheck(node *ast.Node, syms []*symtable.SymbolTable) {
	if node.Tag == ast.Cond {
		a.recurseExpressionCheck(node.Children[0], syms)
		a.recurseExpressionCheck(node.Children[1], syms)
	} else if node.Tag == ast.Odd {
		a.recurseExpressionCheck(node.Children[0], syms)
	}
}

// findSymbolInTables returns a bool representing whether or not a symbol was found in the list of
// symbol tables provided.
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

// appendError takes in a Token and appends a semantic error at the Token's line number to the
// Analyser's error list.
func (a *Analyser) appendError(tok *token.Token) {
	a.err = append(a.err, fmt.Errorf("Semantic error near line %d.", tok.Ln))
}
