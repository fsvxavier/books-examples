package main

import (
	"fmt"
	"runtime/metrics"
	"sync"
	"time"

	logs "exemplos/livro2/capitulo-21/topico-4"
)

func p99(h *metrics.Float64Histogram) float64 {
	var total uint64
	for _, c := range h.Counts {
		total += c
	}
	alvo, acum := uint64(float64(total)*0.99), uint64(0)
	for i, c := range h.Counts {
		acum += c
		if acum > alvo {
			return h.Buckets[i+1]
		}
	}
	return 0
}

func main() {
	nomes := []string{
		"/sched/gomaxprocs:threads",
		"/gc/heap/allocs:bytes",
		"/gc/heap/live:bytes",
		"/gc/cycles/total:gc-cycles",
		"/memory/classes/total:bytes",
		"/cgo/go-to-c-calls:calls",
		"/sched/latencies:seconds",
		"/gc/pauses:seconds",
	}
	amostras := make([]metrics.Sample, len(nomes))
	for i, n := range nomes {
		amostras[i].Name = n
	}

	// carga: 8 goroutines resumindo logs
	linhas := logs.Gerar(20000)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 10 {
				logs.ResumirV1(linhas)
			}
		})
	}
	time.Sleep(100 * time.Millisecond)
	durante := []metrics.Sample{{Name: "/sched/goroutines:goroutines"}}
	metrics.Read(durante) // enquanto a carga roda
	fmt.Printf("%-32s %d (durante a carga)\n", durante[0].Name, durante[0].Value.Uint64())
	wg.Wait()
	metrics.Read(amostras) // depois que ela termina

	for _, a := range amostras {
		switch a.Value.Kind() {
		case metrics.KindUint64:
			fmt.Printf("%-32s %d\n", a.Name, a.Value.Uint64())
		case metrics.KindFloat64Histogram:
			fmt.Printf("%-32s p99 < %.6f s\n", a.Name, p99(a.Value.Float64Histogram()))
		}
	}
	fmt.Println("metricas disponiveis nesta versao:", len(metrics.All()))
}
