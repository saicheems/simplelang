// Package symtable implements the symbol table.
package symtable

const (
	Constant  = iota // ex. CONST a;
	Integer          // ex. VAR a; b := 3 + c;
	Procedure        // ex. CALL myfunc;
)

// EmtpyValue is a Value with all fields initialized to nil.
var EmptyValue *Value = &Value{}

// Key implements a key for the symbol table. Should be initialized with a tag (const defined by
// this package) and a lexeme.
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

// New returns a new SymbolTable.
func New() *SymbolTable {
	s := new(SymbolTable)
	s.table = make(map[Key]*Value)
	return s
}

// Put adds the specified key and value to the symbol table. An entry will be overriden if the key
// is not unique.
func (s *SymbolTable) Put(key Key, value *Value) {
	s.table[key] = value
}

// Get returns a *Value corresponding to the specified Key or nil if there is no entry corresponding
// to that key.
func (s *SymbolTable) Get(key Key) *Value {
	return s.table[key]
}
