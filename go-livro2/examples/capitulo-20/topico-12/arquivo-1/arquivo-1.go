package main

/*
static long total;

static void somar_um(void) { total++; }
static long ler(void) { return total; }
*/
import "C"

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 100_000 {
				C.somar_um()
			}
		})
	}
	wg.Wait()
	fmt.Println("esperado 400000, obtido", C.ler())
}
