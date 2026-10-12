package main

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
	"unique"
)

const n = 500_000

func heap() float64 {
	for range 3 { // varios ciclos: a limpeza do unique acontece depois do GC que detecta a perda
		runtime.GC()
		time.Sleep(20 * time.Millisecond)
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.HeapAlloc) / (1 << 20)
}

func chave(i int) string { return "identificador-unico-de-sessao-numero-" + strconv.Itoa(i) }

func main() {
	base := heap()
	fmt.Printf("heap inicial:                          %6.1f MiB\n", base)

	// 1) unique.Make com valores TODOS DIFERENTES
	hs := make([]unique.Handle[string], 0, n)
	for i := range n {
		hs = append(hs, unique.Make(chave(i)))
	}
	fmt.Printf("unique: %d handles distintos vivos:  %6.1f MiB\n", n, heap())
	runtime.KeepAlive(hs) // sem isto o compilador considera hs morta e o GC a coleta antes da medida
	hs = nil
	fmt.Printf("unique: handles descartados + GC:      %6.1f MiB\n", heap())

	// 2) internador manual (mapa global) com valores todos diferentes
	m := make(map[string]string, n)
	for i := range n {
		s := chave(i)
		m[s] = s
	}
	fmt.Printf("mapa:   %d entradas vivas:           %6.1f MiB\n", n, heap())
	// Mesmo que ninguem mais use as strings, o mapa as segura.
	fmt.Printf("mapa:   ninguem usa mais (mapa vivo):  %6.1f MiB\n", heap())
	runtime.KeepAlive(m)
	m = nil
	fmt.Printf("mapa:   mapa descartado + GC:          %6.1f MiB\n", heap())
}
