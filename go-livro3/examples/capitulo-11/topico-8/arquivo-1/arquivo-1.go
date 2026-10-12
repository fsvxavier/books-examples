// Comando impasse mostra dois agentes que se travam enviando um ao outro.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"
)

type msg struct{ de, texto string }

func agente(ctx context.Context, nome string, entrada <-chan msg, saida chan<- msg, avisos chan<- string, seguro bool) {
	// Cada agente começa mandando um pedido ao outro e só depois lê a sua caixa.
	m := msg{de: nome, texto: "preciso da sua opinião"}
	if seguro {
		select {
		case saida <- m:
		case <-ctx.Done():
			return
		}
	} else {
		saida <- m // bloqueia até o outro ler, mas o outro também está enviando
	}
	recebida := <-entrada
	avisos <- fmt.Sprintf("%s recebeu de %s: %q", nome, recebida.de, recebida.texto)
}

func main() {
	seguro := flag.Bool("corrigido", false, "usa caixas com capacidade 1 e contexto")
	flag.Parse()

	capacidade := 0
	if *seguro {
		capacidade = 1
	}
	aParaB := make(chan msg, capacidade)
	bParaA := make(chan msg, capacidade)
	avisos := make(chan string, 2)

	ctx, cancelar := context.WithTimeout(context.Background(), time.Second)
	defer cancelar()
	go agente(ctx, "A", bParaA, aParaB, avisos, *seguro)
	go agente(ctx, "B", aParaB, bParaA, avisos, *seguro)

	for range 2 {
		fmt.Println(<-avisos)
	}
}
