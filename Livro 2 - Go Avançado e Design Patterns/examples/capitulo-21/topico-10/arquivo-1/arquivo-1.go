package main

import (
	"fmt"
	"runtime"
	"time"

	logs "exemplos/livro2/capitulo-21/topico-4"
)

type no struct {
	prox *no
	dado [56]byte
}

func main() {
	// heap vivo: ~100 MB de objetos pequenos que o GC precisa varrer
	var vivos []*no
	for i := 0; i < 1_500_000; i++ {
		vivos = append(vivos, &no{})
	}

	linhas := logs.Gerar(20000)
	inicio := time.Now()
	for range 150 {
		logs.ResumirV1(linhas) // gera ~7,5 MB de lixo por chamada
	}
	dur := time.Since(inicio)

	var m runtime.MemStats
	defer runtime.KeepAlive(vivos) // sem isto o coletor pode liberar a lista antes do fim
	runtime.ReadMemStats(&m)
	fmt.Printf("tempo %v | ciclos de GC %d | pausa total %v | heap em uso %d MB | vivos %d\n",
		dur.Round(time.Millisecond), m.NumGC, time.Duration(m.PauseTotalNs).Round(time.Microsecond),
		m.HeapInuse>>20, len(vivos))
}
