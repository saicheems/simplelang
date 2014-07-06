package token

type SymbolTable struct {
	prev  *SymbolTable   // Pointer to higher level on tree.
	table map[string]int // Map of words to id's.
}

func NewSymbolTable() *SymbolTable {
	s := new(SymbolTable)
	s.table = make(map[string]int)
	return s
}
