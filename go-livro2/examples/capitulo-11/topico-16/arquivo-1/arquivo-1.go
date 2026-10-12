package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func sinais() {
	fmt.Println("== signal.Notify: o canal é a inscrição ==")
	// Dois assinantes independentes para o mesmo sinal.
	a := make(chan os.Signal, 1)
	b := make(chan os.Signal, 1)
	signal.Notify(a, syscall.SIGUSR1)
	signal.Notify(b, syscall.SIGUSR1)

	// Três sinais seguidos, antes de qualquer leitura.
	for range 3 {
		_ = syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println("a guardou:", len(a), "| b guardou:", len(b), "(buffer 1: o resto foi descartado)")

	<-a
	<-b
	signal.Stop(a) // cancela só a inscrição de a
	_ = syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("depois de Stop(a): a recebeu", len(a), "| b recebeu", len(b))
	signal.Stop(b)
}

func afterFunc() {
	fmt.Println("== context.AfterFunc: observar o cancelamento ==")
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	stop := context.AfterFunc(ctx, func() {
		fmt.Println("tratador rodou: ctx cancelado, err =", ctx.Err())
		wg.Done()
	})
	_ = stop
	cancel()
	wg.Wait()

	ctx2, cancel2 := context.WithCancel(context.Background())
	parou := context.AfterFunc(ctx2, func() { fmt.Println("NÃO deveria rodar") })
	fmt.Println("stop() antes do cancelamento devolve:", parou())
	cancel2()
	time.Sleep(50 * time.Millisecond)
	fmt.Println("stop() de novo devolve:", parou())
}

func desligamento() {
	fmt.Println("== Server.RegisterOnShutdown: ganchos de desligamento ==")
	srv := &http.Server{Addr: "127.0.0.1:0"}
	var wg sync.WaitGroup
	for _, nome := range []string{"fecha-websockets", "descarrega-cache"} {
		wg.Add(1)
		srv.RegisterOnShutdown(func() {
			defer wg.Done()
			fmt.Println("gancho rodou:", nome)
		})
	}
	_ = srv.Shutdown(context.Background())
	wg.Wait()
}

func broadcast() {
	fmt.Println("== sync.Cond.Broadcast: acordar todos ==")
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	pronto := false
	var wg sync.WaitGroup
	acordou := make(chan int, 3)
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			for !pronto {
				cond.Wait()
			}
			mu.Unlock()
			acordou <- i
		}()
	}
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	pronto = true
	mu.Unlock()
	cond.Broadcast()
	wg.Wait()
	close(acordou)
	soma := 0
	for i := range acordou {
		soma += i
	}
	fmt.Println("3 goroutines acordaram, soma dos ids =", soma)
}

func main() {
	sinais()
	afterFunc()
	desligamento()
	broadcast()
}
