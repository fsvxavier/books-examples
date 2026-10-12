package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func logSilencioso() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func chamar(h http.Handler) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/pedido", nil))
	return rec
}

func TestPedidoDentroDoPrazo(t *testing.T) {
	h := ComTraceID(geradorSequencial(), NovoHandler(500*time.Millisecond, 10*time.Millisecond, logSilencioso()))
	rec := chamar(h)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, quer 200", rec.Code)
	}
	if got := rec.Header().Get("X-Trace-ID"); got != "trace-1" {
		t.Errorf("cabeçalho X-Trace-ID = %q", got)
	}
	if !strings.Contains(rec.Body.String(), "trace-1") {
		t.Errorf("o corpo deveria citar o trace: %q", rec.Body.String())
	}
}

func TestPedidoExcedeOPrazo(t *testing.T) {
	h := ComTraceID(geradorSequencial(), NovoHandler(20*time.Millisecond, 2*time.Second, logSilencioso()))

	inicio := time.Now()
	rec := chamar(h)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("status = %d, quer 504", rec.Code)
	}
	if time.Since(inicio) > time.Second {
		t.Error("o handler não deveria esperar o trabalho inteiro: o contexto expira antes")
	}
}

func TestClienteCancelaARequisicao(t *testing.T) {
	h := ComTraceID(geradorSequencial(), NovoHandler(time.Minute, 2*time.Second, logSilencioso()))

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/pedido", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	feito := make(chan struct{})
	go func() { h.ServeHTTP(rec, req); close(feito) }()
	cancel() // o cliente desiste

	select {
	case <-feito:
	case <-time.After(time.Second):
		t.Fatal("o handler ignorou o cancelamento do cliente")
	}
	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("status = %d", rec.Code)
	}
}
