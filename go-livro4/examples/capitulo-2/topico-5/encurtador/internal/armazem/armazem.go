package armazem

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNaoEncontrado = errors.New("link não encontrado")
	ErrCodigoEmUso   = errors.New("código já está em uso")
)

// link é o registro guardado para cada endereço encurtado.
type Link struct {
	Codigo  string    `json:"codigo"`
	Destino string    `json:"destino"`
	Criado  time.Time `json:"criado"`
	Acessos int       `json:"acessos"`
}

// armazem guarda os links em memória, protegido por um mutex.
type Armazem struct {
	mu    sync.RWMutex
	links map[string]*Link
}

func Novo() *Armazem {
	return &Armazem{links: make(map[string]*Link)}
}

func (a *Armazem) Salvar(l Link) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, existe := a.links[l.Codigo]; existe {
		return ErrCodigoEmUso
	}
	a.links[l.Codigo] = &l
	return nil
}

func (a *Armazem) Buscar(codigo string) (Link, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	l, ok := a.links[codigo]
	if !ok {
		return Link{}, ErrNaoEncontrado
	}
	return *l, nil
}

// registrarAcesso soma um acesso e devolve o link já atualizado.
func (a *Armazem) RegistrarAcesso(codigo string) (Link, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	l, ok := a.links[codigo]
	if !ok {
		return Link{}, ErrNaoEncontrado
	}
	l.Acessos++
	return *l, nil
}
