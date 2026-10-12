package e03

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"testing"
	"testing/slogtest"
)

var ctx = context.Background()

func BenchmarkSlogJSON(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, nil))
	b.ReportAllocs()
	for b.Loop() {
		l.Info("pedido criado", "pedido_id", 42, "cliente", "ana", "total", 129.9)
	}
}

func BenchmarkSlogJSONLogAttrs(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, nil))
	b.ReportAllocs()
	for b.Loop() {
		l.LogAttrs(ctx, slog.LevelInfo, "pedido criado",
			slog.Int("pedido_id", 42), slog.String("cliente", "ana"), slog.Float64("total", 129.9))
	}
}

func BenchmarkSlogDesabilitado(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	b.ReportAllocs()
	for b.Loop() {
		l.Info("pedido criado", "pedido_id", 42, "cliente", "ana", "total", 129.9)
	}
}

func BenchmarkSlogDesabilitadoComCalculoCaro(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	b.ReportAllocs()
	for b.Loop() {
		l.Debug("estado", "dump", caro())
	}
}

func BenchmarkSlogDesabilitadoComEnabled(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	b.ReportAllocs()
	for b.Loop() {
		if l.Enabled(ctx, slog.LevelDebug) {
			l.Debug("estado", "dump", caro())
		}
	}
}

//go:noinline
func caro() string { return string(make([]byte, 256)) }

// nulo descarta a saida, mas NAO e io.Discard: o pacote log tem um atalho para
// io.Discard que pula a formatacao e tornaria a comparacao injusta.
type nulo struct{}

func (nulo) Write(p []byte) (int, error) { return len(p), nil }

func BenchmarkPacoteLog(b *testing.B) {
	l := log.New(nulo{}, "", log.LstdFlags)
	b.ReportAllocs()
	for b.Loop() {
		l.Printf("pedido criado pedido_id=%d cliente=%s total=%.1f", 42, "ana", 129.9)
	}
}

// slogtest verifica se um handler respeita o contrato do slog.
func TestHandlerJSONRespeitaContrato(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, nil)
	results := func() []map[string]any {
		var ms []map[string]any
		for _, linha := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(linha) == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(linha, &m); err != nil {
				t.Fatal(err)
			}
			ms = append(ms, m)
		}
		return ms
	}
	if err := slogtest.TestHandler(h, results); err != nil {
		t.Fatal(err)
	}
}
