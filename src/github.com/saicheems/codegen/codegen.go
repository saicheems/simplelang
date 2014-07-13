// Package codegen implements the code generation phase of the compilation. It generates MIPS
// assembly code from the semantically correct abstract syntax tree provided by the analyser.
package codegen

import (
	"bytes"
	"fmt"

	"github.com/saicheems/analyser"
	"github.com/saicheems/ast"
	"github.com/saicheems/symtable"
)

type entry struct {
	sym  *symtable.SymbolTable
	name string
}

// CodeGenerator implements the code generation phase of the compilation.
type CodeGenerator struct {
	a     *analyser.Analyser
	buf   *bytes.Buffer
	info  map[entry]string
	count int
}

// New returns a new Analyer that prints to stdout. TODO: change to file...
func New(a *analyser.Analyser) *CodeGenerator {
	c := new(CodeGenerator)
	c.a = a
	c.info = make(map[entry]string)
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
	c.generateProcedure(proc, bloc.Sym)
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
	c.generateStatement(stmt, bloc.Sym)
	c.emitLoadInt("$v0", 10)
	c.emitSyscall()
}

// generateBlock begins generation of a block node.
func (c *CodeGenerator) generateBlock(node *ast.Node) {
	// We won't bother with constants here - just insert their values into the assembly
	// instructions automatically.
	// c.generateConst(node)
	// We won't bother with vars here either - they're taken care of in generateProgram for the
	// statements in main, and in generateProcedure for any vars of procedures.
	// c.generateVar(node)
	c.generateProcedure(node.Children[2], node.Sym)
	c.generateStatement(node.Children[3], node.Sym)
}

// generateBlock begins generation of a procedure node. We pass in the number of vars declared
// before the function so we can add the right number of variables to the stack.
func (c *CodeGenerator) generateProcedure(node *ast.Node, sym *symtable.SymbolTable) {
	for _, node := range node.Children {
		bloc := node.Children[1]
		// Find out how many variables we have so we can set up the activation record.
		numVars := len(bloc.Children[1].Children)
		// Emit the procedure label.
		label := c.emitNewProcedureLabel()
		c.info[entry{sym, node.Children[0].Tok.Lex}] = label
		// Store the old frame pointer on the stack.
		c.emitStoreWord("$fp", "$sp", 0)
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
		c.generateBlock(bloc)
		// Emit the done tag for the function.
		c.emitLabel(label + "_done")
		// Load the return address from the stack.
		c.emitLoadWord("$ra", "$sp", 4)
		// Reset the stack to the original position.
		c.emitAddUnsigned("$sp", "$sp", 4*numVars+8)
		// Load the old frame pointer.
		c.emitLoadWord("$fp", "$sp", 0)
		c.emitJumpReturn()
	}
}

func (c *CodeGenerator) generateStatement(node *ast.Node, sym *symtable.SymbolTable) {
	if node.Tag == ast.Assignment {
	} else if node.Tag == ast.Call {
		id := node.Children[0]
		c.emitJumpAndLink(c.info[entry{sym, id.Tok.Lex}])
	} else if node.Tag == ast.Begin {
		for _, node := range node.Children {
			c.generateStatement(node, sym)
		}
	} else if node.Tag == ast.IfThen {
	} else if node.Tag == ast.WhileDo {
	} else {
		// This can't possibly happen...
		fmt.Println("A terrible error occurred.",
			"The abstract syntax tree is wrong and I'm generating code...")
	}
}

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

// emitSubtractUnsigned emits a subu instruction.
func (c *CodeGenerator) emitAddUnsigned(target string, source string, val int) {
	c.writeOut(fmt.Sprintf("addu %s %s %d\n", target, source, val))
}

// emitSubtractUnsigned emits a subu instruction.
func (c *CodeGenerator) emitSubtractUnsigned(target string, source string, val int) {
	c.writeOut(fmt.Sprintf("subu %s %s %d\n", target, source, val))
}

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

// emitLabel takes a label name and writes the assembly form.
func (c *CodeGenerator) emitLabel(label string) {
	c.writeOut(label + ":\n")
}

// emitSyscall emits the syscall instruction.
func (c *CodeGenerator) emitSyscall() {
	c.writeOut("syscall\n")
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