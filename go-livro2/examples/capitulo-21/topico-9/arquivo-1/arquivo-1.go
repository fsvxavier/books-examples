package main

import (
	"context"
	"fmt"
	"os"
	"runtime/trace"
	"sync"
	"time"
)

var (
	banco   sync.Mutex // recurso compartilhado: um gravador por vez
	gravado int
)

func validar(ctx context.Context, id int) {
	defer trace.StartRegion(ctx, "validar").End()
	x := 0
	for i := 0; i < 200_000; i++ { // CPU
		x += i % (id + 1)
	}
	_ = x
}

func gravar(ctx context.Context, id int) {
	defer trace.StartRegion(ctx, "gravar").End()
	banco.Lock() // fila: todos esperam o mesmo mutex
	defer banco.Unlock()
	time.Sleep(300 * time.Microsecond)
	gravado++
}

func notificar(ctx context.Context, id int) {
	defer trace.StartRegion(ctx, "notificar").End()
	time.Sleep(time.Millisecond) // espera de rede simulada
}

func atender(ctx context.Context, id int) {
	ctx, tarefa := trace.NewTask(ctx, "pedido")
	defer tarefa.End()
	trace.Logf(ctx, "pedido", "id=%d", id)
	validar(ctx, id)
	gravar(ctx, id)
	notificar(ctx, id)
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := trace.Start(f); err != nil {
		panic(err)
	}
	defer trace.Stop()

	pedidos := make(chan int)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for id := range pedidos {
				atender(context.Background(), id)
			}
		})
	}
	for id := 0; id < 400; id++ {
		pedidos <- id
	}
	close(pedidos)
	wg.Wait()
	fmt.Println("gravados:", gravado)
}
