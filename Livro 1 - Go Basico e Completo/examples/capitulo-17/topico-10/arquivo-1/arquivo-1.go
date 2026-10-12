package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // registra /debug/pprof/* no DefaultServeMux
	"runtime"
	"sync"
	"time"
)

var (
	mu    sync.Mutex
	conta int
)

func trabalho(w http.ResponseWriter, r *http.Request) {
	// CPU: uma conta inútil
	s := 0
	for i := 0; i < 5_000_000; i++ {
		s += i % 7
	}
	// contenção: todo mundo disputa a mesma trava
	mu.Lock()
	time.Sleep(2 * time.Millisecond)
	conta++
	mu.Unlock()
	fmt.Fprintln(w, "ok", s, conta)
}

func main() {
	runtime.SetMutexProfileFraction(1) // amostrar toda disputa de Mutex (custa desempenho: só para investigar)
	runtime.SetBlockProfileRate(1)     // amostrar todo bloqueio
	http.HandleFunc("/trabalho", trabalho)
	// o servidor de diagnóstico fica em outra porta, só em localhost
	go func() { log.Fatal(http.ListenAndServe("localhost:6060", nil)) }()
	log.Fatal(http.ListenAndServe("localhost:8077", nil))
}
