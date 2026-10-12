// Package barramento entrega cada evento, na mesma goroutine e em ordem, aos
// manipuladores inscritos. É o barramento mais simples possível: em memória,
// síncrono e de um só processo.
package barramento

import (
	"context"
	"log/slog"
	"sync"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-6/topico-4/encurtador/internal/dominio"
)

// Manipulador reage a um evento. Um erro é registrado e não interrompe os
// demais: uma reação que falha não pode impedir as outras, nem desfazer o
// fato, que já aconteceu.
type Manipulador func(ctx context.Context, e dominio.Evento) error

// Barramento entrega eventos aos manipuladores inscritos.
type Barramento struct {
	log *slog.Logger

	mu            sync.RWMutex
	manipuladores []Manipulador
}

// Novo devolve um barramento sem inscritos.
func Novo(log *slog.Logger) *Barramento {
	return &Barramento{log: log}
}

// Inscrever acrescenta um manipulador. É chamada na composição, antes de o
// programa começar a atender.
func (b *Barramento) Inscrever(m Manipulador) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.manipuladores = append(b.manipuladores, m)
}

// Publicar entrega cada evento a cada manipulador, na ordem da inscrição.
func (b *Barramento) Publicar(ctx context.Context, eventos ...dominio.Evento) {
	b.mu.RLock()
	manipuladores := b.manipuladores
	b.mu.RUnlock()

	for _, e := range eventos {
		for _, m := range manipuladores {
			if err := m(ctx, e); err != nil {
				b.log.ErrorContext(ctx, "manipulador de evento falhou", "evento", e.Nome(), "erro", err)
			}
		}
	}
}
