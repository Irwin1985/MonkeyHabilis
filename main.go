package main

import (
	"MonkeyHabilis/compiler"
	"MonkeyHabilis/lexer"
	"MonkeyHabilis/parser"
	"MonkeyHabilis/token"
	"MonkeyHabilis/vm"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
)

func main() {
	// testLexer()
	// testParser()
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the Monkey Habilis programming language!\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")
	//testVirtualMachine()
	//repl.Start(os.Stdin, os.Stdout)
	testMyVersion()
}

func testMyVersion() {
	//mode := "debug"
	mode := "console"

	if mode == "console" {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print(">> ")
			scanned := scanner.Scan()
			if !scanned {
				return
			}

			line := scanner.Text()
			l := lexer.New(line)
			p := parser.New(l)

			program := p.Program()
			if len(p.Errors) != 0 {
				fmt.Printf("errors: %v", p.Errors)
				continue
			}

			comp := compiler.New()
			err := comp.Compile(program)
			if err != nil {
				fmt.Fprintf(os.Stdout, "Woops! Compilation failed:\n %s\n", err)
			}

			machine := vm.New(comp.GetByteCode())
			err = machine.Run()

			if err != nil {
				fmt.Fprintf(os.Stdout, "Woops! Executing bytecode failed:\n %s\n", err)
				continue
			}

			stackTop := machine.LastPoppedStackElem()
			io.WriteString(os.Stdout, stackTop.Inspect())
			io.WriteString(os.Stdout, "\n")
		}
	} else {
		input := `if( 5 == 5 ) {"Correcto!"} else {"Que?, estas loco?"};`
		l := lexer.New(input)
		p := parser.New(l)
		program := p.Program()

		c := compiler.New()
		c.Compile(program)

		// DEBUG
		strBytecode := c.DumpInstructions()
		fmt.Print(strBytecode)

		vm := vm.New(c.GetByteCode())
		err := vm.Run()
		if err != nil {
			fmt.Printf("Woops! Executing bytecode failed:\n %s\n", err)
		} else {
			obj := vm.LastPoppedStackElem()
			fmt.Print(obj.Inspect())
		}
	}
}

func testParser() {
	var input = `
		while (true) {
			print("forever loop!");
		};
	`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.Program()

	if len(p.Errors) > 0 {
		for _, msg := range p.Errors {
			fmt.Print(msg)
		}
	} else {
		if program != nil {
			fmt.Print(program.String())
		}
	}

}

func testLexer() {
	var input = `
	/*
	 * BIENVENIDO AL TUTORIAL MONKEYHABILIS
	 * EL SIGUIENTE TOUR TE MOSTRARÁ TODO LO
	 * QUE PUEDES HACER CON EL LENGUAJE.
	*/

	// expressiones aritmeticas
	1 + 1;
	5 - 2;
	3 * 3;
	4 / 2;
	
	// expressiones lógicas
	true && false;
	false || true;
	!true && !false;
	!!false || !!true;
	
	// expressiones relacionales
	5 <= 3;
	3 >= 4;
	4 == 5;
	5 != 6;

	// string
	let nombre = "Irwin";
	let apellido = "Rodriguez";
	let nombreCompleto = nombre + " " + apellido;

	// arrays
	let lenguages = ["Java", "C++", "Go"];
	
	// arrays anónimos
	["Manzanas", "Peras", "Piñas"];

	// diccionarios
	let keywords = {
		"fn":     token.FUNCTION,
		"let":    token.LET,
		"true":   token.TRUE,
		"false":  token.FALSE,
		"if":     token.IF,
		"else":   token.ELSE,
		"return": token.RETURN
	};

	// diccionarios anónimos
	{"name": "John", "band": "The Beatles", "year": 1963}["year"]; // 1963

	// funciones
	let sumar = fn(a, b){
		return a + b;
	};
	
	let factorial = fn(n) {
		if (n <= 1) {
			return n
		} else {
			return n * factorial(n-1);
		}
	};
	factorial(4); // 24

	// funciones anónimas
	fn(a, b){
		if (a > b){
			return a;
		} else{
			return b;
		}
	}(5, 10);

	// funciones de primera clase
	let cuadrado = fn(n) {
		return n * n;
	};
	let cubo = fn(base, fnCuadrado){
		return fnCuadrado(n) * n;
	};
	// uso
	cubo(3, cuadrado);

	// closures
	let iniciar = fn(){
		let nombre = "Irwin"; // esta variable es local creada por "iniciar".
		let mostrarNombre = fn() {
			print(nombre); // usa la variable declarada en la función externa.
		}
		mostrarNombre();
	};

	// uso del closure
	iniciar();

	// funciones integradas
	/*
	* todo tipo en MonkeyHabilis subyace en un tipo de clase por lo que
	* es posible invocar algunos métodos predefinidos según su tipo.
	*/
	let nombre = "Irwin"; // declara una variable string
	nombre.size(); // el método size() está disponible para este tipo.
	nombre.substr(1, 3); // substr(nStart, nLengh) es único para este tipo.
	nombre.at('w'); // devuelve el índice donde se encuentre el caracter en la cadena.
	`
	input = "1985;"
	var l = lexer.New(input)
	var tok = l.NextToken()
	for tok.Type != token.EOF {
		fmt.Printf("Type: '%s', Literal: '%s'\n", tok.Type, tok.Literal)
		tok = l.NextToken()
	}
}
