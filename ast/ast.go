package ast

import "MonkeyHabilis/token"

type Type int

// El tipo de ast
const (
	// contenedores
	PROGRAM = iota
	BLOCK
	EXPRESSION

	// tipos nativos
	INT
	IDENT
	STRING
	BOOLEAN
	NULL

	// computadores
	BINARY
	UNARY

	// control de flujo
	IF
	FUNCTION
	LET
	RETURN
	ARRAY
	HASH
	CALL
)

type Node interface {
	Type() Type
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// contenedores de sentencias
type ProgramNode struct {
	Statements []Statement
}

func (p *ProgramNode) Type() Type {
	return PROGRAM
}

type BlockStmtNode struct {
	Statements []Statement
}

func (b *BlockStmtNode) Type() Type {
	return BLOCK
}

func (b *BlockStmtNode) statementNode() {}

// Sentencias
type LetStmtNode struct {
	Name  IdentifierNode
	Value Expression
}

func (ls *LetStmtNode) statementNode() {}
func (ls *LetStmtNode) Type() Type {
	return LET
}

type ReturnStmtNode struct {
	Value Expression
}

func (rs *ReturnStmtNode) statementNode() {}
func (rs *ReturnStmtNode) Type() Type {
	return RETURN
}

type ExpressionStmtNode struct {
	Expression Expression
}

func (e *ExpressionStmtNode) statementNode() {}
func (e *ExpressionStmtNode) Type() Type {
	return EXPRESSION
}

// Computadores
type Binary struct {
	Left  Expression
	Op    token.Token
	Right Expression
}

func (b *Binary) expressionNode() {}
func (b *Binary) Type() Type {
	return BINARY
}

type Unary struct {
	Op    token.Token
	Right Expression
}

func (u *Unary) expressionNode() {}
func (u *Unary) Type() Type {
	return UNARY
}

type CallExprNode struct {
	Callee    Expression
	Arguments []Expression
}

func (cn *CallExprNode) expressionNode() {}
func (cn *CallExprNode) Type() Type {
	return CALL
}

// controladores de flujo
type IfExprNode struct {
	Condition   Expression
	Consequence *BlockStmtNode
	Alternative *BlockStmtNode
}

func (i *IfExprNode) expressionNode() {}
func (i *IfExprNode) Type() Type {
	return IF
}

type FunLiteralNode struct {
	Parameters []IdentifierNode
	Body       *BlockStmtNode
}

func (fl *FunLiteralNode) expressionNode() {}
func (fl *FunLiteralNode) Type() Type {
	return FUNCTION
}

type ArrayLiteralNode struct {
	Elements []Expression
}

func (an *ArrayLiteralNode) expressionNode() {}
func (an *ArrayLiteralNode) Type() Type {
	return ARRAY
}

type HashLiteralNode struct {
	Pairs map[Expression]Expression
}

func (hn *HashLiteralNode) expressionNode() {}
func (hn *HashLiteralNode) Type() Type {
	return HASH
}

// tipos nativos
type IdentifierNode struct {
	Value string
}

func (id *IdentifierNode) expressionNode() {}
func (id *IdentifierNode) Type() Type {
	return IDENT
}

type IntegerNode struct {
	Value int
}

func (in *IntegerNode) expressionNode() {}
func (in *IntegerNode) Type() Type {
	return INT
}

type StringNode struct {
	Value string
}

type BooleanNode struct {
	Value bool
}

func (bn *BooleanNode) expressionNode() {}
func (bn *BooleanNode) Type() Type {
	return BOOLEAN
}

func (sn *StringNode) expressionNode() {}
func (sn *StringNode) Type() Type {
	return STRING
}

type NullNode struct {
	// nada, es null
}

func (nn *NullNode) expressionNode() {}
func (nn *NullNode) Type() Type {
	return NULL
}
