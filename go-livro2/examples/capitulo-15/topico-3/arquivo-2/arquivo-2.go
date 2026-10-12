package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

var canalGlobal = make(chan int) // alcançável por uma variável de pacote

func vazaLocal() {
	ch := make(chan int)
	go func() { ch <- 1 }() // ninguém mais enxerga ch: vazamento detectável
}

func vazaGlobal() {
	go func() { canalGlobal <- 1 }() // o canal segue acessível: o runtime não prova o vazamento
}

func main() {
	for range 3 {
		vazaLocal()
	}
	for range 3 {
		vazaGlobal()
	}
	time.Sleep(50 * time.Millisecond)
	fmt.Println("perfil goroutineleak:")
	pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
	fmt.Println("---")
	fmt.Println("perfil goroutine (todas as presas):")
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
}
