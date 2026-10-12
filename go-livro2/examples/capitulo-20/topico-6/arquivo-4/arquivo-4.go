package main

import (
	"fmt"
	"runtime/cgo"
)

func main() {
	h := cgo.NewHandle("dado do Go")
	fmt.Printf("handle: %d (um inteiro, que o C pode guardar)\n", uintptr(h))
	fmt.Println("valor:", h.Value())
	h.Delete()

	defer func() { fmt.Println("recuperado:", recover()) }()
	fmt.Println(h.Value())
}
