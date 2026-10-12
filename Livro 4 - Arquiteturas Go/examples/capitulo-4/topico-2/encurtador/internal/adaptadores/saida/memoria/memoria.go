// Package memoria implementa nucleo.Repositorio guardando os links em um
// mapa. Serve para desenvolvimento, testes e instalações pequenas.
package memoria

import (
	"context"
	"sync"

	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

// Repositorio guarda os links em memória e é seguro para uso concorrente.
type Repositorio struct {
	mu    sync.RWMutex
	links map[string]*dominio.Link
}

// Novo devolve um repositório vazio.
func Novo() *Repositorio {
	return &Repositorio{links: make(map[string]*dominio.Link)}
}

// Salvar guarda o link, ou devolve dominio.ErrCodigoEmUso.
func (r *Repositorio) Salvar(_ context.Context, l dominio.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, existe := r.links[l.Codigo]; existe {
		return dominio.ErrCodigoEmUso
	}
	r.links[l.Codigo] = &l
	return nil
}

// Buscar devolve uma cópia do link, ou dominio.ErrNaoEncontrado.
func (r *Repositorio) Buscar(_ context.Context, codigo string) (dominio.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	l, ok := r.links[codigo]
	if !ok {
		return dominio.Link{}, dominio.ErrNaoEncontrado
	}
	return *l, nil
}

// RegistrarAcesso soma um acesso e devolve o link já atualizado.
func (r *Repositorio) RegistrarAcesso(_ context.Context, codigo string) (dominio.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.links[codigo]
	if !ok {
		return dominio.Link{}, dominio.ErrNaoEncontrado
	}
	l.Acessos++
	return *l, nil
}
