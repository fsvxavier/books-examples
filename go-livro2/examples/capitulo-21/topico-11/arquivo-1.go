package main

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"sync"

	logs "github.com/fsvxavier/books-examples/go-livro-2/capitulo-21/topico-4"
)

// atender gasta CPU proporcional ao tamanho do lote do cliente.
func atender(ctx context.Context, cliente string, linhas []string) {
	pprof.Do(ctx, pprof.Labels("cliente", cliente, "rota", "/resumir"), func(ctx context.Context) {
		for range 400 {
			logs.ResumirV2(linhas)
		}
	})
}

func main() {
	f, _ := os.Create("cpu.prof")
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	lotes := map[string][]string{
		"A": logs.Gerar(2000),
		"B": logs.Gerar(20000), // o cliente "pesado"
		"C": logs.Gerar(4000),
	}
	var wg sync.WaitGroup
	for nome, linhas := range lotes {
		wg.Go(func() { atender(context.Background(), nome, linhas) })
	}
	wg.Wait()
	fmt.Println("pronto")
}
