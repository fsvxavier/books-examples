package leak

import (
	"context"
	"time"
)

// Consultar simula uma chamada lenta que entrega o resultado por um canal sem
// buffer. A goroutine fica presa no envio quando o chamador desiste.
func Consultar(ctx context.Context, id int) int {
	ch := make(chan int) // sem buffer: o envio só termina se alguém receber
	go func() {
		time.Sleep(20 * time.Millisecond)
		ch <- id * 2 // RUIM: se o chamador desistiu, esta goroutine fica presa aqui
	}()
	select {
	case v := <-ch:
		return v
	case <-ctx.Done():
		return -1
	}
}

// ConsultarOk usa canal com buffer 1: o envio nunca bloqueia, mesmo sem receptor.
func ConsultarOk(ctx context.Context, id int) int {
	ch := make(chan int, 1)
	go func() {
		time.Sleep(20 * time.Millisecond)
		ch <- id * 2
	}()
	select {
	case v := <-ch:
		return v
	case <-ctx.Done():
		return -1
	}
}
