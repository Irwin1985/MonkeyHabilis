package ast

import (
	"MonkeyHabilis/token"
	"bytes"
	"fmt"
	"strings"
)

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
	String() string
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

func (p *ProgramNode) Type() Type { return PROGRAM }
func (p *ProgramNode) String() string {
	var out bytes.Buffer

	if len(p.Statements) > 0 {
		for _, stmt := range p.Statements {
			out.WriteString(stmt.String())
			out.WriteString(";\n")
		}
	}
	return out.String()
}

type BlockStmtNode struct {
	Statements []Statement
}

func (b *BlockStmtNode) statementNode() {}
func (b *BlockStmtNode) Type() Type     { return BLOCK }
func (b *BlockStmtNode) String() string {
	var out bytes.Buffer
	out.WriteString("{\n")
	if len(b.Statements) > 0 {
		for _, stmt := range b.Statements {
			out.WriteString(fmt.Sprintf("%c%s", '\t', stmt.String()))
			out.WriteString(";\n")
		}
	}
	out.WriteString("}")
	return out.String()
}

// Sentencias
type LetStmtNode struct {
	Name  IdentifierNode
	Value Expression
}

func (ls *LetStmtNode) statementNode() {}
func (ls *LetStmtNode) Type() Type     { return LET }
func (ls *LetStmtNode) String() string {
	return fmt.Sprintf("let %s = %s", ls.Name.String(), ls.Value.String())
}

type ReturnStmtNode struct {
	Value Expression
}

func (rs *ReturnStmtNode) statementNode() {}
func (rs *ReturnStmtNode) Type() Type     { return RETURN }
func (rs *ReturnStmtNode) String() string {
	return fmt.Sprintf("return %s", rs.Value.String())
}

type ExpressionStmtNode struct {
	Expression Expression
}

func (e *ExpressionStmtNode) statementNode() {}
func (e *ExpressionStmtNode) Type() Type     { return EXPRESSION }
func (e *ExpressionStmtNode) String() string {
	return e.Expression.String()
}

// Computadores
type Binary struct {
	Left  Expression
	Op    token.Token
	Right Expression
}

func (b *Binary) expressionNode() {}
func (b *Binary) Type() Type      { return BINARY }
func (b *Binary) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left, b.Op.Literal, b.Right)
}

type Unary struct {
	Op    token.Token
	Right Expression
}

func (u *Unary) expressionNode() {}
func (u *Unary) Type() Type      { return UNARY }
func (u *Unary) String() string {
	return fmt.Sprintf("(%s %s)", u.Op.Literal, u.Right.String())
}

type CallExprNode struct {
	Callee    Expression
	Arguments []Expression
}

func (cn *CallExprNode) expressionNode() {}
func (cn *CallExprNode) Type() Type      { return CALL }
func (cn *CallExprNode) String() string {
	var out bytes.Buffer
	out.WriteString(cn.Callee.String())
	out.WriteString("(")
	if len(cn.Arguments) > 0 {
		var arguments = []string{}
		for _, argument := range cn.Arguments {
			arguments = append(arguments, argument.String())
		}
		out.WriteString(strings.Join(arguments, ","))
	}
	out.WriteString(")")
	return out.String()
}

// controladores de flujo
type IfExprNode struct {
	Condition   Expression
	Consequence *BlockStmtNode
	Alternative *BlockStmtNode
}

func (i *IfExprNode) expressionNode() {}
func (i *IfExprNode) Type() Type      { return IF }
func (i *IfExprNode) String() string {
	var out bytes.Buffer

	out.WriteString("if (")
	out.WriteString(i.Condition.String())
	out.WriteString(")")

	// consecuencia
	out.WriteString(i.Consequence.String())

	// alternativa
	if i.Alternative != nil {
		out.WriteString("else")
		out.WriteString(i.Alternative.String())
	}

	return out.String()
}

type FunLiteralNode struct {
	Parameters []IdentifierNode
	Body       *BlockStmtNode
}

func (fl *FunLiteralNode) expressionNode() {}
func (fl *FunLiteralNode) Type() Type      { return FUNCTION }
func (fl *FunLiteralNode) String() string {
	var out bytes.Buffer

	out.WriteString("fn(")

	// lista de parametros
	if len(fl.Parameters) > 0 {
		var parameters = []string{}
		for _, param := range fl.Parameters {
			parameters = append(parameters, param.String())
		}
		out.WriteString(strings.Join(parameters, ","))
	}
	out.WriteString(")")

	// cuerpo de la función
	out.WriteString(fl.Body.String())

	return out.String()
}

type ArrayLiteralNode struct {
	Elements []Expression
}

func (an *ArrayLiteralNode) expressionNode() {}
func (an *ArrayLiteralNode) Type() Type      { return ARRAY }
func (an *ArrayLiteralNode) String() string {
	var out bytes.Buffer
	out.WriteString("[")

	if len(an.Elements) > 0 {
		var elements = []string{}
		for _, element := range an.Elements {
			elements = append(elements, element.String())
		}
		out.WriteString(strings.Join(elements, ","))
	}
	out.WriteString("]")
	return out.String()
}

type HashLiteralNode struct {
	Pairs map[Expression]Expression
}

func (hn *HashLiteralNode) expressionNode() {}
func (hn *HashLiteralNode) Type() Type      { return HASH }
func (hn *HashLiteralNode) String() string {
	var out bytes.Buffer
	out.WriteString("{")
	// imprimir las claves y los valores del diccionario
	if len(hn.Pairs) > 0 {
		var pairs = []string{}
		for key, value := range hn.Pairs {
			pairs = append(pairs, fmt.Sprintf("%s: %s", key, value))
		}
		out.WriteString(strings.Join(pairs, ","))
	}
	out.WriteString("}")
	return out.String()
}

// tipos nativos
type IdentifierNode struct {
	Value string
}

func (id *IdentifierNode) expressionNode() {}
func (id *IdentifierNode) Type() Type      { return IDENT }
func (id *IdentifierNode) String() string {
	return id.Value
}

type IntegerNode struct {
	Value int64
}

func (in *IntegerNode) expressionNode() {}
func (in *IntegerNode) Type() Type      { return INT }
func (in *IntegerNode) String() string {
	return fmt.Sprintf("%d", in.Value)
}

type StringNode struct {
	Value string
}

func (sn *StringNode) expressionNode() {}
func (sn *StringNode) Type() Type      { return STRING }
func (sn *StringNode) String() string {
	return string("\"" + sn.Value + "\"")
}

type BooleanNode struct {
	Value bool
}

func (bn *BooleanNode) expressionNode() {}
func (bn *BooleanNode) Type() Type      { return BOOLEAN }
func (bn *BooleanNode) String() string {
	if bn.Value {
		return "true"
	} else {
		return "false"
	}
}

type NullNode struct {
	// nada, es null
}

func (nn *NullNode) expressionNode() {}
func (nn *NullNode) Type() Type      { return NULL }
func (nn *NullNode) String() string  { return "null" }
