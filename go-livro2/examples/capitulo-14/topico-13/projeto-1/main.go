package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type registro struct {
	mu  sync.Mutex
	ini time.Time
}

func (r *registro) log(f string, a ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Printf("  +%4dms %s\n", time.Since(r.ini).Milliseconds()/10*10, fmt.Sprintf(f, a...))
}

func main() {
	r := &registro{ini: time.Now()}

	// --- Componente 1: um "banco" que só precisa abrir e fechar.
	banco := Componente{
		Nome:    "banco",
		Iniciar: func(ctx context.Context) error { <-ctx.Done(); return nil },
		Parar: func(context.Context) error {
			time.Sleep(20 * time.Millisecond)
			r.log("banco: conexões fechadas")
			return nil
		},
	}

	// --- Componente 2: um worker que trata um job de 100 ms por vez.
	var wgWorker sync.WaitGroup
	worker := Componente{
		Nome: "worker",
		Iniciar: func(ctx context.Context) error {
			wgWorker.Add(1)
			defer wgWorker.Done()
			for n := 1; ; n++ {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(30 * time.Millisecond):
				}
				r.log("worker: começou o job %d", n)
				time.Sleep(100 * time.Millisecond) // o job NÃO é interrompido no meio
				r.log("worker: terminou o job %d", n)
			}
		},
		Parar: func(ctx context.Context) error {
			feito := make(chan struct{})
			go func() { wgWorker.Wait(); close(feito) }()
			select {
			case <-feito:
				r.log("worker: saiu depois de terminar o job em andamento")
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}

	// --- Componente 3: o servidor HTTP, com uma rota lenta (300 ms).
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	mux := http.NewServeMux()
	mux.HandleFunc("/lento", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		fmt.Fprint(w, "terminei")
	})
	srv := &http.Server{Handler: mux}
	servidor := Componente{
		Nome: "http",
		Iniciar: func(ctx context.Context) error {
			if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
		Parar: func(ctx context.Context) error {
			err := srv.Shutdown(ctx) // espera as requisições em andamento
			r.log("http: Shutdown devolveu %v", err)
			return err
		},
	}

	// O sinal vira o cancelamento do contexto raiz.
	ctx, parar := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer parar()

	// Um cliente com uma requisição lenta já em andamento quando o sinal chega.
	resposta := make(chan string, 1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		c, err := http.Get("http://" + ln.Addr().String() + "/lento")
		if err != nil {
			resposta <- "erro: " + err.Error()
			return
		}
		defer c.Body.Close()
		resposta <- fmt.Sprintf("resposta %d", c.StatusCode)
	}()
	// E o "operador", que manda SIGTERM ao processo depois de 120 ms.
	go func() {
		time.Sleep(120 * time.Millisecond)
		r.log("operador: kill -TERM")
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}()

	fmt.Println("1) encerramento ordenado por sinal")
	err := Orquestrar(ctx, 2*time.Second, r.log, banco, worker, servidor)
	fmt.Println("  Orquestrar devolveu:", err)
	fmt.Println("  cliente da requisição lenta:", <-resposta)

	// --- Cenário 2: um componente falha ao iniciar; os outros são parados.
	fmt.Println("2) um componente falha ao iniciar")
	r2 := &registro{ini: time.Now()}
	porta := Componente{
		Nome:    "porta",
		Iniciar: func(ctx context.Context) error { return errors.New("porta 8080 em uso") },
		Parar:   func(context.Context) error { return nil },
	}
	outro := Componente{
		Nome:    "cache",
		Iniciar: func(ctx context.Context) error { <-ctx.Done(); return nil },
		Parar:   func(context.Context) error { r2.log("cache: limpo"); return nil },
	}
	err = Orquestrar(context.Background(), time.Second, r2.log, outro, porta)
	fmt.Println("  Orquestrar devolveu:", err)
}
