package code

type OpCode byte

const (
	OpConstant OpCode = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpNull
	OpLess
	OpLessEq
	OpGreater
	OpGreaterEq
	OpEqual
	OpNotEq
	OpNegInt
	OpNegBool
	OpAnd
	OpOr
	OpJumpNotTrue
	OpJump
	OpSetGlobal
	OpGetGlobal
	OpPop // le indica a la vm que limpie la pila
)

// nemónicos para emitir el log con las instrucciones
// solo sirve para depuración
var Nemmonics = map[OpCode]string{
	OpConstant:    "PUSH",
	OpAdd:         "ADD",
	OpSub:         "SUB",
	OpMul:         "MUL",
	OpDiv:         "DIV",
	OpTrue:        "PUSH true",
	OpFalse:       "PUSH false",
	OpNull:        "PUSH null",
	OpLess:        "LESS",
	OpLessEq:      "LESS_EQ",
	OpGreater:     "GREATER",
	OpGreaterEq:   "GREATER_EQ",
	OpEqual:       "EQUAL",
	OpNotEq:       "NOT_EQ",
	OpNegInt:      "NEG_INT",
	OpNegBool:     "NEG_BOOL",
	OpAnd:         "AND",
	OpOr:          "OR",
	OpJumpNotTrue: "JUMP_NOT_TRUE",
	OpJump:        "JUMP",
	OpSetGlobal:   "SET",
	OpGetGlobal:   "GET",
	OpPop:         "POP",
}
