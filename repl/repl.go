package repl

import (
	"MonkeyHabilis/compiler"
	"MonkeyHabilis/lexer"
	"MonkeyHabilis/object"
	"MonkeyHabilis/parser"
	"MonkeyHabilis/vm"
	"bufio"
	"fmt"
	"io"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	objectPool := []object.Object{}
	globals := make([]object.Object, vm.GLOBAL_SIZE)

	symbolTable := compiler.NewSymbolTable()
	for i, builtin := range object.Builtins {
		symbolTable.DefineBuiltin(i, builtin.Name)
	}

	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.Program()
		if len(p.Errors) != 0 {
			printParserErrors(out, p.Errors)
			continue
		}

		comp := compiler.NewWithState(symbolTable, objectPool)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		/**************************INICIO DEBUG************************/
		strBytecode := comp.PrintInstructions(comp.GetInstructions())
		fmt.Print(strBytecode)
		/**************************FIN DEBUG***************************/

		// Obtenemos el bytecode y mantenemos la lista de constantes
		byteCode := comp.GetByteCode()
		objectPool = byteCode.ObjectPool

		machine := vm.NewWithGlobalsStore(byteCode, globals)
		err = machine.Run()

		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElem()
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}
}

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
