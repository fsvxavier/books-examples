// Package consolida passa fatos da conversa (curto prazo) para a memória de
// longo prazo, em segundo plano e com vários trabalhadores.
package consolida

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-12/topico-10/projeto-1/internal/curto"
	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-12/topico-10/projeto-1/internal/janela"
	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-12/topico-10/projeto-1/internal/longo"
)

// Sessao identifica uma conversa terminada e de quem ela é.
type Sessao struct {
	Usuario  string
	Conversa string
}

// Relatorio resume o que a consolidação fez.
type Relatorio struct {
	Sessoes  int64
	Fatos    int64 // fatos extraídos
	Novos    int64 // gravados como memória nova
	Reforcos int64 // repetidos, só reforçaram uma memória existente
	Erros    int64
}

// Consolidador lê conversas do Valkey e grava memórias no PostgreSQL.
type Consolidador struct {
	Historico     *curto.Historico
	Loja          *longo.Loja
	Trabalhadores int
}

// Executar consolida todas as sessões, no máximo Trabalhadores por vez. Para
// quando o contexto acaba. Cada sessão é independente: a falha de uma não
// derruba as outras, e o relatório conta os erros.
func (c *Consolidador) Executar(ctx context.Context, sessoes []Sessao) (Relatorio, error) {
	var rel Relatorio
	fila := make(chan Sessao)
	var wg sync.WaitGroup
	for range max(c.Trabalhadores, 1) {
		wg.Go(func() {
			for s := range fila {
				if err := c.uma(ctx, s, &rel); err != nil {
					atomic.AddInt64(&rel.Erros, 1)
					_ = err // em produção: registrar com slog
				}
			}
		})
	}
enviar:
	for _, s := range sessoes {
		select {
		case fila <- s:
		case <-ctx.Done():
			break enviar
		}
	}
	close(fila)
	wg.Wait()
	return rel, ctx.Err()
}

func (c *Consolidador) uma(ctx context.Context, s Sessao, rel *Relatorio) error {
	msgs, err := c.Historico.Ler(ctx, s.Conversa)
	if err != nil {
		return err
	}
	atomic.AddInt64(&rel.Sessoes, 1)
	for _, m := range msgs {
		if m.Papel != janela.Usuario {
			continue
		}
		for _, f := range longo.Extrair(m.Texto) {
			atomic.AddInt64(&rel.Fatos, 1)
			_, novo, err := c.Loja.Guardar(ctx, s.Usuario, f.Tipo, f.Conteudo, f.Importancia)
			if err != nil {
				return fmt.Errorf("guardar para %s: %w", s.Usuario, err)
			}
			if novo {
				atomic.AddInt64(&rel.Novos, 1)
			} else {
				atomic.AddInt64(&rel.Reforcos, 1)
			}
		}
	}
	return nil
}
