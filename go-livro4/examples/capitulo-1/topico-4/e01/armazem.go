package main

import (
	"errors"
	"sync"
	"time"
)

var (
	errNaoEncontrado = errors.New("link não encontrado")
	errCodigoEmUso   = errors.New("código já está em uso")
)

// link é o registro guardado para cada endereço encurtado.
type link struct {
	Codigo  string    `json:"codigo"`
	Destino string    `json:"destino"`
	Criado  time.Time `json:"criado"`
	Acessos int       `json:"acessos"`
}

// armazem guarda os links em memória, protegido por um mutex.
type armazem struct {
	mu    sync.RWMutex
	links map[string]*link
}

func novoArmazem() *armazem {
	return &armazem{links: make(map[string]*link)}
}

func (a *armazem) salvar(l link) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, existe := a.links[l.Codigo]; existe {
		return errCodigoEmUso
	}
	a.links[l.Codigo] = &l
	return nil
}

func (a *armazem) buscar(codigo string) (link, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	l, ok := a.links[codigo]
	if !ok {
		return link{}, errNaoEncontrado
	}
	return *l, nil
}

// registrarAcesso soma um acesso e devolve o link já atualizado.
func (a *armazem) registrarAcesso(codigo string) (link, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	l, ok := a.links[codigo]
	if !ok {
		return link{}, errNaoEncontrado
	}
	l.Acessos++
	return *l, nil
}
