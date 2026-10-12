package main

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
)

// ErrTransicao é o erro-base de qualquer recusa da máquina.
var ErrTransicao = errors.New("transição recusada")

// ErroTransicao diz de qual estado, com qual evento e por quê.
type ErroTransicao[S, E comparable] struct {
	De     S
	Evento E
	Motivo string
}

func (e *ErroTransicao[S, E]) Error() string {
	return fmt.Sprintf("evento %v em %v recusado: %s", e.Evento, e.De, e.Motivo)
}

func (e *ErroTransicao[S, E]) Is(alvo error) bool { return alvo == ErrTransicao }

type par[S, E comparable] struct {
	de S
	ev E
}

// Regra descreve uma transição permitida.
type Regra[S comparable] struct {
	Para   S
	Guarda func() bool  // nil: sempre permitido
	Acao   func() error // roda antes de confirmar; se falhar, nada muda
}

// Mudanca é o que os observadores recebem.
type Mudanca[S, E comparable] struct {
	De     S
	Evento E
	Para   S
	Versao uint64
}

// Maquina é uma máquina de estados finita, segura para uso concorrente.
type Maquina[S, E comparable] struct {
	mu         sync.Mutex
	atual      S
	versao     uint64
	regras     map[par[S, E]]Regra[S]
	proibidos  map[par[S, E]]bool
	observores []func(Mudanca[S, E])
}

func Nova[S, E comparable](inicial S) *Maquina[S, E] {
	return &Maquina[S, E]{
		atual:     inicial,
		regras:    map[par[S, E]]Regra[S]{},
		proibidos: map[par[S, E]]bool{},
	}
}

// Permitir registra uma transição.
func (m *Maquina[S, E]) Permitir(de S, ev E, r Regra[S]) *Maquina[S, E] {
	m.regras[par[S, E]{de, ev}] = r
	return m
}

// Proibir registra, de forma explícita, que o par NÃO é permitido.
func (m *Maquina[S, E]) Proibir(de S, ev E) *Maquina[S, E] {
	m.proibidos[par[S, E]{de, ev}] = true
	return m
}

func (m *Maquina[S, E]) Observar(f func(Mudanca[S, E])) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observores = append(m.observores, f)
}

func (m *Maquina[S, E]) Estado() S {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.atual
}

// Enviar tenta a transição. A checagem, a ação e a troca de estado
// acontecem sob a mesma trava: o evento é atômico. Os observadores são
// chamados depois, fora da trava.
func (m *Maquina[S, E]) Enviar(ev E) error {
	m.mu.Lock()
	r, ok := m.regras[par[S, E]{m.atual, ev}]
	if !ok {
		de := m.atual
		motivo := "sem regra para este par"
		if m.proibidos[par[S, E]{de, ev}] {
			motivo = "proibido neste estado"
		}
		m.mu.Unlock()
		return &ErroTransicao[S, E]{de, ev, motivo}
	}
	if r.Guarda != nil && !r.Guarda() {
		de := m.atual
		m.mu.Unlock()
		return &ErroTransicao[S, E]{de, ev, "a guarda recusou"}
	}
	if r.Acao != nil {
		if err := r.Acao(); err != nil {
			de := m.atual
			m.mu.Unlock()
			return fmt.Errorf("ação de %v em %v: %w", ev, de, err)
		}
	}
	mud := Mudanca[S, E]{De: m.atual, Evento: ev, Para: r.Para}
	m.atual = r.Para
	m.versao++
	mud.Versao = m.versao
	obs := slices.Clone(m.observores)
	m.mu.Unlock()

	for _, f := range obs {
		f(mud)
	}
	return nil
}

// Lacunas devolve os pares (estado, evento) que não são nem permitidos
// nem proibidos de forma explícita: são, quase sempre, esquecimentos.
func (m *Maquina[S, E]) Lacunas(estados []S, eventos []E) []string {
	var l []string
	for _, s := range estados {
		for _, e := range eventos {
			p := par[S, E]{s, e}
			if _, ok := m.regras[p]; ok {
				continue
			}
			if m.proibidos[p] {
				continue
			}
			l = append(l, fmt.Sprintf("%v + %v", s, e))
		}
	}
	return l
}

// Mermaid exporta as regras como um diagrama de estados.
func (m *Maquina[S, E]) Mermaid() string {
	var linhas []string
	for p, r := range m.regras {
		rot := fmt.Sprint(p.ev)
		if r.Guarda != nil {
			rot += " [guarda]"
		}
		linhas = append(linhas, fmt.Sprintf("    %v --> %v: %s", p.de, r.Para, rot))
	}
	sort.Strings(linhas)
	return "stateDiagram-v2\n" + strings.Join(linhas, "\n")
}
