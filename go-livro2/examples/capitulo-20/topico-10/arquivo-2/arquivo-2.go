package main

/*
#include <unistd.h>

static void dormir_ms(int ms) { usleep(ms * 1000); }
*/
import "C"

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

func threads() string {
	b, _ := os.ReadFile("/proc/self/status")
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "Threads:") {
			return strings.TrimSpace(strings.TrimPrefix(l, "Threads:"))
		}
	}
	return "?"
}

// vagas limita quantas goroutines estao DENTRO do C ao mesmo tempo.
var vagas = make(chan struct{}, 8)

func dormirLimitado(ms int) {
	vagas <- struct{}{}
	defer func() { <-vagas }()
	C.dormir_ms(C.int(ms))
}

func main() {
	var wg sync.WaitGroup
	inicio := time.Now()
	for range 100 {
		wg.Go(func() { dormirLimitado(100) })
	}
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("100 goroutines, 8 vagas: %s threads do SO\n", threads())
	wg.Wait()
	fmt.Printf("tempo total: %v (100 chamadas de 100 ms, 8 por vez)\n", time.Since(inicio).Round(100*time.Millisecond))
}
