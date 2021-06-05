package code

type OpCode byte

// Este objeto hace de tupla para la máquina virtual.
type Instruction struct {
	OpCode   OpCode
	Position int
	Literal  string // el literal de las constantes
	Id       int    // número de instrucción
}

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
	OpArray
	OpHash
	OpAccess
	OpCall
	OpReturnValue // retorna el objeto de la pila
	OpReturn      // retorna desde la función actual
	OpPop         // le indica a la vm que limpie la pila
)

// nemónicos para emitir el log con las instrucciones
// solo sirve para depuración
var nemmonics = map[OpCode]string{
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
	OpSetGlobal:   "STORE",
	OpGetGlobal:   "FETCH",
	OpArray:       "ARRAY OF",
	OpHash:        "HASH OF",
	OpAccess:      "ACCESS",
	OpCall:        "CALL",
	OpReturnValue: "RETURN_VALUE",
	OpReturn:      "RETURN",
	OpPop:         "POP",
}

// devuelve el OpCode en String
func OpCodeToString(opCode OpCode) string {
	return nemmonics[opCode]
}
