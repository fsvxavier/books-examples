package main

import (
	"fmt"

	"golang.org/x/sync/errgroup"
)

func main() {
	var g errgroup.Group
	g.Go(func() error {
		var m map[string]int
		m["x"] = 1 // panic dentro da goroutine do grupo
		return nil
	})
	err := g.Wait()
	fmt.Println("Wait devolveu:", err) // nunca chega aqui
}
