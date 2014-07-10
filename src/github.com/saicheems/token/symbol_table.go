package token

const (
	SymbolConstant  = 0
	SymbolInteger   = 1
	SymbolProcedure = 2
)

type Symbol struct {
	Tag int
	Lex string
}

type SymbolTable struct {
	table map[Symbol]bool // Map of words to tags.
}

func NewSymbolTable() *SymbolTable {
	s := new(SymbolTable)
	s.table = make(map[Symbol]bool)
	return s
}

func (s *SymbolTable) Put(sym Symbol) {
	// Prevent replacement...
	if !s.table[sym] {
		s.table[sym] = true
	}
}

func (s *SymbolTable) Get(lex string, tag int) bool {
	return s.table[Symbol{Tag: tag, Lex: lex}]
}
