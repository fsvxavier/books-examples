package main

import (
	"errors"
	"fmt"
)

// ErrTransicao é o erro devolvido por qualquer forma da máquina quando a
// transição não é permitida.
var ErrTransicao = errors.New("transição inválida")

// Maquina é o que as quatro formas têm em comum.
type Maquina interface {
	Publicar() error
	Rejeitar() error
	Estado() string
}

// NovoDocSwitch cria o documento implementado com switch, em rascunho.
func NovoDocSwitch(admin bool) *DocSwitch { return &DocSwitch{estado: "rascunho", Admin: admin} }

func (d *DocSwitch) Estado() string { return d.estado }

type DocSwitch struct {
	estado string
	Admin  bool
}

func (d *DocSwitch) Publicar() error {
	switch d.estado {
	case "rascunho":
		d.estado = "moderacao"
	case "moderacao":
		if !d.Admin {
			return fmt.Errorf("publicar em moderacao sem ser admin: %w", ErrTransicao)
		}
		d.estado = "publicado"
	case "publicado":
		// já publicado: não faz nada
	}
	return nil
}

func (d *DocSwitch) Rejeitar() error {
	switch d.estado {
	case "rascunho":
		return fmt.Errorf("rejeitar em rascunho: %w", ErrTransicao)
	case "moderacao":
		d.estado = "rascunho"
	case "publicado":
		return fmt.Errorf("rejeitar em publicado: %w", ErrTransicao)
	}
	return nil
}

type Estado interface {
	Nome() string
	Publicar(d *Documento) error
	Rejeitar(d *Documento) error
}

type Documento struct {
	estado Estado
	Admin  bool
}

func NovoDocumento(admin bool) *Documento { return &Documento{estado: rascunho{}, Admin: admin} }

func (d *Documento) Estado() string  { return d.estado.Nome() }
func (d *Documento) Publicar() error { return d.estado.Publicar(d) }
func (d *Documento) Rejeitar() error { return d.estado.Rejeitar(d) }

type rascunho struct{}
type moderacao struct{}
type publicado struct{}

func (rascunho) Nome() string { return "rascunho" }
func (rascunho) Publicar(d *Documento) error {
	d.estado = moderacao{}
	return nil
}
func (rascunho) Rejeitar(*Documento) error {
	return fmt.Errorf("rejeitar em rascunho: %w", ErrTransicao)
}

func (moderacao) Nome() string { return "moderacao" }
func (moderacao) Publicar(d *Documento) error {
	if !d.Admin {
		return fmt.Errorf("publicar em moderacao sem ser admin: %w", ErrTransicao)
	}
	d.estado = publicado{}
	return nil
}
func (moderacao) Rejeitar(d *Documento) error {
	d.estado = rascunho{}
	return nil
}

func (publicado) Nome() string              { return "publicado" }
func (publicado) Publicar(*Documento) error { return nil }
func (publicado) Rejeitar(*Documento) error {
	return fmt.Errorf("rejeitar em publicado: %w", ErrTransicao)
}

type Est int

const (
	Rascunho Est = iota
	Moderacao
	Publicado
)

func (e Est) String() string { return [...]string{"rascunho", "moderacao", "publicado"}[e] }

type Evento int

const (
	EvPublicar Evento = iota
	EvRejeitar
)

func (e Evento) String() string { return [...]string{"publicar", "rejeitar"}[e] }

type chave struct {
	de Est
	ev Evento
}

type Regra struct {
	Para   Est
	Guarda func(admin bool) bool // nil = sempre permitido
}

var tabela = map[chave]Regra{
	{Rascunho, EvPublicar}:  {Para: Moderacao},
	{Moderacao, EvPublicar}: {Para: Publicado, Guarda: func(a bool) bool { return a }},
	{Moderacao, EvRejeitar}: {Para: Rascunho},
	{Publicado, EvPublicar}: {Para: Publicado},
}

type DocTabela struct {
	estado Est
	Admin  bool
}

func NovoDocTabela(admin bool) *DocTabela { return &DocTabela{Admin: admin} }

func (d *DocTabela) Estado() string  { return d.estado.String() }
func (d *DocTabela) Publicar() error { return d.Enviar(EvPublicar) }
func (d *DocTabela) Rejeitar() error { return d.Enviar(EvRejeitar) }

func (d *DocTabela) Enviar(ev Evento) error {
	r, ok := tabela[chave{d.estado, ev}]
	if !ok || (r.Guarda != nil && !r.Guarda(d.Admin)) {
		return fmt.Errorf("%s em %s: %w", ev, d.estado, ErrTransicao)
	}
	d.estado = r.Para
	return nil
}

type Fluxo struct {
	Admin   bool
	eventos []Evento
	nome    string
	Erros   []error
}

type fnEstado func(f *Fluxo) fnEstado

func (f *Fluxo) proximo() (Evento, bool) {
	if len(f.eventos) == 0 {
		return 0, false
	}
	ev := f.eventos[0]
	f.eventos = f.eventos[1:]
	return ev, true
}

func (f *Fluxo) recusar(ev Evento) {
	f.Erros = append(f.Erros, fmt.Errorf("%s em %s: %w", ev, f.nome, ErrTransicao))
}

func fnRascunho(f *Fluxo) fnEstado {
	f.nome = "rascunho"
	ev, ok := f.proximo()
	if !ok {
		return nil
	}
	if ev == EvPublicar {
		return fnModeracao
	}
	f.recusar(ev)
	return fnRascunho
}

func fnModeracao(f *Fluxo) fnEstado {
	f.nome = "moderacao"
	ev, ok := f.proximo()
	if !ok {
		return nil
	}
	switch {
	case ev == EvPublicar && f.Admin:
		return fnPublicado
	case ev == EvRejeitar:
		return fnRascunho
	}
	f.recusar(ev)
	return fnModeracao
}

func fnPublicado(f *Fluxo) fnEstado {
	f.nome = "publicado"
	ev, ok := f.proximo()
	if !ok {
		return nil
	}
	if ev == EvPublicar {
		return fnPublicado
	}
	f.recusar(ev)
	return fnPublicado
}

// Rodar consome todos os eventos e devolve o nome do estado final.
func (f *Fluxo) Rodar(eventos ...Evento) string {
	f.eventos = eventos
	for est := fnEstado(fnRascunho); est != nil; {
		est = est(f)
	}
	return f.nome
}

func descrever(m Maquina, nome string, eventos ...Evento) {
	fmt.Printf("%-10s", nome)
	for _, ev := range eventos {
		var err error
		if ev == EvPublicar {
			err = m.Publicar()
		} else {
			err = m.Rejeitar()
		}
		fmt.Printf(" %s->%s", ev, m.Estado())
		if err != nil {
			fmt.Print("(recusado)")
		}
	}
	fmt.Println()
}

func main() {
	roteiro := []Evento{EvPublicar, EvPublicar, EvRejeitar, EvPublicar, EvPublicar}
	fmt.Println("autor (sem admin):")
	descrever(NovoDocSwitch(false), "switch", roteiro...)
	descrever(NovoDocumento(false), "State", roteiro...)
	descrever(NovoDocTabela(false), "tabela", roteiro...)
	fmt.Println("administrador:")
	descrever(NovoDocSwitch(true), "switch", roteiro...)
	descrever(NovoDocumento(true), "State", roteiro...)
	descrever(NovoDocTabela(true), "tabela", roteiro...)

	f := &Fluxo{Admin: false}
	final := f.Rodar(roteiro...)
	fmt.Printf("%-10s final=%s, recusas=%d\n", "funções", final, len(f.Erros))
	for _, e := range f.Erros {
		fmt.Println("  ", e)
	}
}
