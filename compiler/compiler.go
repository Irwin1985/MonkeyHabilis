package compiler

import (
	"MonkeyHabilis/ast"
	"MonkeyHabilis/code"
	"MonkeyHabilis/object"
	"MonkeyHabilis/token"
	"bytes"
	"fmt"
)

// representa el bytecode para la vm
type ByteCode struct {
	Instructions []code.Instruction
	ObjectPool   []object.Object
}

// Para gestionar los ámbitos de compilación
type CompiledFrame struct {
	instructions []code.Instruction
	ic           int //contador de instrucciones
}

// Compiler se encarga de recorrer el AST y emitir el bytecode
type Compiler struct {
	objectPool  []object.Object
	symbolTable *SymbolTable
	frames      []CompiledFrame
	frameIndex  int
	curFrame    *CompiledFrame
}

// Creamos una instancia del compilador
func New() *Compiler {
	mainFrame := CompiledFrame{
		instructions: []code.Instruction{},
		ic:           0,
	}
	comp := &Compiler{
		objectPool:  []object.Object{},
		symbolTable: NewSymbolTable(),
		frames:      []CompiledFrame{mainFrame},
		frameIndex:  0,
	}
	comp.curFrame = &comp.frames[comp.frameIndex]
	return comp
}

// Crea un nuevo Compiler pero con SymbolTable
func NewWithState(s *SymbolTable, objectPool []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.objectPool = objectPool

	return compiler
}

// Crea un nuevo ámbito de instrucciones
func (c *Compiler) loadFrame() {
	newFrame := CompiledFrame{
		instructions: []code.Instruction{},
		ic:           0,
	}
	c.frames = append(c.frames, newFrame)
	c.frameIndex += 1
	// actualizamos el currentFrame
	c.curFrame = &newFrame
}

// Sale del ámbito de instrucciones actual
func (c *Compiler) unloadFrame() CompiledFrame {
	// obtenemos el frame a devolver
	deletedFrame := *c.curFrame

	// truncamos el último frame (el guardado)
	c.frames = c.frames[:len(c.frames)-1]
	// eliminamos el frame del contador de frames
	c.frameIndex -= 1
	// actualizamos el currentFrame
	c.curFrame = &c.frames[c.frameIndex]

	return deletedFrame
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
		c.addInstruction(code.OpPop, 0, "")

	case *ast.LetStmtNode:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		// el chiste es generar un símbolo con un índice único.
		symbol := c.symbolTable.Define(node.Name.Value)
		c.addInstruction(code.OpSetGlobal, symbol.Index, node.Name.Value)

	case *ast.IdentifierNode:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.addInstruction(code.OpGetGlobal, symbol.Index, node.Value)

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
			c.addInstruction(code.OpAdd, 0, "")
		case token.MINUS:
			c.addInstruction(code.OpSub, 0, "")
		case token.ASTERISK:
			c.addInstruction(code.OpMul, 0, "")
		case token.SLASH:
			c.addInstruction(code.OpDiv, 0, "")
		case token.LT:
			c.addInstruction(code.OpLess, 0, "")
		case token.LT_EQ:
			c.addInstruction(code.OpLessEq, 0, "")
		case token.GT:
			c.addInstruction(code.OpGreater, 0, "")
		case token.GT_EQ:
			c.addInstruction(code.OpGreaterEq, 0, "")
		case token.EQ:
			c.addInstruction(code.OpEqual, 0, "")
		case token.NOT_EQ:
			c.addInstruction(code.OpNotEq, 0, "")
		case token.AND:
			c.addInstruction(code.OpAnd, 0, "")
		case token.OR:
			c.addInstruction(code.OpOr, 0, "")
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
			c.addInstruction(code.OpNegInt, 0, "")
		case token.BANG:
			c.addInstruction(code.OpNegBool, 0, "")
		default:
			return fmt.Errorf("unknown operator for unary expression %s", node.Op.Literal)
		}

	case *ast.IntegerNode:
		// Creamos el objeto
		intObj := &object.Integer{Value: node.Value}
		// Guardamos el objeto en la pila de constantes
		index := c.addConstant(intObj)
		// este index es el que sirve para armar el bytecode.
		c.addInstruction(code.OpConstant, index, fmt.Sprint(node.Value))

	case *ast.StringNode:
		stringObj := &object.String{Value: node.Value}
		index := c.addConstant(stringObj)
		c.addInstruction(code.OpConstant, index, fmt.Sprintf("\"%s\"", node.Value))

	case *ast.BooleanNode:
		opCode := code.OpFalse
		if node.Value {
			opCode = code.OpTrue
		}
		c.addInstruction(opCode, 0, "")

	case *ast.NullNode:
		c.addInstruction(code.OpNull, 0, "")

	case *ast.FunLiteralNode:
		// entramos en un nuevo ámbito de instrucciones para la función
		c.loadFrame()

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// Revisamos si la última instrucción emitida es un OpPop
		// para cambiarla por un Return y así evitar que la
		// máquina virtual se la cargue.
		if c.lastInstructionIs(code.OpPop) {
			lastIndex := len(c.curFrame.instructions) - 1
			c.curFrame.instructions[lastIndex].OpCode = code.OpReturnValue
		}
		// Si no hay ni una expresión suelta (controlada arriba) ni un OpReturnValue
		// entonces es una funcion que no retorna nada y eso es malo, para arreglarlo
		// creamos a huevo una instrucción OpReturn que retorna null (la vm lo hará.)
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.addInstruction(code.OpReturn, 0, "null")
		}

		// dejamos el ámbito y lo guardamos para la función
		functionFrame := c.unloadFrame()

		// creamos el objeto compiledFunction
		functionObj := &object.CompiledFunction{
			Instructions: functionFrame.instructions,
			StrByteCode:  c.PrintInstructions(functionFrame.instructions),
		}
		index := c.addConstant(functionObj)
		c.addInstruction(code.OpConstant, index, "FUNCTION")

	case *ast.CallExprNode:
		err := c.Compile(node.Callee)
		if err != nil {
			return err
		}
		c.addInstruction(code.OpCall, 0, "")

	case *ast.ArrayLiteralNode:
		// compilamos los elementos del array en modo inverso
		// para que la máquina virtual los agregue en el orden correcto.
		// recuerda que la pila trabaja en modo LIFO
		size := len(node.Elements) - 1
		for i := size; i >= 0; i-- {
			c.Compile(node.Elements[i])
		}
		// emitimos una instucción OpArray cuyo índice es el total de
		// elementos que la vm deberá sacar de la pila.
		c.addInstruction(code.OpArray, size, fmt.Sprintf("%d", size+1))

	case *ast.HashLiteralNode:
		// compilamos los elementos del diccionario.
		for key, value := range node.Pairs {
			// compilamos primero el value
			err := c.Compile(value)
			if err != nil {
				return err
			}
			// luego compilamos el key para que el primer
			// pop() de la vm sea el value y el segundo pop() sea el key.
			err = c.Compile(key)
			if err != nil {
				return err
			}
		}
		// ahora emitimos la instrucción OpHash
		size := len(node.Pairs) - 1
		c.addInstruction(code.OpHash, size, fmt.Sprintf("%d", size+1))

	case *ast.IndexExprNode:

		// compilar el callee
		err := c.Compile(node.Callee)
		if err != nil {
			return err
		}

		// compilamos el índice (númerico o string)
		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.addInstruction(code.OpAccess, 0, "")

	case *ast.IfExprNode:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		// emitimos la instrucción OpJumpNotTrue con un valor falso
		// que luego actualizaremos con el real.
		jumpNotTruePos := c.addInstruction(code.OpJumpNotTrue, 0, "")

		// ahora compilamos la consecuencia
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		// emitimos el comando Jump para que salte
		// una vez ejecutado el bloque del if.
		jumpOpPos := c.addInstruction(code.OpJump, 0, "")

		// actualizamos la posición del OpJumpNotTrue
		//c.updateOpCodePosition(jumpNotTruePos, len(c.getInstructions()))
		c.updateOpCodePosition(jumpNotTruePos, len(c.curFrame.instructions))

		if node.Alternative != nil {
			// ahora compilamos la alternativa
			err = c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			// revisamos si la última instrucción es un OpPop para eliminarlo porque nos dará morcilla luego.
			//c.checkLastPop()
			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		} else {
			// agregar un null por defecto
			c.addInstruction(code.OpNull, 0, "")
		}
		// actualizamos la posición del OpJump
		//c.updateOpCodePosition(jumpOpPos, len(c.getInstructions()))
		c.updateOpCodePosition(jumpOpPos, len(c.curFrame.instructions))

	case *ast.ReturnStmtNode:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		// emitimos el opCode
		c.addInstruction(code.OpReturnValue, 0, "")
	}
	return nil
}

// Agrega un literal u Objeto Constante en el array de objetos
// y devuelve su indice.
func (c *Compiler) addConstant(obj object.Object) int {
	current_position := len(c.objectPool)
	c.objectPool = append(c.objectPool, obj)

	return current_position
}

// Agrega la instrucción al array de instructiones
func (c *Compiler) addInstruction(opCode code.OpCode, index int, literal string) int {
	// Obtenemos el índice actual de la instrucción
	current_position := len(c.curFrame.instructions)

	// Crea la instrucción y la agrega al array
	instruction := code.Instruction{
		OpCode:   opCode,
		Position: index,
		Literal:  literal,
		Id:       c.curFrame.ic,
	}
	c.curFrame.instructions = append(c.curFrame.instructions, instruction)

	// incrementamos el contador de instrucciones
	c.curFrame.ic += 1

	return current_position
}

// Elimina la última instrucción emitida
func (c *Compiler) removeLastPop() {
	endIndex := len(c.curFrame.instructions) - 1
	newInstructions := c.curFrame.instructions[:endIndex]
	c.curFrame.instructions = newInstructions
	c.curFrame.ic -= 1 // descontamos una instrucción
}

func (c *Compiler) lastInstructionIs(op code.OpCode) bool {
	if len(c.curFrame.instructions) == 0 {
		return false
	}
	index := len(c.curFrame.instructions) - 1
	return c.curFrame.instructions[index].OpCode == op
}

// Actualiza la posición de un OpCode emitido
func (c *Compiler) updateOpCodePosition(listPosition int, newIndex int) {
	//instructions := c.getInstructions()
	c.curFrame.instructions[listPosition].Position = newIndex
}

// Genera y devuelve el objeto Bytecode final.
func (c *Compiler) GetByteCode() *ByteCode {
	bytecode := &ByteCode{
		Instructions: c.curFrame.instructions,
		ObjectPool:   c.objectPool,
	}
	return bytecode
}

// Delvuelve el conjunto de instrucciones
func (c *Compiler) GetInstructions() []code.Instruction {
	return c.curFrame.instructions
}

/**************************INICIO DEBUG************************/
func (c *Compiler) PrintInstructions(instructions []code.Instruction) string {
	var out bytes.Buffer

	for _, instruction := range instructions {
		opCode := code.OpCodeToString(instruction.OpCode)
		literal := instruction.Literal
		strIns := fmt.Sprintf("%d %s %d %s", instruction.Id, opCode, instruction.Position, literal)
		out.WriteString(fmt.Sprintf("%s\n", strIns))
	}

	return out.String()
}

/**************************FIN DEBUG***************************/
