// Package symtable implements the symbol table.
package symtable

const (
	Constant  = iota // ex. CONST a;
	Integer          // ex. VAR a; b := 3 + c;
	Procedure        // ex. CALL myfunc;
)

// EmtpyValue is a Value with all fields initialized to nil.
var EmptyValue *Value = &Value{}

// Key implements a key for the symbol table. Should be initialized with a tag defined by this
// package and a lexeme.
type Key struct {
	Tag int    // One of Constant, Integer, or Procedure.
	Lex string // Lexeme of Token.
}

// Value contains information needed by the code generation phase.
type Value struct {
	Label string // Assembly label of function for code generation purposes.
	Order int    // The position in the stack frame of the variable (nth VAR).
	Val   int    // For constants.
}

// SymbolTable implements a symbol table as a map with key Key and value *Value.
type SymbolTable struct {
	table map[Key]*Value
}

// New returns a new SymbolTable whose table has been initialized.
func New() *SymbolTable {
	s := new(SymbolTable)
	s.table = make(map[Key]*Value)
	return s
}

// Put adds a symbol to the symbol table. Its arguments are a (key) Symbol and a Value.
func (st *SymbolTable) Put(s Key, v *Value) {
	st.table[s] = v
}

// Get returns a boolean that represents whether or not a symbol is present in the symbol table. It
// takes in a symbol tag (one of Constant, Integer, or Procedure) and a lexeme.
func (st *SymbolTable) Get(s Key) *Value {
	return st.table[s]
}
