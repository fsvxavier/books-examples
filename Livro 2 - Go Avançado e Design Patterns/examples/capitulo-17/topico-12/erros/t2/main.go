package main

import "fmt"

// Impossivel não tem nenhum tipo: nenhum tipo é int e string ao mesmo tempo.
type Impossivel interface {
	~int
	~string
}

func Usa[T Impossivel](x T) T { return x }

func main() { fmt.Println(Usa(1)) }
