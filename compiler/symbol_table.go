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
	LocalScope   SymbolScope = "LOCAL"
	GlobalScope  SymbolScope = "GLOBAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope    SymbolScope = "FREE"
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
	Outer          *SymbolTable
	store          map[string]Symbol
	numDefinitions int
	FreeSymbols    []Symbol
}

// Crea una tabla de símbolos
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer

	return s
}

// define un símbolo
func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Index: s.numDefinitions,
	}
	// Si tenemos un parent entonces todos
	// nuestros binding son locales.
	if s.Outer != nil {
		symbol.Scope = LocalScope
	} else {
		symbol.Scope = GlobalScope
	}

	s.store[name] = symbol
	s.numDefinitions += 1

	return symbol
}

// define un builtin
func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Index: index,
		Scope: BuiltinScope,
	}
	s.store[name] = symbol

	return symbol
}

// define free
func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)

	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1}
	symbol.Scope = FreeScope
	s.store[original.Name] = symbol

	return symbol
}

// resuelve un símbolo
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}
		if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
			return obj, ok
		}
		free := s.defineFree(obj)

		return free, true
	}
	return obj, ok
}
