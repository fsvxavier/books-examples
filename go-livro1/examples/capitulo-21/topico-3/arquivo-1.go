package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
)

type chave struct{}

// ComRequisicao guarda atributos de log no contexto.
func ComRequisicao(ctx context.Context, attrs ...slog.Attr) context.Context {
	atuais, _ := ctx.Value(chave{}).([]slog.Attr)
	return context.WithValue(ctx, chave{}, append(atuais[:len(atuais):len(atuais)], attrs...))
}

// ContextHandler acrescenta a cada registro os atributos guardados no contexto.
type ContextHandler struct{ slog.Handler }

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(chave{}).([]slog.Attr); ok {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs e WithGroup precisam devolver o NOSSO tipo, senao o contexto some
// a primeira vez que alguem chamar logger.With(...).
func (h ContextHandler) WithAttrs(a []slog.Attr) slog.Handler {
	return ContextHandler{h.Handler.WithAttrs(a)}
}
func (h ContextHandler) WithGroup(n string) slog.Handler {
	return ContextHandler{h.Handler.WithGroup(n)}
}

func semHora(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
}

func main() {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: semHora})
	log1 := slog.New(ContextHandler{base})

	ctx := ComRequisicao(context.Background(),
		slog.String("request_id", "req-123"), slog.String("usuario", "ana"))

	log1.InfoContext(ctx, "inicio")
	log1.With("componente", "estoque").InfoContext(ctx, "baixa de estoque", "sku", "A1")
	log1.Info("sem contexto") // Info sem Context nao ve os atributos

	// MultiHandler (Go 1.26): JSON completo para stdout, so erros para um "alerta"
	var alerta bytes.Buffer
	multi := slog.NewMultiHandler(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: semHora}),
		slog.NewTextHandler(&alerta, &slog.HandlerOptions{Level: slog.LevelError, ReplaceAttr: semHora}),
	)
	m := slog.New(multi)
	m.Info("rotina")
	m.Error("pagamento falhou", "pedido", 9)
	fmt.Printf("alerta recebeu: %q\n", strings.TrimSpace(alerta.String()))

	// AddSource mostra arquivo e linha
	src := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, ReplaceAttr: func(g []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		if a.Key == slog.SourceKey {
			s := a.Value.Any().(*slog.Source)
			s.File = s.File[strings.LastIndex(s.File, "/")+1:]
		}
		return a
	}}))
	src.Warn("com origem")

	// slog.SetDefault redireciona tambem o pacote log
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: semHora})))
	log.Printf("mensagem do pacote log antigo")
	slog.Info("mensagem do slog padrao")
}
