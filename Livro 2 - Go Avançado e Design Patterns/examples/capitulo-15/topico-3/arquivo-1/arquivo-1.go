package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// consultar simula uma chamada lenta que entrega o resultado por um canal sem buffer.
func consultar(ctx context.Context, id int) int {
	ch := make(chan int) // sem buffer: o envio só termina se alguém receber
	go func() {
		time.Sleep(20 * time.Millisecond)
		ch <- id * 2 // RUIM: se o chamador desistiu, esta goroutine fica presa aqui para sempre
	}()
	select {
	case v := <-ch:
		return v
	case <-ctx.Done():
		return -1
	}
}

// consultarOk usa canal com buffer 1: o envio nunca bloqueia, mesmo sem receptor.
func consultarOk(ctx context.Context, id int) int {
	ch := make(chan int, 1)
	go func() {
		time.Sleep(20 * time.Millisecond)
		ch <- id * 2
	}()
	select {
	case v := <-ch:
		return v
	case <-ctx.Done():
		return -1
	}
}

func medir(nome string, f func(context.Context, int) int) {
	runtime.GC()
	antes := runtime.NumGoroutine()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	memAntes := ms.HeapAlloc + ms.StackInuse

	for i := range 10000 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond) // o chamador desiste em 1 ms
		f(ctx, i)
		cancel()
	}
	time.Sleep(200 * time.Millisecond) // tempo de sobra para quem ia terminar terminar
	runtime.GC()
	depois := runtime.NumGoroutine()
	runtime.ReadMemStats(&ms)
	memDepois := ms.HeapAlloc + ms.StackInuse
	fmt.Printf("%-22s goroutines: %5d -> %5d   memória: +%.1f MiB\n", nome, antes, depois,
		float64(memDepois-memAntes)/(1<<20))
}

func main() {
	medir("canal sem buffer", consultar)
	medir("canal com buffer 1", consultarOk)
}
