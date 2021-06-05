package compiler

// Alias para el tipo de alcance
// el valor no es muy importante
// siempre y cuando sea único
// porque necesitaremos diferenciar
// entre distintos ámbitos.
// es mejor usar int pero string nos
// ayuda para el debug.
type SymbolScope string

// Lista de ámbitos
const (
	GlobalScope SymbolScope = "GLOBAL"
)

// contiene la información necesaria acerca
// del símbolo que el compilador se encuentre.
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// asocia keys con símbolos y mantiene rastro de
// su definición numérica. Las keys son realmente
// los nombres literales del identificador.
type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

// Crea una tabla de símbolos
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

// define un símbolo
func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Index: s.numDefinitions,
		Scope: GlobalScope,
	}
	s.store[name] = symbol
	s.numDefinitions += 1

	return symbol
}

// resuelve un símbolo
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	return obj, ok
}
