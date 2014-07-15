// Package codegen implements the code generation phase of the compilation. It generates MIPS
// assembly code from the semantically correct abstract syntax tree provided by the analyser.
package codegen

import (
	"bytes"
	"fmt"

	"github.com/saicheems/analyser"
	"github.com/saicheems/ast"
	"github.com/saicheems/symtable"
	"github.com/saicheems/token"
)

type entry struct {
	sym  *symtable.SymbolTable
	name string
}

// CodeGenerator implements the code generation phase of the compilation.
type CodeGenerator struct {
	a     *analyser.Analyser
	buf   *bytes.Buffer
	count int
}

// New returns a new Analyer that prints to stdout. TODO: change to file...
func New(a *analyser.Analyser) *CodeGenerator {
	c := new(CodeGenerator)
	c.a = a
	return c
}

// NewToString returns a new Analyer that prints to a byte buffer.
func NewToString(a *analyser.Analyser) *CodeGenerator {
	c := new(CodeGenerator)
	c.a = a
	c.buf = bytes.NewBufferString("")
	return c
}

// Returns the string in the CodeGenerator's buffer.
func (c *CodeGenerator) ToString() string {
	return c.buf.String()
}

// Generate uses the abstract syntax tree returned by the Analyser and begins code generation if the
// tree is not nil. Otherwise it returns.
func (c *CodeGenerator) Generate() {
	ast := c.a.Analyse()
	if ast == nil {
		return
	}
	c.generateProgram(ast)
}

// generateProgram begins generation at the head node.
func (c *CodeGenerator) generateProgram(node *ast.Node) {
	bloc := node.Children[0]
	vars := bloc.Children[1]
	proc := bloc.Children[2]
	stmt := bloc.Children[3]
	// We'll lay out the procedures first at the top of the assembly output.
	c.generateProcedure(proc, []*symtable.SymbolTable{bloc.Sym}, 0)
	c.emitMainLabel()
	numVars := len(vars.Children)

	// Set up the current frame pointer.
	c.emitMove("$fp", "$sp")
	// Load all the variables in this scope onto the current frame. Initialize to 0.
	for i := 0; i < numVars; i++ {
		c.emitLoadInt("$a0", 0)
		c.emitStoreWord("$a0", "$sp", 0)
		c.emitSubtractUnsigned("$sp", "$sp", 4)
	}
	c.generateStatement(stmt, []*symtable.SymbolTable{bloc.Sym})
	c.emitLoadInt("$v0", 10)
	c.emitSyscall()
}

// generateBlock begins generation of a block node.
func (c *CodeGenerator) generateBlock(node *ast.Node, sym []*symtable.SymbolTable, level int) {
	// We won't bother with constants here - just insert their values into the assembly
	// instructions automatically.
	// c.generateConst(node)
	// We won't bother with vars here either - they're taken care of in generateProgram for the
	// statements in main, and in generateProcedure for any vars of procedures.
	// c.generateVar(node)
	sym = append(sym, node.Sym)
	c.generateProcedure(node.Children[2], sym, level)
	c.generateStatement(node.Children[3], sym)
}

// generateBlock begins generation of a procedure node. We pass in the number of vars declared
// before the function so we can add the right number of variables to the stack.
func (c *CodeGenerator) generateProcedure(node *ast.Node, sym []*symtable.SymbolTable, level int) {
	for _, node := range node.Children {
		id := node.Children[0]
		bloc := node.Children[1]
		// Find out how many variables we have so we can set up the activation record.
		numVars := len(bloc.Children[1].Children)
		// Emit the procedure label.
		label := c.emitNewProcedureLabel()
		c.getClosestSymbolTable(sym).Put(symtable.Symbol{symtable.Procedure, id.Tok.Lex},
			&symtable.Value{label, 0, 0})
		// Store the old frame pointer on the stack. DYNAMIC LINK.
		c.emitStoreWord("$fp", "$sp", 0)
		c.emitSubtractUnsigned("$sp", "$sp", 4)

		// Calculate the STATIC LINK.
		c.emitMove("$a0", "$fp") // Points to frame of main if we're at depth 0.
		for i := 0; i < level; i++ {
			c.emitLoadWord("$a0", "$a0", 4)
		}
		// Store the STATIC LINK on the stack.
		c.emitStoreWord("$a0", "$sp", 0)
		c.emitSubtractUnsigned("$sp", "$sp", 4)

		// Have the new frame pointer point to the stack.
		c.emitMove("$fp", "$sp")
		// Load all the variables in this scope onto the current frame. Initialize to 0.
		for i := 0; i < numVars; i++ {
			c.emitLoadInt("$a0", 0)
			c.emitStoreWord("$a0", "$sp", 0)
			c.emitSubtractUnsigned("$sp", "$sp", 4)
		}
		// Store the return address on the stack.
		c.emitStoreWord("$ra", "$sp", 0)
		c.emitSubtractUnsigned("$sp", "$sp", 4)
		// Generate code for the body.
		c.generateBlock(bloc, sym, level+1)
		// Emit the done tag for the function.
		c.emitLabel(label + "_done")
		// Load the return address from the stack.
		c.emitLoadWord("$ra", "$sp", 4)
		// Reset the stack to the original position.
		c.emitAddUnsigned("$sp", "$sp", 4*numVars+12)
		// Load the old frame pointer.
		c.emitLoadWord("$fp", "$sp", 0)
		c.emitJumpReturn()
	}
}

func (c *CodeGenerator) generateStatement(node *ast.Node, syms []*symtable.SymbolTable) {
	if node.Tag == ast.Assignment {
		iden := node.Children[0]
		n, s := c.getClosestSymbolTableWithSymbol(symtable.Symbol{symtable.Integer,
			iden.Tok.Lex}, syms)
		// Indicates which variable on the frame corresponds to the left hand side.
		o := s.Get(symtable.Symbol{symtable.Integer, iden.Tok.Lex}).Order

		c.loadAddressOfPreviousFrame("$t0", n, o)

		// Rest of stuff.
		c.generateExpression(node.Children[1], syms)

		c.emitAddUnsigned("$sp", "$sp", 4)
		c.emitLoadWord("$a0", "$sp", 0) // Load result onto $a0
		c.emitStoreWord("$a0", "$t0", 0)
	} else if node.Tag == ast.Call {
		id := node.Children[0]
		_, s := c.getClosestSymbolTableWithSymbol(
			symtable.Symbol{symtable.Procedure, id.Tok.Lex}, syms)
		label := s.Get(symtable.Symbol{symtable.Procedure, id.Tok.Lex}).Label
		c.emitJumpAndLink(label)
	} else if node.Tag == ast.Begin {
		for _, node := range node.Children {
			c.generateStatement(node, syms)
		}
	} else if node.Tag == ast.IfThen {
		cond := node.Children[0]
		stmt := node.Children[1]

		ifLabel := c.getNewLabel("if")
		doneLabel := c.getNewLabel("done")
		c.count++

		c.generateCondition(cond, ifLabel, syms)

		c.emitJump(doneLabel)
		c.emitLabel(ifLabel)
		c.generateStatement(stmt, syms)
		c.emitLabel(doneLabel)
	} else if node.Tag == ast.WhileDo {
		cond := node.Children[0]
		stmt := node.Children[1]
		whileLabel := c.getNewLabel("while")
		doLabel := c.getNewLabel("do")
		doneLabel := c.getNewLabel("done")
		c.count++
		c.emitLabel(whileLabel)
		c.generateCondition(cond, doLabel, syms)
		c.emitJump(doneLabel)
		c.emitLabel(doLabel)
		c.generateStatement(stmt, syms)
		c.emitJump(whileLabel)
		c.emitLabel(doneLabel)
	} else if node.Tag == ast.Print {
		expr := node.Children[0]
		c.generateExpression(expr, syms)
		c.emitAddUnsigned("$sp", "$sp", 4)
		c.emitLoadWord("$a0", "$sp", 0)
		c.emitLoadInt("$v0", 1)
		c.emitSyscall()
	} else {
		// This can't possibly happen...
		fmt.Println("A terrible error occurred.",
			"The abstract syntax tree is wrong and I'm generating code...")
	}
}

func (c *CodeGenerator) generateCondition(node *ast.Node, label string,
	syms []*symtable.SymbolTable) {
	if node.Tag == ast.Odd {
		// TODO: Will do.... later.
		c.generateExpression(node.Children[0], syms)
		c.emitAddUnsigned("$sp", "$sp", 4)
		c.emitLoadWord("$t1", "$sp", 0)
		c.emitAndImmediate("$t1", "$t1", 1)
		c.emitBranchOnGreaterThanZero("$t1", label)
		return
	}

	c.generateExpression(node.Children[0], syms)
	c.generateExpression(node.Children[1], syms)
	// Load the expressions.
	c.emitAddUnsigned("$sp", "$sp", 4)
	c.emitLoadWord("$t1", "$sp", 0)
	c.emitAddUnsigned("$sp", "$sp", 4)
	c.emitLoadWord("$t2", "$sp", 0)
	//c.emitAddUnsigned("$sp", "$sp", 4)
	if node.Op == token.Equals {
		c.emitBranchOnEqual("$t1", "$t2", label)
	} else if node.Op == token.NotEquals {
		c.emitBranchNotEqual("$t1", "$t2", label)
	} else if node.Op == token.LessThan {
		c.emitSubtract("$t1", "$t1", "$t2")
		c.emitBranchOnGreaterThanZero("$t1", label)
	} else if node.Op == token.GreaterThan {
		c.emitSubtract("$t1", "$t2", "$t1")
		c.emitBranchOnGreaterThanZero("$t1", label)
	} else if node.Op == token.LessThanEqualTo {
		c.emitSubtract("$t1", "$t1", "$t2")
		c.emitBranchOnGreaterThanOrEqualZero("$t1", label)
	} else if node.Op == token.GreaterThanEqualTo {
		c.emitSubtract("$t1", "$t2", "$t1")
		c.emitBranchOnGreaterThanOrEqualZero("$t1", label)
	} else {
		// This can't possibly happen...
		fmt.Println("A terrible error occurred.",
			"The abstract syntax tree is wrong and I'm generating code...")
	}
}

// generateExpression evaluates an expression and places the result on the stack.
func (c *CodeGenerator) generateExpression(node *ast.Node, syms []*symtable.SymbolTable) {
	if node.Tag == ast.Terminal {
		// Only look through the symbol table if it's an idenfitier!
		if node.Tok.Tag == token.Identifier {
			n, s := c.getClosestSymbolTableWithSymbol(symtable.Symbol{symtable.Integer,
				node.Tok.Lex}, syms)
			if s == nil {
				_, s := c.getClosestSymbolTableWithSymbol(
					symtable.Symbol{symtable.Constant, node.Tok.Lex}, syms)
				val := s.Get(symtable.Symbol{symtable.Constant, node.Tok.Lex}).Val
				// It's a constant if we can't find the symbol. TODO: clean.
				c.emitLoadInt("$a0", val)
				c.emitStoreWord("$a0", "$sp", 0)
				c.emitSubtractUnsigned("$sp", "$sp", 4)
				return
			}
			// Indicates which variable on the frame corresponds to the left hand side.
			o := s.Get(symtable.Symbol{symtable.Integer, node.Tok.Lex}).Order
			c.loadAddressOfPreviousFrame("$a0", n, o)
			c.emitLoadWord("$a0", "$a0", 0)
			c.emitStoreWord("$a0", "$sp", 0)
			c.emitSubtractUnsigned("$sp", "$sp", 4)
		} else if node.Tok.Tag == token.Integer {
			c.emitLoadInt("$a0", node.Tok.Val)
			c.emitStoreWord("$a0", "$sp", 0)
			c.emitSubtractUnsigned("$sp", "$sp", 4)
		} else {
			// This can't possibly happen...
			fmt.Println("A terrible error occurred.",
				"The abstract syntax tree is wrong and I'm generating code...")
		}
		return
	}
	left := node.Children[0]
	c.generateExpression(left, syms)
	right := node.Children[1]
	c.generateExpression(right, syms)

	c.emitAddUnsigned("$sp", "$sp", 4)
	c.emitLoadWord("$t1", "$sp", 0)
	c.emitAddUnsigned("$sp", "$sp", 4)
	c.emitLoadWord("$t2", "$sp", 0)
	if node.Op == token.Plus {
		c.emitAdd("$t1", "$t1", "$t2")
	} else if node.Op == token.Minus {
		// TODO: Does this cause any issues?
		c.emitSubtract("$t1", "$t2", "$t1")
	} else if node.Op == token.Times {
		c.emitMult("$t1", "$t2")
		c.emitMoveFromLo("$t1")
	} else if node.Op == token.Divide {
		// TODO: Does this cause any issues?
		c.emitDiv("$t2", "$t1")
		c.emitMoveFromLo("$t1")
	} else {
		// This can't possibly happen...
		fmt.Println("A terrible error occurred.",
			"The abstract syntax tree is wrong and I'm generating code...")
	}
	c.emitStoreWord("$t1", "$sp", 0)
	c.emitSubtractUnsigned("$sp", "$sp", 4)
}

// loadAddressOfPreviousFrame loads the address of the variable n levels back at position o into
// register dest.
func (c *CodeGenerator) loadAddressOfPreviousFrame(dest string, n int, o int) {
	c.emitMove(dest, "$fp")
	// Need to go back n levels to position o.
	for i := 0; i < n; i++ { // If we need to go back.
		c.emitLoadWord(dest, dest, 4) // Load old frame pointer.
	}
	c.emitSubtractUnsigned(dest, dest, 4*o)
}

// emitAndImmediate emits a andi instruction.
func (c *CodeGenerator) emitAndImmediate(dest string, source string, val int) {
	c.writeOut(fmt.Sprintf("andi %s %s %d\n", dest, source, val))
}

// emitBranchOnGreaterThanZero emits a bgtz instruction.
func (c *CodeGenerator) emitBranchOnGreaterThanZero(source string, label string) {
	c.writeOut(fmt.Sprintf("bgtz %s %s\n", source, label))
}

// emitBranchOnGreaterThanOrEqualZero emits a bgez instruction.
func (c *CodeGenerator) emitBranchOnGreaterThanOrEqualZero(source string, label string) {
	c.writeOut(fmt.Sprintf("bgez %s %s\n", source, label))
}

// emitBranchNotEquals emits a bne instruction.
func (c *CodeGenerator) emitBranchOnEqual(source1 string, source2 string, label string) {
	c.writeOut(fmt.Sprintf("beq %s %s %s\n", source1, source2, label))
}

// emitBranchNotEquals emits a bne instruction.
func (c *CodeGenerator) emitBranchNotEqual(source1 string, source2 string, label string) {
	c.writeOut(fmt.Sprintf("bne %s %s %s\n", source1, source2, label))
}

// emitJumAndLink emits a j instruction to some label.
func (c *CodeGenerator) emitJump(label string) {
	c.writeOut(fmt.Sprintf("j %s\n", label))
}

// emitJumAndLink emits a jal instruction to some label.
func (c *CodeGenerator) emitJumpAndLink(label string) {
	c.writeOut(fmt.Sprintf("jal %s\n", label))
}

// emitJumpReturn emits a jr instruction to $ra.
func (c *CodeGenerator) emitJumpReturn() {
	// I'm pretty sure we'll only jr to $ra.
	c.writeOut("jr $ra\n")
}

// emitLoadWord emits a lw instruction.
func (c *CodeGenerator) emitLoadWord(source string, target string, offset int) {
	c.writeOut(fmt.Sprintf("lw %s %d(%s)\n", source, offset, target))
}

// emitLoadInt emits a li instruction.
func (c *CodeGenerator) emitLoadInt(target string, val int) {
	c.writeOut(fmt.Sprintf("li %s %d\n", target, val))
}

// emitStoreWord emits a sw instruction.
func (c *CodeGenerator) emitStoreWord(source string, target string, offset int) {
	c.writeOut(fmt.Sprintf("sw %s %d(%s)\n", source, offset, target))
}

// emitAddUnsigned emits a addu instruction.
func (c *CodeGenerator) emitAddUnsigned(target string, source string, val int) {
	c.writeOut(fmt.Sprintf("addu %s %s %d\n", target, source, val))
}

// emitSubtractUnsigned emits a subu instruction.
func (c *CodeGenerator) emitSubtractUnsigned(target string, source string, val int) {
	c.writeOut(fmt.Sprintf("subu %s %s %d\n", target, source, val))
}

// emitAdd emits a add instruction.
func (c *CodeGenerator) emitAdd(target string, source string, source2 string) {
	c.writeOut(fmt.Sprintf("add %s %s %s\n", target, source, source2))
}

// emitSubtract emits a sub instruction.
func (c *CodeGenerator) emitSubtract(target string, source string, source2 string) {
	c.writeOut(fmt.Sprintf("sub %s %s %s\n", target, source, source2))
}

// emitMult emits a mult instruction.
func (c *CodeGenerator) emitMult(source1 string, source2 string) {
	c.writeOut(fmt.Sprintf("mult %s %s\n", source1, source2))
}

// emitMult emits a mult instruction.
func (c *CodeGenerator) emitDiv(source1 string, source2 string) {
	c.writeOut(fmt.Sprintf("div %s %s\n", source1, source2))
}

func (c *CodeGenerator) emitMoveFromLo(target string) {
	c.writeOut(fmt.Sprintf("mflo %s\n", target))
}

// emitMove emits a move instruction.
func (c *CodeGenerator) emitMove(target string, source string) {
	c.writeOut(fmt.Sprintf("move %s %s\n", target, source))
}

// emitMainLabel emits the main: label.
func (c *CodeGenerator) emitMainLabel() {
	c.writeOut("main:\n")
}

// emitNewProcedureLabel returns a unique procedure label using the global count. The global
// count is then incremented.
func (c *CodeGenerator) emitNewProcedureLabel() string {
	label := fmt.Sprintf("procedure%d", c.count)
	c.emitLabel(label)
	c.count++
	return label
}

func (c *CodeGenerator) emitCustomLabel(cus string) string {
	label := fmt.Sprintf("if%d", c.count)
	c.emitLabel(label)
	c.count++
	return label
}

func (c *CodeGenerator) getNewLabel(base string) string {
	return fmt.Sprintf("%s%d", base, c.count)
}

// emitLabel takes a label name and writes the assembly form.
func (c *CodeGenerator) emitLabel(label string) {
	c.writeOut(label + ":\n")
}

// emitSyscall emits the syscall instruction.
func (c *CodeGenerator) emitSyscall() {
	c.writeOut("syscall\n")
}

// getClosestSymbolTable returns the closest symbol table to the current context in an array of
// SymbolTables.
func (c *CodeGenerator) getClosestSymbolTable(s []*symtable.SymbolTable) *symtable.SymbolTable {
	return s[len(s)-1]
}

func (c *CodeGenerator) getClosestSymbolTableWithSymbol(sym symtable.Symbol,
	syms []*symtable.SymbolTable) (int, *symtable.SymbolTable) {
	for i := len(syms) - 1; i >= 0; i-- {
		if syms[i].Get(sym) != nil {
			return len(syms) - 1 - i, syms[i]
		}
	}
	return 0, nil
}

// writeOut takes a string and prints it to stdout for now. Should eventually print to a file.
// TODO: Print to a file.
func (c *CodeGenerator) writeOut(s string) {
	if c.buf != nil {
		c.buf.WriteString(s)
	} else {
		fmt.Print(s)
	}
}
