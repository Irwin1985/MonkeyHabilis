package token

// Type es el tipo del token
type Type string

// Estructura Token
type Token struct {
	Type    Type
	Literal string
}

// lista de constantes (en otros paquetes se acceden así: `token.EOF`)
const (
	ILLEGAL = "ILLEGAL"

	EOF = "EOF"

	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STRING"
	NULL   = "NULL"

	ASSIGN   = "ASSIGN"
	PLUS     = "PLUS"
	MINUS    = "MINUS"
	BANG     = "BANG"
	ASTERISK = "ASTERISK"
	SLASH    = "SLASH"

	LT     = "LT"
	LT_EQ  = "LT_EQ"
	GT     = "GT"
	GT_EQ  = "GT_EQ"
	EQ     = "EQ"
	NOT_EQ = "NOT_EQ"

	COMMA     = "COMMA"
	SEMICOLON = "SEMICOLON"
	DOT       = "DOT"
	COLON     = "COLON"
	AND       = "AND"
	OR        = "OR"

	LPAREN   = "LPAREN"
	RPAREN   = "RPAREN"
	LBRACE   = "LBRACE"
	RBRACE   = "RBRACE"
	LBRACKET = "LBRACKET"
	RBRACKET = "RBRACKET"

	// keywords
	FUNCTION = "FUN"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	WHILE    = "WHILE"
)

// Diccionario con las palabras reservadas
var Keywords = map[string]Type{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"while":  WHILE,
}

// ¿Este literal es una palabra reservada o un identificador?
func IsKeyword(literal string) Type {
	if tokenType, ok := Keywords[literal]; ok {
		return tokenType
	}
	return IDENT
}
