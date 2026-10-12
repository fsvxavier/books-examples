package main

/*
typedef int (*operacao)(int, int);

int soma(int a, int b)  { return a + b; }
int mult(int a, int b)  { return a * b; }

// o C chama o ponteiro de funcao; o Go so o carrega de um lado para o outro
static int aplicar(operacao f, int a, int b) { return f(a, b); }
*/
import "C"

import "fmt"

func main() {
	tabela := map[string]C.operacao{
		"soma": C.operacao(C.soma),
		"mult": C.operacao(C.mult),
	}
	for _, nome := range []string{"soma", "mult"} {
		fmt.Println(nome, C.aplicar(tabela[nome], 6, 7))
	}
}
