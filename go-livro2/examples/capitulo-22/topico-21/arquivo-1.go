// Comando fv chama funções por valor em um laço quente e grava o próprio perfil.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

type operacao func(x uint64) uint64

// mistura cabe no orçamento normal de inlining, mas só pode ser expandida
// quando o compilador souber que é ela que está sendo chamada.
func mistura(x uint64) uint64 {
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	x *= 0xc4ceb9fe1a85ec53
	x ^= x >> 33
	if x%7 == 0 {
		x += 0x9e3779b97f4a7c15
	}
	for i := 0; i < 2; i++ {
		x = x<<13 | x>>51
		x ^= uint64(i) * 0x2545f4914f6cdd1d
	}
	return x
}

func gira(x uint64) uint64 { return x<<7 | x>>57 }

func main() {
	saida := flag.String("o", "", "arquivo de perfil de CPU (opcional)")
	dur := flag.Duration("d", 3*time.Second, "duração")
	flag.Parse()

	if *saida != "" {
		f, err := os.Create(*saida)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	// 31 de cada 32 posições usam mistura.
	ops := make([]operacao, 32)
	for i := range ops {
		ops[i] = mistura
	}
	ops[31] = gira

	var x uint64 = 1
	var voltas int64
	inicio := time.Now()
	for time.Since(inicio) < *dur {
		for i := 0; i < 100000; i++ {
			f := ops[i&31]
			x = f(x)
		}
		voltas += 100000
	}
	fmt.Printf("%.2f ns por chamada (x=%d)\n", float64(time.Since(inicio).Nanoseconds())/float64(voltas), x)
}
