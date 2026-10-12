// Package memoria implementa nucleo.Repositorio e nucleo.Contador guardando os
// dados em mapas. Serve para desenvolvimento, testes e instalações pequenas.
package memoria

import (
	"context"
	"sync"
	"time"

	"exemplos/livro4/capitulo-5/topico-3/encurtador/internal/nucleo"
	"exemplos/livro4/capitulo-5/topico-3/encurtador/internal/nucleo/dominio"
)

// registro é o estado de um link em tipos simples. O armazenamento guarda isto
// e nunca o agregado: o Link tem eventos pendentes e campos privados, e é
// reconstituído a cada leitura.
type registro struct {
	codigo, destino string
	criado, expira  time.Time
}

// Armazenamento guarda links e contadores em memória, seguro para uso concorrente.
type Armazenamento struct {
	mu     sync.RWMutex
	links  map[string]registro
	totais map[string]int
}

var _ nucleo.Armazenamento = (*Armazenamento)(nil)

// Novo devolve um armazenamento vazio.
func Novo() *Armazenamento {
	return &Armazenamento{links: make(map[string]registro), totais: make(map[string]int)}
}

// Salvar guarda o link, ou devolve dominio.ErrCodigoEmUso.
func (a *Armazenamento) Salvar(_ context.Context, l dominio.Link) error {
	expira, _ := l.Expira()
	r := registro{
		codigo: l.Codigo().String(), destino: l.Destino().String(),
		criado: l.Criado(), expira: expira,
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, existe := a.links[r.codigo]; existe {
		return dominio.ErrCodigoEmUso
	}
	a.links[r.codigo] = r
	return nil
}

// Buscar reconstitui o link, ou devolve dominio.ErrNaoEncontrado.
func (a *Armazenamento) Buscar(_ context.Context, codigo dominio.Codigo) (dominio.Link, error) {
	a.mu.RLock()
	r, ok := a.links[codigo.String()]
	a.mu.RUnlock()
	if !ok {
		return dominio.Link{}, dominio.ErrNaoEncontrado
	}
	return dominio.Reconstituir(r.codigo, r.destino, r.criado, r.expira)
}

// Incrementar soma um acesso ao código.
func (a *Armazenamento) Incrementar(_ context.Context, codigo dominio.Codigo) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.totais[codigo.String()]++
	return nil
}

// Total devolve os acessos do código.
func (a *Armazenamento) Total(_ context.Context, codigo dominio.Codigo) (int, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.totais[codigo.String()], nil
}
