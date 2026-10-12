// Command e16 usa o método genérico (*Rand).N de math/rand/v2, novo no Go 1.27.
package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

type Porta uint16

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	// Antes: a função genérica do pacote, rand.N, usa o gerador global.
	// Agora: o método genérico N, no gerador que escolhemos (semente fixa, resultado repetível).
	fmt.Println("r.N(10):          ", r.N(10), r.N(10), r.N(10))
	fmt.Printf("r.N(Porta(1000)): %d (%T)\n", r.N(Porta(1000)), r.N(Porta(1000)))
	fmt.Println("r.N(time.Second): ", r.N(time.Second) < time.Second)
	fmt.Printf("tipo do método instanciado: %T\n", r.N[int])
}
