package vm

import (
	"MonkeyHabilis/code"
	"MonkeyHabilis/compiler"
	"MonkeyHabilis/object"
	"fmt"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Stack Pointer o puntero de la pila. Apunta siempre al siguiente valor.
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
}

// devuelve el último elemento de la pila
func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

// empuja un objeto en la pila
func (vm *VM) push(element object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack Overflow")
	}

	vm.stack[vm.sp] = element
	vm.sp += 1

	return nil
}

// remueve y devuelve el elemento de la pila
func (vm *VM) pop() object.Object {
	element := vm.stack[vm.sp-1]
	vm.sp -= 1
	return element
}

// ejecuta el bytecode
func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()
			leftValue := left.(*object.Integer).Value
			rightValue := right.(*object.Integer).Value

			result := leftValue + rightValue
			vm.push(&object.Integer{Value: result})
		case code.OpSub:
			right := vm.pop()
			left := vm.pop()
			result := left.(*object.Integer).Value - right.(*object.Integer).Value
			vm.push(&object.Integer{Value: result})
		case code.OpMul:
			right := vm.pop()
			left := vm.pop()
			result := left.(*object.Integer).Value * right.(*object.Integer).Value
			vm.push(&object.Integer{Value: result})
		case code.OpDiv:
			rightValue := vm.pop().(*object.Integer).Value
			if rightValue == 0 {
				return fmt.Errorf("division by zero")
			}
			leftValue := vm.pop().(*object.Integer).Value
			vm.push(&object.Integer{Value: leftValue / rightValue})
		}
	}
	return nil
}
