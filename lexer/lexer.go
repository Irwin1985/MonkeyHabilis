package lexer

import (
	"MonkeyHabilis/token"
	"fmt"
)

// primero creamos el objeto `Lexer`
type Lexer struct {
	input        string
	pos          int
	current_char byte
}

// creamos el metodo new para crear un objeto Lexer
func New(input string) *Lexer {
	var lexer = &Lexer{input: input, pos: 0}
	// apuntamos al primer caracter
	lexer.current_char = lexer.input[lexer.pos]

	return lexer
}

// creamos el método newToken para crear un token
func newToken(tType token.Type, tLiteral string) token.Token {
	return token.Token{Type: tType, Literal: tLiteral}
}

// avanzamos un caracter
func (l *Lexer) advance() {
	l.pos += 1
	if l.pos >= len(l.input) {
		l.current_char = 0
	} else {
		l.current_char = l.input[l.pos]
	}
}

// miramos el siguiente caracter de `input`
func (l *Lexer) peek() byte {
	var peekPos = l.pos + 1
	if peekPos >= len(l.input) {
		return 0
	} else {
		return l.input[peekPos]
	}
}

// determina si el caracter es un espacio en blanco
func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

// determina si el caracter es un digito
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// determina si el caracter es una letra del alfabeto
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

// determina si el caracter es un identificador válido
func isIdentifier(ch byte) bool {
	return ch == '_' || isLetter(ch) || isDigit(ch)
}

// nos saltamos los espacios en blanco
func (l *Lexer) skipWhitespace() {
	for l.current_char != 0 && isSpace(l.current_char) {
		l.advance()
	}
}

// nos saltamos los comentarios de una linea
func (l *Lexer) skipSingleComment() {
	for l.current_char != 0 && l.current_char != '\n' {
		l.advance()
	}
}

// nos saltamos los comentarios multilinea
func (l *Lexer) skipMultiComment() {
	for l.current_char != 0 {
		if l.current_char == '*' && l.peek() == '/' {
			break
		}
		l.advance()
	}
	l.advance() // avanza el asterisco '*'
	l.advance() // avanza el slash '/'
}

// detectamos un número entero y retornamos un Token
func (l *Lexer) getNumber() token.Token {
	var startPos = l.pos
	for l.current_char != 0 && isDigit(l.current_char) {
		l.advance()
	}
	return newToken(token.INT, l.input[startPos:l.pos])
}

// detectamos un string y retornamos un Token
func (l *Lexer) getString(strDelim byte) token.Token {
	var lexeme string
	l.advance() // avanza el delimitador inicial
	for l.current_char != 0 && l.current_char != strDelim {
		lexeme += string(l.current_char)
		// TODO: quieres dar soporte a escape de caracteres? implementalo aquí...
		l.advance()
	}
	l.advance() // avanza el delimitador final
	return newToken(token.STRING, lexeme)
}

// detectamos un identificador o palabra reservada
func (l *Lexer) getIdentifier() token.Token {
	var startPos = l.pos
	for l.current_char != 0 && isIdentifier(l.current_char) {
		l.advance()
	}
	var lexeme = l.input[startPos:l.pos]

	return newToken(token.IsKeyword(lexeme), lexeme)
}

// generamos un Token
func (l *Lexer) NextToken() token.Token {
	for l.current_char != 0 {
		if isSpace(l.current_char) {
			l.skipWhitespace()
			continue
		}
		if l.current_char == '/' && l.peek() == '/' {
			l.skipSingleComment()
			continue
		}
		if l.current_char == '/' && l.peek() == '*' {
			l.skipMultiComment()
			continue
		}
		if isDigit(l.current_char) {
			return l.getNumber()
		}
		if l.current_char == '\'' || l.current_char == '"' {
			return l.getString(l.current_char)
		}
		if isIdentifier(l.current_char) {
			return l.getIdentifier()
		}
		// caracteres especiales sencillos
		if l.current_char == '+' {
			l.advance()
			return newToken(token.PLUS, "+")
		}
		if l.current_char == '-' {
			l.advance()
			return newToken(token.MINUS, "-")
		}
		if l.current_char == '*' {
			l.advance()
			return newToken(token.ASTERISK, "*")
		}
		if l.current_char == '/' {
			l.advance()
			return newToken(token.SLASH, "/")
		}
		if l.current_char == ',' {
			l.advance()
			return newToken(token.COMMA, ",")
		}
		if l.current_char == ';' {
			l.advance()
			return newToken(token.SEMICOLON, ";")
		}
		if l.current_char == '.' {
			l.advance()
			return newToken(token.DOT, ".")
		}
		if l.current_char == ':' {
			l.advance()
			return newToken(token.COLON, ":")
		}
		if l.current_char == '(' {
			l.advance()
			return newToken(token.LPAREN, "(")
		}
		if l.current_char == ')' {
			l.advance()
			return newToken(token.RPAREN, ")")
		}
		if l.current_char == '{' {
			l.advance()
			return newToken(token.LBRACE, "{")
		}
		if l.current_char == '}' {
			l.advance()
			return newToken(token.RBRACE, "}")
		}
		if l.current_char == '[' {
			l.advance()
			return newToken(token.LBRACKET, "[")
		}
		if l.current_char == ']' {
			l.advance()
			return newToken(token.RBRACKET, "]")
		}
		// caracteres especiales compuestos
		if l.current_char == '<' {
			l.advance()
			if l.current_char == '=' {
				l.advance()
				return newToken(token.LT_EQ, "<=")
			}
			return newToken(token.LT, "<")
		}
		if l.current_char == '>' {
			l.advance()
			if l.current_char == '=' {
				l.advance()
				return newToken(token.GT_EQ, ">=")
			}
			return newToken(token.GT, ">")
		}
		if l.current_char == '!' {
			l.advance()
			if l.current_char == '=' {
				l.advance()
				return newToken(token.NOT_EQ, "!=")
			}
			return newToken(token.BANG, "!")
		}
		if l.current_char == '=' {
			l.advance()
			if l.current_char == '=' {
				l.advance()
				return newToken(token.EQ, "==")
			}
			return newToken(token.ASSIGN, "=")
		}
		if l.current_char == '&' && l.peek() == '&' {
			l.advance()
			l.advance()
			return newToken(token.AND, "&&")
		}
		if l.current_char == '|' && l.peek() == '|' {
			l.advance()
			l.advance()
			return newToken(token.OR, "||")
		}
		fmt.Printf("unknown character: %c\n", l.current_char)
	}
	return newToken(token.EOF, "")
}
