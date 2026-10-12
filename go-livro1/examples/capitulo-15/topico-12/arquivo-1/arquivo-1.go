// Comando worker é o segundo binário do projeto: processa tarefas em segundo plano.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer parar()

	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker encerrado")
			return
		case <-t.C:
			fmt.Println("worker: verificando fila...")
		}
	}
}
