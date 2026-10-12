package main

/*
#include <stdint.h>

// uma biblioteca com estado POR THREAD (como errno, OpenGL, muitas bibliotecas de GUI)
static __thread int contador;

static int proximo(void) { return ++contador; }
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// trabalhar chama proximo() n vezes e conta quantas vezes o valor subiu de 1 em 1.
func trabalhar(n int, travar bool) (certas int) {
	if travar {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}
	anterior := int(C.proximo())
	for range n - 1 {
		time.Sleep(50 * time.Microsecond) // deixa o escalonador trocar de thread
		atual := int(C.proximo())
		if atual == anterior+1 {
			certas++
		}
		anterior = atual
	}
	return certas
}

func rodar(travar bool) {
	const goroutines, chamadas = 8, 200
	var wg sync.WaitGroup
	var mu sync.Mutex
	total := 0
	for range goroutines {
		wg.Go(func() {
			c := trabalhar(chamadas, travar)
			mu.Lock()
			total += c
			mu.Unlock()
		})
	}
	wg.Wait()
	esperado := goroutines * (chamadas - 1)
	fmt.Printf("LockOSThread=%-5v saltos de +1 certos: %4d de %d\n", travar, total, esperado)
}

func main() {
	rodar(false)
	rodar(true)
}
