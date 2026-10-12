package main

import (
	"fmt"
	"sync"
)

func main() {
	contador := 0
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() { contador++ })
	}
	wg.Wait()
	fmt.Println("contador:", contador)
}
