package main

/*
#include <unistd.h>

static void dormir_ms(int ms) { usleep(ms * 1000); }
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
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

func rodar(rotulo string, n int, espera func()) {
	var wg sync.WaitGroup
	for range n {
		wg.Go(espera)
	}
	// amostra enquanto as goroutines esperam
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("%-22s %3d goroutines bloqueadas: %3s threads do SO\n", rotulo, n, threads())
	wg.Wait()
}

func main() {
	runtime.GOMAXPROCS(2)
	fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0), "| threads no inicio:", threads(), "| chamadas ao C no inicio:", runtime.NumCgoCall())

	rodar("time.Sleep (Go)", 100, func() { time.Sleep(500 * time.Millisecond) })
	rodar("C.usleep (cgo)", 100, func() { C.dormir_ms(500) })

	s := []metrics.Sample{{Name: "/cgo/go-to-c-calls:calls"}}
	metrics.Read(s)
	fmt.Println("chamadas Go->C contadas:", s[0].Value.Uint64(), "| runtime.NumCgoCall:", runtime.NumCgoCall())

	if len(os.Args) > 1 {
		debug.SetMaxThreads(40)
		fmt.Println("limite de threads reduzido para 40")
		rodar("C.usleep (cgo)", 100, func() { C.dormir_ms(500) })
	}
}
