package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime/pprof"
	"strings"
	"time"
)

// Cada variante deixa 2 goroutines presas para sempre (n=3, uma recebe).

// capturaI: o closure captura a variavel do laco.
func capturaI(n int) {
	ch := make(chan int)
	for i := 0; i < n; i++ {
		go func() { ch <- i }()
	}
	<-ch
}

// copiaK: o closure captura uma copia local.
func copiaK(n int) {
	ch := make(chan int)
	for i := 0; i < n; i++ {
		k := i
		go func() { ch <- k }()
	}
	<-ch
}

// argumento: o valor entra como argumento da goroutine.
func argumento(n int) {
	ch := make(chan int)
	for i := 0; i < n; i++ {
		go func(v int) { ch <- v }(i)
	}
	<-ch
}

// constante: o closure so captura o canal.
func constante(n int) {
	ch := make(chan int)
	for range n {
		go func() { ch <- 1 }()
	}
	<-ch
}

func main() {
	variantes := map[string]func(int){
		"capturaI": capturaI, "copiaK": copiaK, "argumento": argumento, "constante": constante,
	}
	nome := os.Args[1]
	variantes[nome](3)
	time.Sleep(100 * time.Millisecond) // todas as goroutines ja bloquearam

	var buf bytes.Buffer
	pprof.Lookup("goroutineleak").WriteTo(&buf, 1)
	primeira, _, _ := strings.Cut(buf.String(), "\n")
	fmt.Printf("%-10s %s\n", nome, primeira)
}
