// Package codegen implements the code generation phase of the compilation. It generates MIPS
// assembly code from the semantically correct abstract syntax tree provided by the analyser.
package codegen

import (
	"bytes"
	"fmt"

	"github.com/saicheems/simplelang/analyser"
	"github.com/saicheems/simplelang/ast"
	"github.com/saicheems/simplelang/symtable"
	"github.com/saicheems/simplelang/token"
)

// CodeGenerator implements the code generation phase of the compilation.
type CodeGenerator struct {
	a     *analyser.Analyser
	buf   *bytes.Buffer // Byte buffer for the output of the code generation.
	count int           // Global label count: ensures labels are unique.
}

// New returns a new Analyer that prints to the internal byte buffer.
func New(a *analyser.Analyser) *CodeGenerator {
	c := new(CodeGenerator)
	c.a = a
	c.buf = bytes.NewBufferString("")
	return c
}

// String returns the contents of the CodeGenerator's buffer as a string.
func (c *CodeGenerator) String() string {
	return c.buf.String()
}

// Generate uses the abstract syntax tree returned by the Analyser and begins code generation if the
// tree is not nil. Otherwise it returns. The output assembly is written to the CodeGenerator's byte
// buffer.
func (c *CodeGenerator) Generate() {
	node := c.a.Analyse()
	if node == nil {
		return
	}
	c.generateProgram(node)
}

// generateProgram begins generation at the program node. It generates the code for procedure
// definitions and places them at the head of the assembly. It also sets up the top level vars and
// generates the top level statement.
func (c *CodeGenerator) generateProgram(node *ast.Node) {
	bloc := node.Children[0]
	vars := bloc.Children[1]
	proc := bloc.Children[2]
	stmt := bloc.Children[3]

	syms := []*symtable.SymbolTable{bloc.Sym}
	// We'll lay out the procedures first at the top of the assembly output.
	c.generateProcedure(proc, syms)
	c.emitLabel("main")
	// Set up the current frame pointer.
	c.emitMove("$fp", "$sp")
	// Load all the variables in this scope onto the current frame. Initialize to 0.
	for i := 0; i < len(vars.Children); i++ {
		c.emitLoadInt("$a0", 0)
		c.emitStoreWord("$a0", "$sp", 0)
		c.emitSubUnsigned("$sp", "$sp", 4)
	}
	// Generate the main statement.
	c.generateStatement(stmt, syms)
	// Generate exit syscall at the end of the program.
	c.emitLoadInt("$v0", 10)
	c.emitSyscall()
}

// generateProcedure begins generation of a procedure node. It generates the definition of the
// current procedure and any nested procedures within. It does not set up the stack for a function
// call. That is left to be done at a CALL statement.
func (c *CodeGenerator) generateProcedure(node *ast.Node, syms []*symtable.SymbolTable) {
	for _, node := range node.Children {
		iden := node.Children[0]
		bloc := node.Children[1]
		// Find out how many variables we have so we can set up the activation record.
		numVars := len(bloc.Children[1].Children)
		// Emit the procedure label.
		label := c.getNewLabel("procedure")
		c.emitLabel(label)
		bodyLabel := label + "_body" // Label of the procedure body.
		doneLabel := label + "_done" // Label of the procedure end.
		// Store the return address on the stack.
		c.emitStoreWord("$ra", "$sp", 0)
		c.emitSubUnsigned("$sp", "$sp", 4)

		key := symtable.Key{symtable.Procedure, iden.Tok.Lex}
		// Value includes the procedure label and how many arguments it has.
		value := symtable.Value{label, 0, 0, numVars}
		syms[len(syms)-1].Put(key, &value) // Write the value with the new info back.
		// Jump to the body so we don't prematurely execute nested procedures.
		c.emitJump(bodyLabel)
		// Generate any nested procedures.
		nestSyms := append(syms, bloc.Sym)
		c.generateProcedure(bloc.Children[2], nestSyms)
		// Generate code for the body.
		c.emitLabel(bodyLabel)
		c.generateStatement(bloc.Children[3], nestSyms)
		// Emit the done tag for the function.
		c.emitLabel(doneLabel)
		// Load the return address from the stack.
		c.emitLoadWord("$ra", "$sp", 4)
		// Reset the stack to the original position.
		c.emitAddUnsigned("$sp", "$sp", 4*numVars+12)
		// Load the old frame pointer.
		c.emitLoadWord("$fp", "$sp", 0)
		c.emitJumpReturn()
	}
}

// generateStatement begins generation of a statement node. It generates assignments, procedure
// calls, if thens, while dos, and print statements.
func (c *CodeGenerator) generateStatement(node *ast.Node, syms []*symtable.SymbolTable) {
	switch node.Tag {
	case ast.Assignment:
		iden := node.Children[0]
		key := symtable.Key{symtable.Integer, iden.Tok.Lex}
		// Find out where/how far up the left hand side was defined.
		n, value := c.getValueFromClosestSymbolTable(key, syms)

		// Rest of stuff.
		c.generateExpression(node.Children[1], syms)
		c.emitAddUnsigned("$sp", "$sp", 4)
		c.emitLoadWord("$a0", "$sp", 0) // Load result onto $a0
		// Indicates which variable on the frame corresponds to the left hand side.
		c.loadAddressOfPreviousRecord("$t0", n, value.Order)
		c.emitStoreWord("$a0", "$t0", 0)
	case ast.Call:
		iden := node.Children[0]
		key := symtable.Key{symtable.Procedure, iden.Tok.Lex}
		n, value := c.getValueFromClosestSymbolTable(key, syms)

		label := value.Label
		numVars := value.NumVars
		// Store the old frame pointer on the stack..
		c.emitStoreWord("$fp", "$sp", 0)
		c.emitSubUnsigned("$sp", "$sp", 4)
		// Calculate the static link.
		c.emitMove("$a0", "$fp") // Points to frame of main if we're at depth 0.
		for i := 0; i < n; i++ {
			c.emitLoadWord("$a0", "$a0", 4)
		}
		// Store the static link on the stack.
		c.emitStoreWord("$a0", "$sp", 0)
		c.emitSubUnsigned("$sp", "$sp", 4)
		// Have the new frame pointer point to the stack.
		c.emitMove("$fp", "$sp")
		// Load all the variables in this scope onto the current frame. Initialize to 0.
		for i := 0; i < numVars; i++ {
			c.emitLoadInt("$a0", 0)
			c.emitStoreWord("$a0", "$sp", 0)
			c.emitSubUnsigned("$sp", "$sp", 4)
		}
		c.emitJumpAndLink(label)
	case ast.Begin:
		// Generate any statements under the begin. Retains same lexical scope.
		for _, node := range node.Children {
			c.generateStatement(node, syms)
		}
	case ast.IfThen:
		cond := node.Children[0]
		stmt := node.Children[1]
		label := c.getNewLabel("if")
		doneLabel := label + "_done"

		c.generateCondition(cond, label, syms)
		// Jump to done if condition evaluates to false.
		c.emitJump(doneLabel)
		c.emitLabel(label)
		c.generateStatement(stmt, syms)
		c.emitLabel(doneLabel)
	case ast.WhileDo:
		cond := node.Children[0]
		stmt := node.Children[1]
		label := c.getNewLabel("while")
		doLabel := label + "_do"
		doneLabel := c.getNewLabel("done")
		c.emitLabel(label)
		c.generateCondition(cond, doLabel, syms)
		// Jump to done if condition evaluates to false.
		c.emitJump(doneLabel)
		c.emitLabel(doLabel)
		c.generateStatement(stmt, syms)
		// Jump to the beginning of the while loop.
		c.emitJump(label)
		c.emitLabel(doneLabel)
	case ast.Print:
		expr := node.Children[0]
		c.generateExpression(expr, syms)
		// Pop result off of the stack.
		c.emitAddUnsigned("$sp", "$sp", 4)
		c.emitLoadWord("$a0", "$sp", 0)
		// Emit syscalls to print an integer and a newline.
		c.emitLoadInt("$v0", 1)
		c.emitSyscall()
		c.emitLoadInt("$a0", 10) // Prints newline character.
		c.emitLoadInt("$v0", 11)
		c.emitSyscall()
	default:
		// This can't possibly happen...
		fmt.Println("A terrible error occurred.",
			"The abstract syntax tree is wrong and I'm generating code...")
	}
}

// generateConditiont begins generation of a condition node. It evaluates the two expressions on
// either side of the condition and compares them with the appropriate branch command. If the
// condition returns true, then the code resumes at the specified label. Otherwise, it continues at
// the next instruction.
func (c *CodeGenerator) generateCondition(node *ast.Node, label string,
	syms []*symtable.SymbolTable) {
	if node.Tag == ast.Odd {
		c.generateExpression(node.Children[0], syms)
		// Pop result off of the stack and check if it's odd.
		c.emitAddUnsigned("$sp", "$sp", 4)
		c.emitLoadWord("$t0", "$sp", 0)
		c.emitAndImmediate("$t0", "$t0", 1)
		c.emitBranchOnGreaterThanZero("$t0", label)
		return
	}
	c.generateExpression(node.Children[0], syms)
	c.generateExpression(node.Children[1], syms)
	// Pop the results off the stack and compare them.
	c.emitAddUnsigned("$sp", "$sp", 4)
	c.emitLoadWord("$t0", "$sp", 0)
	c.emitAddUnsigned("$sp", "$sp", 4)
	c.emitLoadWord("$t1", "$sp", 0)
	switch node.Op {
	case token.Equals:
		c.emitBranchOnEqual("$t0", "$t1", label)
	case token.NotEquals:
		c.emitBranchNotEqual("$t0", "$t1", label)
	case token.LessThan:
		c.emitSub("$t0", "$t0", "$t1")
		c.emitBranchOnGreaterThanZero("$t0", label)
	case token.GreaterThan:
		c.emitSub("$t0", "$t1", "$t0")
		c.emitBranchOnGreaterThanZero("$t1", label)
	case token.LessThanEqualTo:
		c.emitSub("$t0", "$t0", "$t1")
		c.emitBranchOnGreaterThanOrEqualZero("$t0", label)
	case token.GreaterThanEqualTo:
		c.emitSub("$t0", "$t1", "$t0")
		c.emitBranchOnGreaterThanOrEqualZero("$t0", label)
	default:
		// This can't possibly happen...
		fmt.Println("A terrible error occurred.",
			"The abstract syntax tree is wrong and I'm generating code...")
	}
}

// generateExpression begins generation of an expression node. It evaluates an expression and places
// the result on the stack.
func (c *CodeGenerator) generateExpression(node *ast.Node, syms []*symtable.SymbolTable) {
	if node.Tag == ast.Terminal {
		// Only look through the symbol table if it's an idenfitier!
		if node.Tok.Tag == token.Identifier {
			key := symtable.Key{symtable.Integer, node.Tok.Lex}
			n, value := c.getValueFromClosestSymbolTable(key, syms)
			// If the value is nil, then the identifier must be a constant.
			if value == nil {
				key := symtable.Key{symtable.Constant, node.Tok.Lex}
				_, value := c.getValueFromClosestSymbolTable(key, syms)
				c.emitLoadInt("$a0", value.Val)
				c.emitStoreWord("$a0", "$sp", 0)
				c.emitSubUnsigned("$sp", "$sp", 4)
				return
			}
			// Load the identifier from the correct activation record.
			c.loadAddressOfPreviousRecord("$a0", n, value.Order)
			c.emitLoadWord("$a0", "$a0", 0)
			c.emitStoreWord("$a0", "$sp", 0)
			c.emitSubUnsigned("$sp", "$sp", 4)
		} else if node.Tok.Tag == token.Integer {
			// If the value is just an integer, then we can go ahead and load it.
			c.emitLoadInt("$a0", node.Tok.Val)
			c.emitStoreWord("$a0", "$sp", 0)
			c.emitSubUnsigned("$sp", "$sp", 4)
		} else {
			// This can't possibly happen...
			fmt.Println("A terrible error occurred.",
				"The abstract syntax tree is wrong and I'm generating code...")
		}
		return
	}
	left := node.Children[0]
	right := node.Children[1]
	c.generateExpression(left, syms)
	c.generateExpression(right, syms)
	// Pop the results off the stack.
	c.emitAddUnsigned("$sp", "$sp", 4)
	c.emitLoadWord("$t0", "$sp", 0)
	c.emitAddUnsigned("$sp", "$sp", 4)
	c.emitLoadWord("$t1", "$sp", 0)
	if node.Op == token.Plus {
		c.emitAdd("$t0", "$t0", "$t1")
	} else if node.Op == token.Minus {
		c.emitSub("$t0", "$t1", "$t0")
	} else if node.Op == token.Times {
		c.emitMul("$t0", "$t1")
		c.emitMoveFromLo("$t0")
	} else if node.Op == token.Divide {
		c.emitDiv("$t1", "$t0")
		c.emitMoveFromLo("$t0")
	} else {
		// This can't possibly happen...
		fmt.Println("A terrible error occurred.",
			"The abstract syntax tree is wrong and I'm generating code...")
	}
	// Store the result on the stack.
	c.emitStoreWord("$t0", "$sp", 0)
	c.emitSubUnsigned("$sp", "$sp", 4)
}

// loadAddressOfPreviousRecord loads the address of the variable n activation records back  at
// position m into register dest.
func (c *CodeGenerator) loadAddressOfPreviousRecord(dest string, n int, m int) {
	c.emitMove(dest, "$fp")
	for i := 0; i < n; i++ { // If we need to go back.
		c.emitLoadWord(dest, dest, 4) // Load old frame pointer.
	}
	// Load var at position m.
	c.emitSubUnsigned(dest, dest, 4*m)
}

// emitAndImmediate emits a andi instruction. $t = $s & imm;
func (c *CodeGenerator) emitAndImmediate(t string, s string, imm int) {
	c.writeOut(fmt.Sprintf("andi %s %s %d\n", t, s, imm))
}

// emitBranchOnGreaterThanZero emits a bgtz instruction. Jumps to l if s is greater than 0.
// if $s > 0 j l;
func (c *CodeGenerator) emitBranchOnGreaterThanZero(s string, l string) {
	c.writeOut(fmt.Sprintf("bgtz %s %s\n", s, l))
}

// emitBranchOnGreaterThanOrEqualZero emits a bgez instruction. Jumps to l if s is greater than or
// equal to 0. if $s >= 0 j l;
func (c *CodeGenerator) emitBranchOnGreaterThanOrEqualZero(s string, l string) {
	c.writeOut(fmt.Sprintf("bgez %s %s\n", s, l))
}

// emitBranchOnEqual emits a beq instruction. Jumps to l if s is equal to t. if $s == $t j l;
func (c *CodeGenerator) emitBranchOnEqual(s string, t string, l string) {
	c.writeOut(fmt.Sprintf("beq %s %s %s\n", s, t, l))
}

// emitBranchNotEqual emits a bne instruction. Jumps to l if s is not equal to t. if $s != $t j l;
func (c *CodeGenerator) emitBranchNotEqual(s string, t string, l string) {
	c.writeOut(fmt.Sprintf("bne %s %s %s\n", s, t, l))
}

// emitJump emits a j instruction to some label. j l;
func (c *CodeGenerator) emitJump(l string) {
	c.writeOut(fmt.Sprintf("j %s\n", l))
}

// emitJumAndLink emits a jal instruction to some label. jal l;
func (c *CodeGenerator) emitJumpAndLink(l string) {
	c.writeOut(fmt.Sprintf("jal %s\n", l))
}

// emitJumpReturn emits a jr instruction to $ra. jr $ra;
func (c *CodeGenerator) emitJumpReturn() {
	// I'm pretty sure we'll only jr to $ra.
	c.writeOut("jr $ra\n")
}

// emitLoadWord emits a lw instruction. $t = MEM[$s + offset];
func (c *CodeGenerator) emitLoadWord(t string, s string, offset int) {
	c.writeOut(fmt.Sprintf("lw %s %d(%s)\n", t, offset, s))
}

// emitLoadInt emits a li instruction. $t = imm
func (c *CodeGenerator) emitLoadInt(t string, imm int) {
	c.writeOut(fmt.Sprintf("li %s %d\n", t, imm))
}

// emitStoreWord emits a sw instruction. MEM[$s + offset] = $t;
func (c *CodeGenerator) emitStoreWord(t string, s string, offset int) {
	c.writeOut(fmt.Sprintf("sw %s %d(%s)\n", t, offset, s))
}

// emitAddUnsigned emits a addu instruction. $d = $s + imm;
func (c *CodeGenerator) emitAddUnsigned(d string, s string, imm int) {
	c.writeOut(fmt.Sprintf("addu %s %s %d\n", d, s, imm))
}

// emitSubUnsigned emits a subu instruction. $d = $s - imm;
func (c *CodeGenerator) emitSubUnsigned(d string, s string, imm int) {
	c.writeOut(fmt.Sprintf("subu %s %s %d\n", d, s, imm))
}

// emitAdd emits an add instruction. $d = $s + $t;
func (c *CodeGenerator) emitAdd(d string, s string, t string) {
	c.writeOut(fmt.Sprintf("add %s %s %s\n", d, s, t))
}

// emitSub emits a sub instruction. $d = $s - $t;
func (c *CodeGenerator) emitSub(d string, s string, t string) {
	c.writeOut(fmt.Sprintf("sub %s %s %s\n", d, s, t))
}

// emitMult emits a mult instruction. $LO = $s * $t;
func (c *CodeGenerator) emitMul(s string, t string) {
	c.writeOut(fmt.Sprintf("mult %s %s\n", s, t))
}

// emitMult emits a div instruction. $LO = $s / $t;
func (c *CodeGenerator) emitDiv(s string, t string) {
	c.writeOut(fmt.Sprintf("div %s %s\n", s, t))
}

// emitMoveFromLo emits a mflo instruction. $d = $LO;
func (c *CodeGenerator) emitMoveFromLo(d string) {
	c.writeOut(fmt.Sprintf("mflo %s\n", d))
}

// emitMove emits a move instruction. $t = $s;
func (c *CodeGenerator) emitMove(t string, s string) {
	c.writeOut(fmt.Sprintf("move %s %s\n", t, s))
}

// getNewLabel returns the specified base label appended with a unique integer.
func (c *CodeGenerator) getNewLabel(base string) string {
	label := fmt.Sprintf("%s%d", base, c.count)
	c.count++
	return label
}

// emitLabel emits the specified label in assembly form.
func (c *CodeGenerator) emitLabel(label string) {
	c.writeOut(label + ":\n")
}

// emitSyscall emits a spim syscall.
func (c *CodeGenerator) emitSyscall() {
	c.writeOut("syscall\n")
}

// getValueFromClosestSymbolTable returns a Value corresponding to the specified Key from the
// closest symbol table which contains it.
func (c *CodeGenerator) getValueFromClosestSymbolTable(key symtable.Key,
	syms []*symtable.SymbolTable) (int, *symtable.Value) {
	for i := len(syms) - 1; i >= 0; i-- {
		value := syms[i].Get(key)
		if value != nil {
			return len(syms) - 1 - i, value
		}
	}
	return 0, nil
}

// writeOut takes a string and prints it to the buffer if not nil. Otherwise it prints to stdout.
func (c *CodeGenerator) writeOut(s string) {
	if c.buf != nil {
		c.buf.WriteString(s)
	} else {
		fmt.Print(s)
	}
}
