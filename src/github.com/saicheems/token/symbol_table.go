package token

type SymbolTable struct {
	prev  *SymbolTable   // Pointer to higher level on tree.
	table map[string]int // Map of words to tags.
}

func NewSymbolTable() *SymbolTable {
	s := new(SymbolTable)
	s.table = make(map[string]int)
	return s
}

func (s *SymbolTable) Put(t int, w string) int {
	// Prevent replacement...
	if s.table[w] == 0 {
		s.table[w] = t
	}
	return s.table[w]
}

func (s *SymbolTable) Get(w string) int {
	return s.table[w]
}
