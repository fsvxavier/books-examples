package main

import "fmt"

type ComIdentificador[K comparable] interface{ ObterID() K }

type RepoBom[K comparable, V ComIdentificador[K]] struct{ dados map[K]V }

// Produto não tem o método ObterID.
type Produto struct{ Nome string }

func main() {
	r := &RepoBom[string, Produto]{dados: map[string]Produto{}}
	fmt.Println(len(r.dados))
}
