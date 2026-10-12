package main

import (
	"fmt"
	"strings"
)

// Mediador e a interface que os COMPONENTES conhecem: um unico metodo de notificacao.
type Mediador interface {
	Notificar(origem Componente, evento string)
}

type Componente interface {
	Nome() string
	DefinirMediador(Mediador)
}

// base guarda o mediador (a unica dependencia de um componente).
type base struct {
	nome string
	med  Mediador
}

func (b *base) Nome() string               { return b.nome }
func (b *base) DefinirMediador(m Mediador) { b.med = m }
func (b *base) avisar(origem Componente, evento string) {
	if b.med != nil {
		b.med.Notificar(origem, evento)
	}
}

// ---------- Componentes: so conhecem o mediador, NUNCA outros componentes ----------

type Caixa struct {
	base
	Marcada bool
}

func NovaCaixa(nome string) *Caixa { return &Caixa{base: base{nome: nome}} }

func (c *Caixa) Alternar() {
	c.Marcada = !c.Marcada
	c.avisar(c, "alternou")
}

type Campo struct {
	base
	Texto   string
	Visivel bool
}

func NovoCampo(nome string) *Campo { return &Campo{base: base{nome: nome}, Visivel: true} }

func (c *Campo) Digitar(t string) {
	c.Texto = t
	c.avisar(c, "digitou")
}

type Botao struct {
	base
	Habilitado bool
}

func NovoBotao(nome string) *Botao { return &Botao{base: base{nome: nome}, Habilitado: true} }

func (b *Botao) Clicar() {
	if b.Habilitado {
		b.avisar(b, "clicou")
	}
}

// ---------- Mediador concreto: a UNICA classe que conhece as relacoes entre os componentes ----------

type DialogoPerfil struct {
	TemCachorro *Caixa
	NomeCliente *Campo
	NomeCao     *Campo
	Salvar      *Botao
	Saida       []string
}

func NovoDialogoPerfil() *DialogoPerfil {
	d := &DialogoPerfil{
		TemCachorro: NovaCaixa("tem cachorro"),
		NomeCliente: NovoCampo("nome do cliente"),
		NomeCao:     NovoCampo("nome do cachorro"),
		Salvar:      NovoBotao("salvar"),
	}
	d.NomeCao.Visivel = false
	d.Salvar.Habilitado = false
	for _, c := range []Componente{d.TemCachorro, d.NomeCliente, d.NomeCao, d.Salvar} {
		c.DefinirMediador(d) // a ligacao e feita em UM lugar
	}
	return d
}

func (d *DialogoPerfil) Notificar(origem Componente, evento string) {
	switch {
	case origem == Componente(d.TemCachorro) && evento == "alternou":
		d.NomeCao.Visivel = d.TemCachorro.Marcada // marcar a caixa revela o campo
		if !d.TemCachorro.Marcada {
			d.NomeCao.Texto = ""
		}
		d.atualizarSalvar()
	case (origem == Componente(d.NomeCliente) || origem == Componente(d.NomeCao)) && evento == "digitou":
		d.atualizarSalvar()
	case origem == Componente(d.Salvar) && evento == "clicou":
		d.Saida = append(d.Saida, fmt.Sprintf("salvo: cliente=%q cachorro=%q", d.NomeCliente.Texto, d.NomeCao.Texto))
	}
}

// atualizarSalvar centraliza a REGRA de validacao: o botao so habilita se tudo estiver preenchido.
func (d *DialogoPerfil) atualizarSalvar() {
	ok := strings.TrimSpace(d.NomeCliente.Texto) != ""
	if d.TemCachorro.Marcada && strings.TrimSpace(d.NomeCao.Texto) == "" {
		ok = false
	}
	d.Salvar.Habilitado = ok
}
