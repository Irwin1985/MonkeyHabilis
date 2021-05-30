package parser

import (
	"MonkeyHabilis/ast"
	"MonkeyHabilis/lexer"
	"MonkeyHabilis/token"
	"fmt"
	"strconv"
)

type Parser struct {
	Lexer    *lexer.Lexer
	curToken token.Token
	pekToken token.Token
	Errors   []string // lista de errores encontrados
}

func New(lexer *lexer.Lexer) *Parser {
	var parser = &Parser{Lexer: lexer}
	parser.Errors = []string{}
	parser.nextToken()
	parser.nextToken()
	return parser
}

// avanza el siguiente token
func (p *Parser) nextToken() {
	p.curToken = p.pekToken
	p.pekToken = p.Lexer.NextToken()
}

// compara el token actual y avanza
func (p *Parser) advance(tType token.Type) {
	if tType == p.curToken.Type {
		p.nextToken()
	} else {
		msg := fmt.Sprintf("Couldn't match the token: %s because %s was found.\n", tType, p.curToken.Literal)
		p.Errors = append(p.Errors, msg)
		// TODO: estabilizar el parser aquí...
	}
}

// skipSemicolon
func (p *Parser) skipSemicolon() {
	if p.curToken.Type == token.SEMICOLON {
		p.advance(token.SEMICOLON)
	}
}

// El punto de entrada del parser
// program ::= ( statement )*
func (p *Parser) Program() *ast.ProgramNode {
	var programNode = &ast.ProgramNode{}
	programNode.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		resultExpr := p.statement()
		p.skipSemicolon()
		if resultExpr != nil {
			programNode.Statements = append(programNode.Statements, resultExpr)
		}
	}
	return programNode
}

// bloqStmt ::= '{' ( statement )* '}'
func (p *Parser) block() *ast.BlockStmtNode {

	var blockStmt = &ast.BlockStmtNode{}
	blockStmt.Statements = []ast.Statement{}

	p.advance(token.LBRACE)

	for p.curToken.Type != token.RBRACE {
		resultExpr := p.statement()
		p.skipSemicolon()
		if resultExpr != nil {
			blockStmt.Statements = append(blockStmt.Statements, resultExpr)
		}
	}
	p.advance(token.RBRACE)

	return blockStmt
}

// statement ::= ( letStmt | returnStmt | expressionStmt ) ';'
func (p *Parser) statement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.letStmt()
	case token.RETURN:
		return p.returnStmt()
	default:
		return p.expressionStmt()
	}
}

// letStmt ::= 'let' identifier '=' expression
func (p *Parser) letStmt() ast.Statement {
	var letStmt = &ast.LetStmtNode{}

	p.advance(token.LET)
	letStmt.Name = p.identifier()

	p.advance(token.ASSIGN)
	letStmt.Value = p.expression()

	return letStmt
}

// returnStmt ::= 'return' expression ?
func (p *Parser) returnStmt() ast.Statement {
	var returnStmt = &ast.ReturnStmtNode{}

	p.advance(token.RETURN)

	// TODO: verificar si hay algo que retornar.
	returnStmt.Value = p.expression()

	return returnStmt
}

// expressionStmt ::= expression
func (p *Parser) expressionStmt() ast.Statement {
	var expressionStmt = &ast.ExpressionStmtNode{}

	expressionStmt.Expression = p.expression()

	return expressionStmt
}

// expression ::= logicOr
func (p *Parser) expression() ast.Expression {
	return p.logicOr()
}

// logicOr ::= logicAnd ('||' logicAnd)*
func (p *Parser) logicOr() ast.Expression {
	node := p.logicAnd()
	for p.curToken.Type == token.OR {
		tok := p.curToken
		p.advance(token.OR)
		node = &ast.Binary{Left: node, Op: tok, Right: p.logicAnd()}
	}
	return node
}

// logicAnd ::= equality ('&&' equality)*
func (p *Parser) logicAnd() ast.Expression {
	node := p.equality()
	for p.curToken.Type == token.AND {
		tok := p.curToken
		p.advance(token.AND)
		node = &ast.Binary{Left: node, Op: tok, Right: p.equality()}
	}
	return node
}

// equality ::= comparison ( ('==' | '!=') comparison )*
func (p *Parser) equality() ast.Expression {
	node := p.comparison()

	for p.curToken.Type == token.EQ || p.curToken.Type == token.NOT_EQ {
		tok := p.curToken
		p.advance(tok.Type)
		node = &ast.Binary{Left: node, Op: tok, Right: p.comparison()}
	}

	return node
}

// comparison ::= term ( ( '<' | '<=' | '>' | '>=' ) term)
func (p *Parser) comparison() ast.Expression {
	node := p.term()

	for p.curToken.Type == token.LT || p.curToken.Type == token.LT_EQ ||
		p.curToken.Type == token.GT || p.curToken.Type == token.GT_EQ {
		tok := p.curToken
		p.advance(tok.Type)
		node = &ast.Binary{Left: node, Op: tok, Right: p.term()}
	}

	return node
}

// term ::= factor ( ('+' | '-') factor )*
func (p *Parser) term() ast.Expression {
	node := p.factor()

	for p.curToken.Type == token.PLUS || p.curToken.Type == token.MINUS {
		tok := p.curToken
		p.advance(tok.Type)
		node = &ast.Binary{Left: node, Op: tok, Right: p.factor()}
	}

	return node
}

// factor ::= unary ( ( '*' | '/' ) unary )*
func (p *Parser) factor() ast.Expression {
	node := p.unary()

	for p.curToken.Type == token.ASTERISK || p.curToken.Type == token.SLASH {
		tok := p.curToken
		p.advance(tok.Type)
		node = &ast.Binary{Left: node, Op: tok, Right: p.unary()}
	}

	return node
}

// unary ::= ( '!' | '-' ) unary | dot
func (p *Parser) unary() ast.Expression {
	for p.curToken.Type == token.MINUS || p.curToken.Type == token.BANG {
		tok := p.curToken
		p.advance(tok.Type)
		return &ast.Unary{Op: tok, Right: p.unary()}
	}

	return p.dot()
}

// dot ::= call ( '.' call )*
func (p *Parser) dot() ast.Expression {
	node := p.call()

	for p.curToken.Type == token.DOT {
		tok := p.curToken
		p.advance(token.DOT)
		node = &ast.Binary{Left: node, Op: tok, Right: p.dot()}
	}

	return node
}

// call ::= primary ( '(' arguments ? ')' )
func (p *Parser) call() ast.Expression {
	node := p.primary()
	for {
		if p.curToken.Type == token.LPAREN || p.curToken.Type == token.LBRACKET {
			node = p.callExpression(node)
		} else {
			break
		}
	}
	return node
}

// primary ::= INTEGER | STRING | IDENT | TRUE | FALSE | NULL | FUNCTION | ARRAY | HASH | IF
func (p *Parser) primary() ast.Expression {
	tok := p.curToken
	switch tok.Type {
	case token.INT:
		p.advance(token.INT)
		value, _ := strconv.ParseInt(tok.Literal, 10, 64)
		return &ast.IntegerNode{Value: value}
	case token.STRING:
		p.advance(token.STRING)
		return &ast.StringNode{Value: tok.Literal}
	case token.IDENT:
		p.advance(token.IDENT)
		return &ast.IdentifierNode{Value: tok.Literal}
	case token.TRUE:
		p.advance(token.TRUE)
		return &ast.BooleanNode{Value: true}
	case token.FALSE:
		p.advance(token.FALSE)
		return &ast.BooleanNode{Value: false}
	case token.NULL:
		p.advance(token.NULL)
		return &ast.NullNode{}
	case token.FUNCTION:
		return p.functionLiteral()
	case token.LBRACKET:
		return p.arrayLiteral()
	case token.LBRACE:
		return p.hashLiteral()
	case token.IF:
		return p.ifExpression()
	case token.LPAREN:
		p.advance(token.LPAREN)
		expr := p.expression()
		p.advance(token.RPAREN)
		return expr
	default:
		msg := fmt.Sprintf("unknown token literal: %s\n", tok.Literal)
		p.Errors = append(p.Errors, msg)
		return nil
	}
}

// callExpression ::= functionCall | arrayCall | hashCall
func (p *Parser) callExpression(callee ast.Expression) ast.Expression {
	var callExpr = &ast.CallExprNode{
		Callee: callee,
	}

	if p.curToken.Type == token.LPAREN {
		p.advance(token.LPAREN)
		if p.curToken.Type != token.RPAREN {
			callExpr.Arguments = p.arguments()
		}
		p.advance(token.RPAREN)
	} else if p.curToken.Type == token.LBRACKET {
		p.advance(token.LBRACKET)
		if p.curToken.Type != token.RBRACKET {
			callExpr.Arguments = p.arguments()
			if len(callExpr.Arguments) > 1 {
				msg := fmt.Sprintf("Invalid subcript reference.\n")
				p.Errors = append(p.Errors, msg)
			}
		}
		p.advance(token.RBRACKET)
	}

	return callExpr
}

// functionLiteral ::= 'fn' '(' parameters ? ')'
func (p *Parser) functionLiteral() ast.Expression {
	var functionNode = &ast.FunLiteralNode{}

	p.advance(token.FUNCTION)

	p.advance(token.LPAREN)
	if p.curToken.Type != token.RPAREN {
		functionNode.Parameters = p.parameters()
	}
	p.advance(token.RPAREN)

	functionNode.Body = p.block()

	return functionNode
}

// arrayLiteral ::= '[' arguments? ']'
func (p *Parser) arrayLiteral() ast.Expression {
	var arrayNode = &ast.ArrayLiteralNode{}

	p.advance(token.LBRACKET)
	if p.curToken.Type != token.RBRACKET {
		arrayNode.Elements = p.arguments()
	}
	p.advance(token.RBRACKET)

	return arrayNode
}

// hashLiteral ::= '{' arguments? '}'
func (p *Parser) hashLiteral() ast.Expression {
	var hashLiteral = &ast.HashLiteralNode{}

	p.advance(token.LBRACE)

	if p.curToken.Type != token.RBRACE {
		hashLiteral.Pairs = p.keyValuePairs()
	}

	p.advance(token.RBRACE)

	return hashLiteral
}

// ifExpression
func (p *Parser) ifExpression() ast.Expression {
	var ifExpr = &ast.IfExprNode{}

	p.advance(token.IF)

	// condition
	p.advance(token.LPAREN)
	ifExpr.Condition = p.expression()
	p.advance(token.RPAREN)

	// consequence
	ifExpr.Consequence = p.block()

	// alternative
	if p.curToken.Type == token.ELSE {
		p.advance(token.ELSE)
		ifExpr.Alternative = p.block()
	}

	return ifExpr
}

// keyValuePairs
func (p *Parser) keyValuePairs() map[ast.Expression]ast.Expression {
	var pairs = make(map[ast.Expression]ast.Expression)

	// parse key -> value
	key := p.expression()
	p.advance(token.COLON)
	value := p.expression()
	pairs[key] = value

	for p.curToken.Type == token.COMMA {
		p.advance(token.COMMA)
		// parse key -> value
		key := p.expression()
		p.advance(token.COLON)
		value := p.expression()
		pairs[key] = value
	}

	return pairs
}

// parameters ::= identifier (',' identifier)*
func (p *Parser) parameters() []ast.IdentifierNode {
	var parameters []ast.IdentifierNode

	parameters = append(parameters, p.identifier())

	for p.curToken.Type == token.COMMA {
		p.advance(token.COMMA)
		parameters = append(parameters, p.identifier())
	}

	return parameters
}

// arguments ::= argument (',' argument)
func (p *Parser) arguments() []ast.Expression {
	var arguments []ast.Expression

	arguments = append(arguments, p.expression())
	for p.curToken.Type == token.COMMA {
		p.advance(token.COMMA)
		arguments = append(arguments, p.expression())
	}

	return arguments
}

// identifier
func (p *Parser) identifier() ast.IdentifierNode {
	tok := p.curToken
	p.advance(token.IDENT)
	return ast.IdentifierNode{Value: tok.Literal}
}
