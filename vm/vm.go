package vm

import (
	"MonkeyHabilis/code"
	"MonkeyHabilis/compiler"
	"MonkeyHabilis/object"
	"fmt"
)

var STACK_SIZE = 2048

// estamos usando un int genérico por lo tanto este número es falso.
var GLOBAL_SIZE = 65536
var MAX_FRAMES = 1024

var TRUE = &object.Boolean{Value: true}
var FALSE = &object.Boolean{Value: false}
var NULL = &object.Null{}

type VM struct {
	objectPool  []object.Object // la lista de literales constantes
	stack       []object.Object // la pila de objetos
	sp          int             // este es el puntero de la pila que siempre apunta al último objeto
	globals     []object.Object // array de variables globales
	frames      []*Frame        // array de Frames
	framesIndex int
}

// Crea la máquina virtual
func New(bytecode *compiler.ByteCode) *VM {
	// Esta es digamos la función principal
	// la máquina virtual creerá que siempre opera sobre frames
	mainFunction := &object.CompiledFunction{
		Instructions: bytecode.Instructions,
	}

	// Creamos el Frame principal
	mainFrame := NewFrame(mainFunction)
	// Creamos el array de frames
	frames := make([]*Frame, MAX_FRAMES)
	// Y le decimos que la posición 0 es el frame principal.
	frames[0] = mainFrame

	vm := &VM{
		objectPool:  bytecode.ObjectPool,
		stack:       make([]object.Object, STACK_SIZE),
		sp:          0,
		globals:     make([]object.Object, GLOBAL_SIZE),
		frames:      frames,
		framesIndex: 1,
	}

	return vm
}

// Crea la máquina virtual con la tabla de símbolos
func NewWithGlobalsStore(bytecode *compiler.ByteCode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s

	return vm
}

// agrega un nuevo frame
func (vm *VM) loadFrame(newFrame *Frame) {
	// ampliamos el array de frames
	vm.frames[vm.framesIndex] = newFrame
	// incrementamos el contador de frames
	vm.framesIndex += 1
}

// elimina el frame
func (vm *VM) unloadFrame() *Frame {
	// guardamos el frame actual
	deletedFrame := vm.frames[vm.framesIndex]
	// y lo eliminamos de la lista
	vm.framesIndex -= 1

	return deletedFrame
}

// devuelve el frame actual
func (vm *VM) curFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

// Comienza el ciclo Fetch-Decode-Execute
func (vm *VM) Run() error {
	for vm.curFrame().ip < len(vm.curFrame().cf.Instructions)-1 {
		vm.curFrame().ip += 1
		// Obtener la instrucción a ejecutar
		instruction := vm.curFrame().cf.Instructions[vm.curFrame().ip]

		switch instruction.OpCode {
		case code.OpConstant:
			// Agregar una constante a la pila, necesitamos su índice entonces
			// lo tomamos de la lista de objetos y usamos el campo instruction.index
			index := instruction.Position
			obj := vm.objectPool[index]
			err := vm.push(obj)
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv, code.OpLess, code.OpLessEq, code.OpGreater, code.OpGreaterEq, code.OpEqual, code.OpNotEq, code.OpAnd, code.OpOr:
			err := vm.executeBinaryOperation(instruction.OpCode)
			if err != nil {
				return err
			}
		case code.OpTrue, code.OpFalse:
			if instruction.OpCode == code.OpTrue {
				vm.push(TRUE)
			} else {
				vm.push(FALSE)
			}
		case code.OpNegBool, code.OpNegInt:
			err := vm.executeUnaryOperation(instruction.OpCode)
			if err != nil {
				return err
			}
		case code.OpNull:
			vm.push(NULL)

		case code.OpJumpNotTrue:
			// TODO: validar el tipo de dato antes.
			if !vm.pop().(*object.Boolean).Value {
				// saltamos a donde nos indique OpJumpNotTrue
				vm.curFrame().ip = instruction.Position - 1 // le resto 1 para que comience exactamente en el número correcto.
			}
		case code.OpJump:
			// saltamos sin preguntar al índice
			vm.curFrame().ip = instruction.Position - 1

		case code.OpPop:
			vm.pop()

		case code.OpSetGlobal:
			// obtenemos el índice que nos dió el compilador
			globalIndex := instruction.Position
			// y por supuesto lo enlazamos con el último elemento de la pila
			// se supone que la sentencia LET lo ha mandado a meter antes en la pila.
			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			// obtenemos el índice del identificador
			globalIndex := instruction.Position
			// empujamos el objeto en la pila
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		case code.OpArray:
			size := instruction.Position
			arrayObj := &object.Array{
				Elements: []object.Object{},
			}

			if size == 0 {
				arrayObj.Elements = append(arrayObj.Elements, NULL)
			} else {
				for i := 0; i <= size; i++ {
					arrayObj.Elements = append(arrayObj.Elements, vm.pop())
				}
			}
			// agregamos el array
			err := vm.push(arrayObj)
			if err != nil {
				return err
			}

		case code.OpHash:
			size := instruction.Position // doble por que son tuplas (key->value)
			// creamos el objeto hash
			hashObj := &object.Hash{
				Pairs: make(map[string]object.Object),
			}
			if size == 0 {
				vm.push(NULL)
			} else {
				for i := 0; i <= size; i++ {
					key := vm.pop().Inspect()
					value := vm.pop()
					hashObj.Pairs[key] = value
				}
			}
			err := vm.push(hashObj)
			if err != nil {
				return err
			}

		case code.OpAccess:
			// recuperamos el objeto que hace de índice
			objIndex := vm.pop()

			// obtenemos el objeto collection de la pila
			objCollection := vm.pop()
			// hacemos las validaciones
			switch objCollection.Type() {
			case object.ARRAY_OBJ:
				// el objeto que sirve de índice debe ser numérico
				if objIndex.Type() != object.INTEGER_OBJ {
					return fmt.Errorf("invalid subscript data type for array access %s", objIndex.Type())
				}
				// convertimos y enviamos a la pila el elemento del array
				arrayObj := objCollection.(*object.Array)
				index := int(objIndex.(*object.Integer).Value)

				if index < 0 || index > len(arrayObj.Elements) {
					return fmt.Errorf("index out of range")
				}

				vm.push(arrayObj.Elements[index])

			case object.HASH_OBJ:
				// el objeto que sirve de índice debe ser string
				if objIndex.Type() != object.STRING_OBJ {
					return fmt.Errorf("invalid subscript data type for array access %s", objIndex.Type())
				}
				// convertimos y enviamos a la pila el elemento del diccionario
				hashObj := objCollection.(*object.Hash)
				key := objIndex.(*object.String).Value
				if objValue, ok := hashObj.Pairs[key]; ok {
					err := vm.push(objValue)
					if err != nil {
						return err
					}
				} else {
					err := vm.push(NULL)
					if err != nil {
						return err
					}
				}
			case object.STRING_OBJ:
				// el objeto que sirve de índice debe ser numérico
				if objIndex.Type() != object.INTEGER_OBJ {
					return fmt.Errorf("invalid subscript data type for array access %s", objIndex.Type())
				}
				stringObj := objCollection.(*object.String)
				index := int(objIndex.(*object.Integer).Value)

				if index < 0 || index > len(stringObj.Value) {
					return fmt.Errorf("index out of range")
				}
				// nuevo string truncado
				newStrObj := &object.String{Value: string(stringObj.Value[index])}
				vm.push(newStrObj)
			}

		case code.OpCall:
			peekObj := vm.peek()
			callee, ok := peekObj.(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function")
			}
			frame := NewFrame(callee)
			vm.loadFrame(frame)

		case code.OpReturnValue:
			returnValue := vm.pop() // obtiene el valor a retornar

			vm.unloadFrame() // sube un frame
			vm.pop()         // quita la función de la pila

			err := vm.push(returnValue) // sube el valor a retornar por el frame anterior
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Agrega un objeto en la pila
func (vm *VM) push(obj object.Object) error {
	if vm.sp >= STACK_SIZE {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = obj
	vm.sp += 1 // incrementa el puntero
	return nil
}

// Quita un elemento de la pila
func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp -= 1 // decrementamos la pila
	return obj
}

// Devuelve el último elemento de la pila
func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

// Devuelve el último elemento sin removerlo de la pila
func (vm *VM) peek() object.Object {
	return vm.stack[vm.sp-1]
}

/*
* FUNCIONES HELPER PARA LA MÁQUINA VIRTUAL
 */
// Ejecuta una operación binaria
func (vm *VM) executeBinaryOperation(op code.OpCode) error {
	// Sumamos los 2 elementos de la pila y devolvemos su resultado
	var right = vm.pop()
	var left = vm.pop()
	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeBinaryInteger(left, op, right)
	} else if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		if op != code.OpAdd {
			return fmt.Errorf("unsupported operator for binary operation: %s %s", left.Type(), right.Type())
		}
		return vm.executeBinaryString(left, op, right)
	} else if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
		return vm.executeBinaryBoolean(left, op, right)
	}
	return fmt.Errorf("unsupported types for binary operation: %s %s", left.Type(), right.Type())
}

// Ejecuta una operación unaria
func (vm *VM) executeUnaryOperation(op code.OpCode) error {
	obj := vm.pop()
	if op == code.OpNegBool {
		if obj.Type() != object.BOOLEAN_OBJ {
			return fmt.Errorf("invalid type for this operation %s", obj.Type())
		}
		if obj.(*object.Boolean).Value {
			vm.push(FALSE)
		} else {
			vm.push(TRUE)
		}
	} else {
		if obj.Type() != object.INTEGER_OBJ {
			return fmt.Errorf("invalid type for this operation %s", obj.Type())
		}
		vm.push(&object.Integer{Value: obj.(*object.Integer).Value * -1})
	}
	return nil
}

// Ejecuta una operación binaria con enteros
func (vm *VM) executeBinaryInteger(left object.Object, op code.OpCode, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	switch op {
	case code.OpAdd:
		vm.push(&object.Integer{Value: leftValue + rightValue})
	case code.OpSub:
		vm.push(&object.Integer{Value: leftValue - rightValue})
	case code.OpMul:
		vm.push(&object.Integer{Value: leftValue * rightValue})
	case code.OpDiv:
		if rightValue == 0 {
			return fmt.Errorf("division by zero")
		}
		vm.push(&object.Integer{Value: leftValue / rightValue})
	case code.OpLess:
		if leftValue < rightValue {
			vm.push(TRUE)
		} else {
			vm.push(FALSE)
		}
	case code.OpLessEq:
		if leftValue <= rightValue {
			vm.push(TRUE)
		} else {
			vm.push(FALSE)
		}
	case code.OpGreater:
		if leftValue > rightValue {
			vm.push(TRUE)
		} else {
			vm.push(FALSE)
		}
	case code.OpGreaterEq:
		if leftValue >= rightValue {
			vm.push(TRUE)
		} else {
			vm.push(FALSE)
		}
	case code.OpEqual:
		if leftValue == rightValue {
			vm.push(TRUE)
		} else {
			vm.push(FALSE)
		}
	case code.OpNotEq:
		if leftValue != rightValue {
			vm.push(TRUE)
		} else {
			vm.push(FALSE)
		}

	default:
		return fmt.Errorf("unsupported operator for binary operation: %s %s", left.Type(), right.Type())
	}
	return nil
}

// Ejecuta una operación binaria con Strings (solo se soporta el operador '+')
func (vm *VM) executeBinaryString(left object.Object, op code.OpCode, right object.Object) error {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch op {
	case code.OpAdd:
		vm.push(&object.String{Value: string(leftVal + rightVal)})
	default:
		return fmt.Errorf("unsupported operator for binary operation: %s %s", left.Type(), right.Type())
	}
	return nil
}

// Ejecuta una operación bnaria con Booleans
func (vm *VM) executeBinaryBoolean(left object.Object, op code.OpCode, right object.Object) error {
	if op == code.OpAnd || op == code.OpOr {
		return vm.executeBinaryLogic(left, op, right)
	}
	if op != code.OpLess && op != code.OpLessEq && op != code.OpGreater && op != code.OpGreaterEq && op != code.OpEqual && op != code.OpNotEq {
		return fmt.Errorf("unsupported operator for binary operation %s %s", left.Type(), right.Type())
	}
	// convertimos boolean a integer
	leftInteger := &object.Integer{Value: 0}
	rightInteger := &object.Integer{Value: 0}

	if left.(*object.Boolean).Value {
		leftInteger.Value = 1
	}

	if right.(*object.Boolean).Value {
		rightInteger.Value = 1
	}

	return vm.executeBinaryInteger(leftInteger, op, rightInteger)
}

// Ejecuta una operación binaria con Lógicos
func (vm *VM) executeBinaryLogic(left object.Object, op code.OpCode, right object.Object) error {
	leftBoolValue := left.(*object.Boolean).Value
	rightBoolValue := right.(*object.Boolean).Value

	switch op {
	case code.OpAnd:
		if leftBoolValue && rightBoolValue {
			return vm.push(TRUE)
		} else {
			return vm.push(FALSE)
		}
	case code.OpOr:
		if leftBoolValue || rightBoolValue {
			return vm.push(TRUE)
		} else {
			return vm.push(FALSE)
		}
	default:
		return fmt.Errorf("unsupported operator for binary operation: %s %s", left.Type(), right.Type())
	}
}
