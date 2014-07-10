// Package symtable implements the symbol table.
package symtable

const (
	Constant  = iota // ex. CONST a;
	Integer          // ex. VAR a; b := 3 + c;
	Procedure        // ex. CALL myfunc;
)

// Symbols are the keys of the symbol table.
type symbol struct {
	tag int
	lex string
}

// SymbolTable implements a symbol table as a map of symbols to bools..
type SymbolTable struct {
	table map[symbol]bool
}

// New returns a new SymbolTable whose table has been initialized.
func New() *SymbolTable {
	s := new(SymbolTable)
	s.table = make(map[symbol]bool)
	return s
}

// Put adds a symbol to the symbol table. Its arguments are a symbol tag (one of Constant, Integer,
// or Procedure) and a lexeme.
func (s *SymbolTable) Put(tag int, lex string) {
	key := symbol{tag, lex}
	// Prevent replacement...
	if !s.table[key] {
		s.table[key] = true
	}
}

// Get returns a boolean that represents whether or not a symbol is present in the symbol table. It
// takes in a symbol tag (one of Constant, Integer, or Procedure) and a lexeme.
func (s *SymbolTable) Get(tag int, lex string) bool {
	return s.table[symbol{tag, lex}]
}
