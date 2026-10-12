package main

import (
	"fmt"
	"slices"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var vistos []int
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			vistos = append(vistos, i)
			mu.Unlock()
		}()
	}
	wg.Wait()
	slices.Sort(vistos)
	fmt.Println("valores vistos pelas goroutines:", vistos)
}
