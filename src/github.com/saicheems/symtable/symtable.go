// Package symtable implements the symbol table.
package symtable

const (
	Constant  = iota // ex. CONST a;
	Integer          // ex. VAR a; b := 3 + c;
	Procedure        // ex. CALL myfunc;
)

var EmptyValue *Value = &Value{}

// Symbol implements the key of the symbol table.
type Symbol struct {
	Tag int
	Lex string
}

// Value contains information needed by the code generation phase.
type Value struct {
	Label string
	Order int
	Val   int // For constants.
}

// SymbolTable implements a symbol table as a map of symbols to bools..
type SymbolTable struct {
	table map[Symbol]*Value
}

// New returns a new SymbolTable whose table has been initialized.
func New() *SymbolTable {
	s := new(SymbolTable)
	s.table = make(map[Symbol]*Value)
	return s
}

// Put adds a symbol to the symbol table. Its arguments are a (key) Symbol and a Value.
func (st *SymbolTable) Put(s Symbol, v *Value) {
	st.table[s] = v
}

// Get returns a boolean that represents whether or not a symbol is present in the symbol table. It
// takes in a symbol tag (one of Constant, Integer, or Procedure) and a lexeme.
func (st *SymbolTable) Get(s Symbol) *Value {
	return st.table[s]
}
