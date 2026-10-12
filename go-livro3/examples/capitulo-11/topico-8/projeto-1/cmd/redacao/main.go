// Comando redacao executa a equipe de agentes em vários cenários.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"exemplo.com/agentes/internal/equipe"
	"exemplo.com/agentes/internal/grafo"
	"exemplo.com/agentes/internal/modelo"
)

func main() {
	var (
		subtemas    = flag.String("subtemas", "", "subtemas separados por vírgula (pesquisa em paralelo)")
		latencia    = flag.Duration("latencia", 0, "latência simulada de cada chamada ao modelo")
		paralelismo = flag.Int("paralelismo", 1, "pesquisas simultâneas")
		minFontes   = flag.Int("min-fontes", 2, "fontes que o revisor exige")
		limite      = flag.Int("limite", 25, "máximo de passos do grafo")
		falharEm    = flag.Int64("falhar-em", 0, "número da chamada ao modelo que falha uma vez")
		tentativas  = flag.Int("tentativas", 1, "tentativas por chamada ao modelo")
		maxTokens   = flag.Int("max-tokens", 0, "teto de tokens (0 = sem teto)")
		ponto       = flag.String("ponto", "", "arquivo do ponto de retomada")
		aprovacao   = flag.String("aprovacao", "", "sim, nao ou espera (humano que nunca responde)")
		prazo       = flag.Duration("prazo", 0, "prazo total da execução")
		quiet       = flag.Bool("quieto", false, "não imprime os passos")
		json        = flag.Bool("json", false, "registra cada passo em JSON (slog)")
	)
	flag.Parse()

	m := &modelo.Falso{Latencia: *latencia, Base: equipe.BaseDemo(), MinFontes: *minFontes, FalharEm: *falharEm}
	e := &equipe.Equipe{Modelo: m, Paralelismo: *paralelismo}

	op := equipe.Opcoes{Tentativas: *tentativas, Espera: 50 * time.Millisecond, MaxTokens: *maxTokens}
	if *aprovacao != "" {
		c := make(chan bool, 1)
		switch *aprovacao {
		case "sim":
			c <- true
		case "nao":
			c <- false
		}
		op.Aprovacao = c
	}
	g := e.Grafo(op)
	g.Limite = *limite
	if *json {
		log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey || a.Key == "duracao" {
					return slog.Attr{} // tira o que muda a cada execução
				}
				return a
			},
		}))
		g.Observar = func(ev grafo.Evento) {
			log.Info("passo", "execucao", "r-001", "n", ev.Passo, "no", ev.No, "proximo", ev.Proximo, "erro", ev.Err)
		}
	} else if !*quiet {
		g.Observar = func(ev grafo.Evento) {
			fmt.Printf("passo %2d  %-10s -> %-8s", ev.Passo, ev.No, ev.Proximo)
			if ev.Err != nil {
				fmt.Printf(" ERRO: %v", ev.Err)
			}
			fmt.Println()
		}
	}

	estado := equipe.Estado{Tema: "goroutines"}
	if *subtemas != "" {
		estado = equipe.Estado{Tema: "concorrência", Subtemas: strings.Split(*subtemas, ",")}
	}

	ctx := context.Background()
	if *prazo > 0 {
		var cancelar context.CancelFunc
		ctx, cancelar = context.WithTimeout(ctx, *prazo)
		defer cancelar()
	}

	var pt grafo.Ponto[equipe.Estado]
	if *ponto != "" {
		pt = grafo.ArquivoJSON[equipe.Estado]{Caminho: *ponto}
	}

	inicio := time.Now()
	final, err := g.ExecutarComPonto(ctx, estado, pt)
	dur := time.Since(inicio).Round(time.Millisecond)
	if err != nil {
		fmt.Printf("ERRO: %v\n", err)
	}
	fmt.Printf("rodadas=%d aprovado=%v tokens=%d chamadas=%d tempo=%v\n",
		final.Rodadas, final.Aprovado, final.Tokens, m.Chamadas(), dur)
	if final.Aprovado {
		fmt.Println("texto final:", strings.TrimSpace(final.Rascunho))
	}
	if err != nil {
		if errors.Is(err, grafo.ErrLimite) {
			os.Exit(3)
		}
		os.Exit(1)
	}
}
