package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Agendador executa tarefa a cada intervalo, até o contexto terminar.
// Quando o contexto termina, espera a execução em andamento acabar antes de retornar.
func Agendador(ctx context.Context, intervalo time.Duration, tarefa func(ctx context.Context)) {
	t := time.NewTicker(intervalo)
	defer t.Stop()

	var emAndamento sync.WaitGroup
	defer emAndamento.Wait() // parada graciosa: não abandona uma execução pela metade

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			emAndamento.Go(func() { tarefa(ctx) })
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()
	execucoes := 0
	Agendador(ctx, 100*time.Millisecond, func(ctx context.Context) { execucoes++ })
	fmt.Println("execuções:", execucoes)
}
