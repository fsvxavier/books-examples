package main

import "fmt"

var versao = "1.0.0" // variável de pacote
var contador, limite = 0, 10

func main() {
	var nome string = "Ana"
	var idade = 32
	cidade := "São Paulo"
	var ativo bool
	fmt.Println(nome, idade, cidade, ativo)

	a, b := 1, 2
	a, b = b, a
	fmt.Println(a, b)

	valor, err := 10, error(nil)
	valor2, err := 20, fmt.Errorf("falhou")
	fmt.Println(valor, valor2, err)

	x := 10
	if x > 5 {
		x := 99 // sombreia o x externo
		fmt.Println("dentro:", x)
	}
	fmt.Println("fora:", x)

	_, resto := 7/2, 7%2
	fmt.Println(resto, versao, contador, limite)
}
