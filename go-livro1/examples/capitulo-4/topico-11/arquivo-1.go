package main

import (
	"fmt"
	"os"
)

var configuracao = carregar()

func carregar() string {
	fmt.Println("1. variável de pacote inicializada")
	return "ok"
}

func init() {
	fmt.Println("2. init #1")
}

func init() {
	fmt.Println("3. init #2")
}

func main() {
	fmt.Println("4. main")
	defer fmt.Println("defer: não roda com os.Exit")
	os.Exit(3)
}
