// Package auditoria implementa nucleo.Auditoria registrando cada evento em um
// log estruturado. É um adaptador secundário de NOTIFICAÇÃO: o núcleo avisa,
// o adaptador decide o que fazer com o aviso.
package auditoria

import (
	"context"
	"log/slog"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

// Slog escreve um registro por evento, no nível Info, com o campo evento.
type Slog struct {
	log *slog.Logger
}

// NovoSlog devolve uma auditoria que escreve no logger dado.
func NovoSlog(log *slog.Logger) *Slog {
	return &Slog{log: log}
}

// LinkCriado registra a criação de um link.
func (a *Slog) LinkCriado(ctx context.Context, l dominio.Link) {
	a.log.InfoContext(ctx, "auditoria", "evento", "link_criado", "codigo", l.Codigo, "destino", l.Destino)
}

// LinkAcessado registra um acesso e o total acumulado.
func (a *Slog) LinkAcessado(ctx context.Context, l dominio.Link) {
	a.log.InfoContext(ctx, "auditoria", "evento", "link_acessado", "codigo", l.Codigo, "acessos", l.Acessos)
}

// Descartar é uma auditoria que não faz nada, para quem não quer auditar.
type Descartar struct{}

// LinkCriado não faz nada.
func (Descartar) LinkCriado(context.Context, dominio.Link) {}

// LinkAcessado não faz nada.
func (Descartar) LinkAcessado(context.Context, dominio.Link) {}
