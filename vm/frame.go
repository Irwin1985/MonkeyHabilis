package vm

import (
	"MonkeyHabilis/object"
)

type Frame struct {
	cf *object.CompiledFunction
	ip int
}

func NewFrame(compiledFunction *object.CompiledFunction) *Frame {
	return &Frame{
		cf: compiledFunction,
		ip: -1, // -1 para que el siguiente loop de la vm lo lleve a cero.
	}
}
