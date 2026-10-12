package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type chaveTraceID struct{}

// TraceID devolve o identificador de rastreamento guardado no contexto.
func TraceID(ctx context.Context) string {
	id, _ := ctx.Value(chaveTraceID{}).(string)
	return id
}

// ComTraceID é um middleware: gera um ID por requisição e o coloca no contexto.
// O gerador é injetado, para os testes poderem usar IDs previsíveis.
func ComTraceID(gerar func() string, proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := gerar()
		w.Header().Set("X-Trace-ID", id)
		ctx := context.WithValue(r.Context(), chaveTraceID{}, id)
		proximo.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Pedido simula o trabalho do serviço, respeitando o contexto.
func Pedido(ctx context.Context, duracao time.Duration) error {
	select {
	case <-time.After(duracao):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// NovoHandler cria o handler de /pedido com um prazo máximo por requisição.
func NovoHandler(prazo, duracaoDoTrabalho time.Duration, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), prazo)
		defer cancel()

		id := TraceID(ctx)
		if err := Pedido(ctx, duracaoDoTrabalho); err != nil {
			log.Warn("pedido não concluído", "trace", id, "erro", err)
			http.Error(w, "tempo limite excedido", http.StatusGatewayTimeout)
			return
		}
		fmt.Fprintf(w, "pedido processado (trace: %s)\n", id)
	})
}

// geradorSequencial produz IDs "trace-1", "trace-2"... (seguro para uso concorrente).
func geradorSequencial() func() string {
	var n atomic.Int64
	return func() string { return "trace-" + strconv.FormatInt(n.Add(1), 10) }
}
