// Package auditoria traz o manipulador de eventos que registra cada fato do
// domínio em um log estruturado. Em vez de implementar uma porta própria, ele
// é UM DOS INSCRITOS no barramento: o núcleo não sabe que ele existe.
package auditoria

import (
	"context"
	"fmt"
	"log/slog"

	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/dominio"
)

// Manipulador devolve a função que escreve um registro por evento, no nível
// Info, com o campo evento. Um evento de tipo desconhecido é registrado só
// pelo nome, para que uma novidade do domínio nunca passe em branco.
func Manipulador(log *slog.Logger) func(context.Context, dominio.Evento) error {
	return func(ctx context.Context, e dominio.Evento) error {
		switch e := e.(type) {
		case dominio.LinkCriado:
			log.InfoContext(ctx, "auditoria", "evento", e.Nome(), "codigo", e.Codigo.String(), "destino", e.Destino.String())
		case dominio.LinkAcessado:
			log.InfoContext(ctx, "auditoria", "evento", e.Nome(), "codigo", e.Codigo.String())
		default:
			log.InfoContext(ctx, "auditoria", "evento", e.Nome(), "tipo", fmt.Sprintf("%T", e))
		}
		return nil
	}
}
