// Comando comparar roda N conversas de três jeitos e mede o tempo.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-11/topico-8/projeto-1/internal/caixa"
	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-11/topico-8/projeto-1/internal/equipe"
	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-11/topico-8/projeto-1/internal/modelo"
)

func main() {
	modo := flag.String("modo", "grafo-seq", "grafo-seq, grafo-conc ou barramento")
	n := flag.Int("n", 20, "número de conversas")
	trab := flag.Int("trabalhadores", 1, "goroutines por agente no modo barramento")
	lat := flag.Duration("latencia", 20*time.Millisecond, "latência de cada chamada ao modelo")
	flag.Parse()

	m := &modelo.Falso{Latencia: *lat, Base: equipe.BaseDemo(), MinFontes: 2}
	antes := runtime.NumGoroutine()
	inicio := time.Now()

	switch *modo {
	case "grafo-seq":
		e := &equipe.Equipe{Modelo: m, Paralelismo: 1}
		for range *n {
			if _, err := e.Grafo().Executar(context.Background(), equipe.Estado{Tema: "goroutines"}); err != nil {
				fatal(err)
			}
		}
	case "grafo-conc":
		e := &equipe.Equipe{Modelo: m, Paralelismo: 1}
		var wg sync.WaitGroup
		for range *n {
			wg.Go(func() {
				if _, err := e.Grafo().Executar(context.Background(), equipe.Estado{Tema: "goroutines"}); err != nil {
					fatal(err)
				}
			})
		}
		wg.Wait()
	case "barramento":
		b := caixa.MontarComTrabalhadores(m, *n, 5, *trab)
		ctx, parar := context.WithCancel(context.Background())
		feito := make(chan error, 1)
		go func() { feito <- b.Executar(ctx) }()
		for i := 1; i <= *n; i++ {
			if err := b.Enviar(ctx, caixa.Mensagem{Conversa: i, Para: "pesquisador", Tipo: "tema", Corpo: "goroutines"}); err != nil {
				fatal(err)
			}
		}
		for range *n {
			<-b.Caixa("saida")
		}
		parar()
		if err := <-feito; err != nil {
			fatal(err)
		}
	default:
		fatal(fmt.Errorf("modo desconhecido: %q", *modo))
	}

	dur := time.Since(inicio).Round(time.Millisecond)
	time.Sleep(10 * time.Millisecond) // deixa as goroutines terminarem de sair
	fmt.Printf("modo=%-10s trab=%d conversas=%d chamadas=%d tempo=%v goroutines antes=%d depois=%d\n",
		*modo, *trab, *n, m.Chamadas(), dur, antes, runtime.NumGoroutine())
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "ERRO:", err)
	os.Exit(1)
}
