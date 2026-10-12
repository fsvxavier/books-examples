package main

import (
	"errors"
	"fmt"
	"slices"
	"sync"
)

// Item é uma linha do orçamento; valores em centavos.
type Item struct {
	Nome  string
	Valor int
}

// Tipo diz o que causou a mudança.
type Tipo int

const (
	Adicionado Tipo = iota
	Removido
	Restaurado
	EmLote
)

func (t Tipo) String() string {
	return [...]string{"adicionado", "removido", "restaurado", "em lote"}[t]
}

// Mudanca é o evento publicado. Versao cresce a cada mudança, inclusive
// as restaurações: ela numera EVENTOS, não estados.
type Mudanca struct {
	Tipo   Tipo
	Versao uint64
	Itens  int
	Total  int
}

// Foto é o memento do orçamento: só os itens, nunca a versão.
type Foto struct{ itens []Item }

var ErrNaoEncontrado = errors.New("item não encontrado")

// Orcamento é, ao mesmo tempo, o originador do Memento e o sujeito do Observer.
type Orcamento struct {
	mu     sync.Mutex
	itens  []Item
	versao uint64
	subs   subs
}

type subs struct {
	mu    sync.Mutex
	prox  int
	lista []sub
}

type sub struct {
	id int
	f  func(Mudanca)
}

func (s *subs) inscrever(f func(Mudanca)) func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.prox
	s.prox++
	s.lista = append(s.lista, sub{id, f})
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.lista = slices.DeleteFunc(s.lista, func(x sub) bool { return x.id == id })
	}
}

func (s *subs) publicar(m Mudanca) {
	s.mu.Lock()
	copia := slices.Clone(s.lista)
	s.mu.Unlock()
	for _, x := range copia {
		func() {
			defer func() { _ = recover() }() // um assinante ruim não derruba os outros
			x.f(m)
		}()
	}
}

// Observar inscreve um assinante e devolve o cancelamento.
func (o *Orcamento) Observar(f func(Mudanca)) (cancelar func()) { return o.subs.inscrever(f) }

func total(itens []Item) int {
	t := 0
	for _, i := range itens {
		t += i.Valor
	}
	return t
}

// aplicar muda o estado sob a trava e publica FORA dela.
func (o *Orcamento) aplicar(tipo Tipo, f func(itens []Item) ([]Item, error)) error {
	o.mu.Lock()
	novo, err := f(slices.Clone(o.itens)) // f trabalha numa cópia: erro = nada mudou
	if err != nil {
		o.mu.Unlock()
		return err
	}
	o.itens = novo
	o.versao++
	m := Mudanca{Tipo: tipo, Versao: o.versao, Itens: len(novo), Total: total(novo)}
	o.mu.Unlock()
	o.subs.publicar(m)
	return nil
}

func (o *Orcamento) Adicionar(nome string, valor int) error {
	return o.aplicar(Adicionado, func(it []Item) ([]Item, error) {
		return append(it, Item{nome, valor}), nil
	})
}

func (o *Orcamento) Remover(nome string) error {
	return o.aplicar(Removido, func(it []Item) ([]Item, error) {
		i := slices.IndexFunc(it, func(x Item) bool { return x.Nome == nome })
		if i < 0 {
			return nil, fmt.Errorf("remover %q: %w", nome, ErrNaoEncontrado)
		}
		return slices.Delete(it, i, i+1), nil
	})
}

// Lote aplica várias operações como uma só: um único evento no fim, ou
// nenhum (e nenhuma mudança) se alguma falhar.
func (o *Orcamento) Lote(ops ...func(*[]Item) error) error {
	return o.aplicar(EmLote, func(it []Item) ([]Item, error) {
		for i, op := range ops {
			if err := op(&it); err != nil {
				return nil, fmt.Errorf("operação %d do lote: %w", i+1, err)
			}
		}
		return it, nil
	})
}

// Salvar tira a foto do estado atual.
func (o *Orcamento) Salvar() Foto {
	o.mu.Lock()
	defer o.mu.Unlock()
	return Foto{slices.Clone(o.itens)}
}

// Restaurar volta aos itens da foto E avisa os observadores: restaurar
// também é uma mudança de estado.
func (o *Orcamento) Restaurar(f Foto) {
	_ = o.aplicar(Restaurado, func([]Item) ([]Item, error) {
		return slices.Clone(f.itens), nil
	})
}

// RestaurarEmSilencio é o erro clássico: muda o estado sem avisar ninguém.
func (o *Orcamento) RestaurarEmSilencio(f Foto) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.itens = slices.Clone(f.itens)
}

func (o *Orcamento) Total() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return total(o.itens)
}

// Sessao é o guardião: pilhas de desfazer e refazer sobre um Orcamento.
type Sessao struct {
	*Orcamento
	desfazer []Foto
	refazer  []Foto
}

func NovaSessao(o *Orcamento) *Sessao { return &Sessao{Orcamento: o} }

func (s *Sessao) registrar() {
	s.desfazer = append(s.desfazer, s.Salvar())
	s.refazer = nil
}

func (s *Sessao) Adicionar(nome string, valor int) error {
	s.registrar()
	err := s.Orcamento.Adicionar(nome, valor)
	if err != nil {
		s.desfazer = s.desfazer[:len(s.desfazer)-1] // nada mudou: não guarda
	}
	return err
}

func (s *Sessao) Remover(nome string) error {
	s.registrar()
	err := s.Orcamento.Remover(nome)
	if err != nil {
		s.desfazer = s.desfazer[:len(s.desfazer)-1]
	}
	return err
}

func (s *Sessao) Desfazer() bool {
	if len(s.desfazer) == 0 {
		return false
	}
	s.refazer = append(s.refazer, s.Salvar())
	n := len(s.desfazer) - 1
	f := s.desfazer[n]
	s.desfazer = s.desfazer[:n]
	s.Restaurar(f)
	return true
}

func (s *Sessao) Refazer() bool {
	if len(s.refazer) == 0 {
		return false
	}
	s.desfazer = append(s.desfazer, s.Salvar())
	n := len(s.refazer) - 1
	f := s.refazer[n]
	s.refazer = s.refazer[:n]
	s.Restaurar(f)
	return true
}
