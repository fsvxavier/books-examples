package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"
)

// semHora remove o horario e deixa a saida reproduzivel no livro.
func semHora(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
}

type Usuario struct {
	ID    int
	Nome  string
	Senha string
}

// LogValue decide o que aparece no log quando um Usuario e registrado.
func (u Usuario) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("id", u.ID),
		slog.String("nome", u.Nome),
		// a senha nunca sai
	)
}

func main() {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: semHora}

	// 1. texto e JSON: mesma chamada, formatos diferentes
	txt := slog.New(slog.NewTextHandler(os.Stdout, opts))
	js := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	txt.Info("pedido criado", "pedido_id", 42, "total", 129.9, "cliente", "ana")
	js.Info("pedido criado", "pedido_id", 42, "total", 129.9, "cliente", "ana")

	// 2. niveis
	txt.Debug("detalhe interno", "passo", 3)
	txt.Warn("estoque baixo", "sku", "A1", "restante", 2)
	txt.Error("falha ao cobrar", "err", errors.New("cartao recusado"))

	// 3. atributos tipados (sem alocar o par chave-valor)
	js.LogAttrs(context.Background(), slog.LevelInfo, "tipado",
		slog.Int("n", 7), slog.Duration("tempo", 1500*time.Millisecond), slog.Bool("ok", true))

	// 4. With e grupos
	req := js.With("servico", "api", "versao", "1.4.0")
	req.Info("iniciando")
	req.WithGroup("http").Info("requisicao", "metodo", "GET", "status", 200)

	// 5. LogValuer esconde a senha
	js.Info("login", "usuario", Usuario{ID: 7, Nome: "ana", Senha: "segredo"})

	// 6. odd number of args: erro silencioso?
	js.Info("mal formado", "chave_sem_valor")

	// 7. nivel dinamico
	var nivel slog.LevelVar
	nivel.Set(slog.LevelWarn)
	din := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: &nivel, ReplaceAttr: semHora}))
	din.Info("nao aparece")
	nivel.Set(slog.LevelDebug)
	din.Debug("agora aparece")

	// 8. DiscardHandler (Go 1.24)
	slog.New(slog.DiscardHandler).Error("ninguem ve isto")
}
