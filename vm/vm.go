package vm

import (
	"MonkeyHabilis/code"
	"MonkeyHabilis/compiler"
	"MonkeyHabilis/object"
	"fmt"
)

var STACK_SIZE = 2048
var TRUE = &object.Boolean{Value: true}
var FALSE = &object.Boolean{Value: false}
var NULL = &object.Null{}

type VM struct {
	stack        []object.Object        // la pila de objetos
	sp           int                    // este es el puntero de la pila que siempre apunta al último objeto
	instructions []compiler.Instruction // la lista de instrucciones a ejecutar
	objectPool   []object.Object        // la lista de literales constantes
}

// Crea la máquina virtual
func New(bytecode *compiler.ByteCode) *VM {
	vm := &VM{
		stack:        make([]object.Object, STACK_SIZE),
		sp:           0,
		instructions: bytecode.Instructions,
		objectPool:   bytecode.ObjectPool,
	}
	return vm
}

// Comienza el ciclo Fetch-Decode-Execute
func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		// paso 1, obtener la instrucción a ejecutar
		instruction := vm.instructions[ip]
		switch instruction.OpCode {
		case code.OpConstant:
			// Agregar una constante a la pila, necesitamos su índice entonces
			// lo tomamos de la lista de objetos y usamos el campo instruction.index
			index := instruction.Position
			obj := vm.objectPool[index]
			vm.push(obj)

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
				ip = instruction.Position - 1 // le resto 1 para que comience exactamente en el número correcto.
			}
		case code.OpJump:
			// saltamos sin preguntar al índice
			ip = instruction.Position - 1
		case code.OpPop:
			vm.pop()
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
