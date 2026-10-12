package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Contador struct {
	mu sync.Mutex
	n  int
}

// RUIM: receptor por valor copia o mutex junto.
func (c Contador) Incrementar() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func copiar(c Contador) int { return c.n }

func iniciar() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_ = ctx // cancel nunca é chamado: o temporizador vaza até expirar
	_ = cancel
}

func iniciarOk() context.CancelFunc {
	_, cancel := context.WithTimeout(context.Background(), time.Second)
	return cancel
}

func iniciarLost() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	fmt.Println(ctx.Err())
	if ctx.Err() != nil {
		return
	}
	cancel()
}

func main() {
	var c Contador
	c.Incrementar()
	fmt.Println("contador depois de Incrementar:", c.n)
	fmt.Println(copiar(c))
	fmt.Printf("%d\n", "texto")
	fmt.Printf("%s %s\n", "falta um argumento")
	s := string(65)
	fmt.Println(s)
	x := 3
	x = x
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Add(1)
		defer wg.Done()
	}()
	wg.Wait()
	iniciarLost()
}
