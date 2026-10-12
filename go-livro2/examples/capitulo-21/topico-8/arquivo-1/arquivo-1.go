package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

var (
	mu       sync.Mutex
	contador int
)

// secaoCritica segura o mutex por 1 ms: todas as outras goroutines esperam.
func secaoCritica() {
	mu.Lock()
	defer mu.Unlock()
	time.Sleep(time.Millisecond)
	contador++
}

func disputa() {
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 20 {
				secaoCritica()
			}
		})
	}
	wg.Wait()
}

// buscarPrimeiro devolve a primeira resposta e abandona as outras
// goroutines, que ficam presas para sempre tentando enviar em um canal
// sem buffer que ninguem mais le.
func buscarPrimeiro(n int) int {
	ch := make(chan int)
	for i := 0; i < n; i++ {
		go func(v int) { ch <- v }(i)
	}
	return <-ch
}

func salvar(nome, arquivo string) {
	f, err := os.Create(arquivo)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	pprof.Lookup(nome).WriteTo(f, 0)
}

func main() {
	runtime.SetMutexProfileFraction(1) // registra todos os eventos de contencao
	runtime.SetBlockProfileRate(1)     // registra todo bloqueio (em ns)

	disputa()
	salvar("mutex", "mutex.prof")
	salvar("block", "block.prof")
	fmt.Println("contador:", contador)

	for range 10 {
		buscarPrimeiro(5)
	}
	time.Sleep(100 * time.Millisecond) // da tempo de todas bloquearem
	fmt.Println("goroutines vivas:", runtime.NumGoroutine())
	pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
}
