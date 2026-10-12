package main

import (
	"fmt"
	"sync"
)

var once sync.Once

func inicializar() {
	once.Do(func() {
		// inicializacao que chama quem a chamou: o segundo Do espera o primeiro terminar.
		inicializar()
	})
}

func main() {
	fmt.Println("antes")
	inicializar()
	fmt.Println("depois (nunca aparece)")
}
