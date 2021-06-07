package vm

import (
	"MonkeyHabilis/object"
)

type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer int // guarda la posición del puntero
}

func NewFrame(cl *object.Closure, basePointer int) *Frame {
	f := &Frame{
		cl:          cl,
		ip:          -1, // -1 para que el siguiente loop de la vm lo lleve a cero.
		basePointer: basePointer,
	}
	return f
}
