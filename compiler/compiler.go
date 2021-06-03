package compiler

import (
	"MonkeyHabilis/ast"
	"MonkeyHabilis/code"
	"MonkeyHabilis/object"
	"MonkeyHabilis/token"
	"bytes"
	"fmt"
)

// Este objeto hace de tupla para la máquina virtual.
type Instruction struct {
	OpCode   code.OpCode
	Position int
}

// representa el bytecode para la vm
type ByteCode struct {
	Instructions []Instruction
	ObjectPool   []object.Object
}

// Compiler se encarga de recorrer el AST y emitir el bytecode
type Compiler struct {
	instructions    []Instruction
	objectPool      []object.Object
	lastInstruction Instruction // La última instrucción emitida

	/**************************INICIO DEBUG************************/
	strByteCode  bytes.Buffer
	strConstants []string
	/**************************FIN DEBUG***************************/
}

// Creamos una instancia del compilador
func New() *Compiler {
	comp := &Compiler{
		instructions:    []Instruction{},
		objectPool:      []object.Object{},
		lastInstruction: Instruction{},

		/**************************INICIO DEBUG************************/
		strByteCode:  bytes.Buffer{},
		strConstants: []string{},
		/**************************FIN DEBUG***************************/
	}
	return comp
}

// compilador
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.ProgramNode:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}

	case *ast.BlockStmtNode:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStmtNode:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		// Como una expresión se representa a sí misma
		// tenemos que enviar un comando Pop para que se limpie la pila
		c.AddInstruction(code.OpPop, 0)

	case *ast.LetStmtNode:
		err := c.Compile(&node.Name)
		if err != nil {
			return err
		}
	case *ast.IdentifierNode:

	case *ast.Binary:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		// Emitimos la instrucción según el tipo de operador binario.
		switch node.Op.Type {
		case token.PLUS:
			c.AddInstruction(code.OpAdd, 0)
		case token.MINUS:
			c.AddInstruction(code.OpSub, 0)
		case token.ASTERISK:
			c.AddInstruction(code.OpMul, 0)
		case token.SLASH:
			c.AddInstruction(code.OpDiv, 0)
		case token.LT:
			c.AddInstruction(code.OpLess, 0)
		case token.LT_EQ:
			c.AddInstruction(code.OpLessEq, 0)
		case token.GT:
			c.AddInstruction(code.OpGreater, 0)
		case token.GT_EQ:
			c.AddInstruction(code.OpGreaterEq, 0)
		case token.EQ:
			c.AddInstruction(code.OpEqual, 0)
		case token.NOT_EQ:
			c.AddInstruction(code.OpNotEq, 0)
		case token.AND:
			c.AddInstruction(code.OpAnd, 0)
		case token.OR:
			c.AddInstruction(code.OpOr, 0)
		default:
			return fmt.Errorf("unknown operator %s", node.Op.Literal)
		}

	case *ast.Unary:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Op.Type {
		case token.MINUS:
			c.AddInstruction(code.OpNegInt, 0)
		case token.BANG:
			c.AddInstruction(code.OpNegBool, 0)
		default:
			return fmt.Errorf("unknown operator for unary expression %s", node.Op.Literal)
		}

	case *ast.IntegerNode:
		/*
		* este nodo es un literal entero por lo tanto
		* tenemos que agregarlo en el array de constantes
		* y solo nos interesa su indice para crear el bytecode.
		 */
		index := c.AddConstant(&object.Integer{Value: node.Value})
		/**************************INICIO DEBUG************************/
		c.addConstantInLog(fmt.Sprint(node.Value))
		/**************************FIN DEBUG***************************/
		// este index es el que sirve para armar el bytecode.
		c.AddInstruction(code.OpConstant, index)

	case *ast.StringNode:
		index := c.AddConstant(&object.String{Value: node.Value})
		/**************************INICIO DEBUG************************/
		c.addConstantInLog(fmt.Sprint(node.Value))
		/**************************FIN DEBUG***************************/
		c.AddInstruction(code.OpConstant, index)

	case *ast.BooleanNode:
		opCode := code.OpFalse
		if node.Value {
			opCode = code.OpTrue
		}
		c.AddInstruction(opCode, 0)

	case *ast.NullNode:
		c.AddInstruction(code.OpNull, 0)

	case *ast.IfExprNode:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		// emitimos la instrucción OpJumpNotTrue con un valor falso
		// que luego actualizaremos con el real.
		jumpNotTruePos := c.AddInstruction(code.OpJumpNotTrue, 9999) // 😅 el 9999 es un indice falso. 🤫

		// ahora compilamos la consecuencia
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		c.checkLastPop()

		// emitimos el comando Jump para que salte
		// una vez ejecutado el bloque del if.
		jumpOpPos := c.AddInstruction(code.OpJump, 9999)

		// actualizamos la posición del OpJumpNotTrue
		c.updateOpCodePosition(jumpNotTruePos, len(c.instructions))

		if node.Alternative != nil {
			// ahora compilamos la alternativa
			err = c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			// revisamos si la última instrucción es un OpPop para eliminarlo porque nos dará morcilla luego.
			c.checkLastPop()
		} else {
			// agregar un null por defecto
			c.AddInstruction(code.OpNull, 0)
		}
		// actualizamos la posición del OpJump
		c.updateOpCodePosition(jumpOpPos, len(c.instructions))
	}
	return nil
}

// Agrega un literal u Objeto Constante en el array de objetos
// y devuelve su indice.
func (c *Compiler) AddConstant(obj object.Object) int {
	current_position := len(c.objectPool)
	c.objectPool = append(c.objectPool, obj)

	return current_position
}

// Agrega la instrucción al array de instructiones
func (c *Compiler) AddInstruction(opCode code.OpCode, index int) int {
	// obtenemos el índice actual de la instrucción
	current_position := len(c.instructions)

	// Crea la instrucción y la agrega al array
	instruction := Instruction{OpCode: opCode, Position: index}
	c.instructions = append(c.instructions, instruction)

	// Guardamos la instrucción
	c.setLastInstruction(opCode, current_position)

	return current_position
}

// Guarda la última instrucción
func (c *Compiler) setLastInstruction(opCode code.OpCode, position int) {
	c.lastInstruction = Instruction{OpCode: opCode, Position: position}
}

// Revisamos si la última instrucción emitida es un OpPop
// para eliminarlo de la lista de instrucciones.
// también actualizamos la última instrucción generada.
func (c *Compiler) checkLastPop() {
	if c.lastInstruction.OpCode == code.OpPop {
		// desde el primero hasta el lastInstruction.Position
		c.instructions = c.instructions[:c.lastInstruction.Position]
		// como estamos eliminando lastInstruction, tenemos que pasarlo al anterior.
		c.lastInstruction = c.instructions[len(c.instructions)-1]
	}
}

// Actualiza la posición de un OpCode emitido
func (c *Compiler) updateOpCodePosition(listPosition int, newIndex int) {
	c.instructions[listPosition].Position = newIndex
}

// Genera y devuelve el objeto Bytecode final.
func (c *Compiler) GetByteCode() *ByteCode {
	bytecode := &ByteCode{
		Instructions: c.instructions,
		ObjectPool:   c.objectPool,
	}
	return bytecode
}

/**************************INICIO DEBUG************************/
// Sirve para generar todo el Bytecode en formato legible.
func (c *Compiler) DumpInstructions() string {
	for ip := 0; ip < len(c.instructions); ip++ {
		// extraemos la instrucción actual
		opCode := c.instructions[ip].OpCode
		position := c.instructions[ip].Position

		strCmd := code.Nemmonics[opCode]
		if opCode == code.OpConstant {
			c.strByteCode.WriteString(fmt.Sprintf("%d %s %s\n", ip, strCmd, c.strConstants[position]))
		} else {
			c.strByteCode.WriteString(fmt.Sprintf("%d %s %d\n", ip, strCmd, position))
		}
	}

	return c.strByteCode.String()
}

// Agrega el valor literal de la constante
// para que se pueda leer mejor en el formato final.
func (c *Compiler) addConstantInLog(value string) {
	c.strConstants = append(c.strConstants, value)
}

/**************************FIN DEBUG***************************/
