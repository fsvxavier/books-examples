package main

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

// ---------- Lado da IMPLEMENTACAO: como desenhar os primitivos ----------

// Renderizador sabe so os passos primitivos; nao sabe de "resumo" nem de "detalhado".
type Renderizador interface {
	Titulo(texto string)
	Item(chave, valor string)
	Resultado() string
}

type Texto struct{ b strings.Builder }

func (t *Texto) Titulo(s string)   { fmt.Fprintf(&t.b, "== %s ==\n", s) }
func (t *Texto) Item(k, v string)  { fmt.Fprintf(&t.b, "%s: %s\n", k, v) }
func (t *Texto) Resultado() string { return strings.TrimRight(t.b.String(), "\n") }

type JSON struct {
	titulo  string
	chaves  []string
	valores []string
}

func (j *JSON) Titulo(s string) { j.titulo = s }
func (j *JSON) Item(k, v string) {
	j.chaves = append(j.chaves, k)
	j.valores = append(j.valores, v)
}
func (j *JSON) Resultado() string {
	var b strings.Builder
	t, _ := json.Marshal(j.titulo)
	fmt.Fprintf(&b, `{"titulo":%s,"itens":[`, t)
	for i := range j.chaves {
		if i > 0 {
			b.WriteByte(',')
		}
		k, _ := json.Marshal(j.chaves[i])
		v, _ := json.Marshal(j.valores[i])
		fmt.Fprintf(&b, `{%s:%s}`, k, v)
	}
	b.WriteString("]}")
	return b.String()
}

type HTML struct{ b strings.Builder }

func (h *HTML) Titulo(s string) { fmt.Fprintf(&h.b, "<h1>%s</h1>\n<ul>\n", html.EscapeString(s)) }
func (h *HTML) Item(k, v string) {
	fmt.Fprintf(&h.b, "  <li>%s: %s</li>\n", html.EscapeString(k), html.EscapeString(v))
}
func (h *HTML) Resultado() string { return h.b.String() + "</ul>" }

// ---------- Lado da ABSTRACAO: o que o relatorio mostra ----------

type Dados struct {
	Periodo     string
	Pedidos     int
	Faturamento float64
	Cancelados  int
	MaiorPedido float64
	Cliente     string
}

// Relatorio e a abstracao de alto nivel; cada variante guarda um Renderizador (a ponte).
type Relatorio interface {
	Gerar(d Dados) string
}

type Resumo struct{ R func() Renderizador }

func (r Resumo) Gerar(d Dados) string {
	out := r.R()
	out.Titulo("Resumo " + d.Periodo)
	out.Item("pedidos", fmt.Sprint(d.Pedidos))
	out.Item("faturamento", fmt.Sprintf("R$ %.2f", d.Faturamento))
	return out.Resultado()
}

type Detalhado struct{ R func() Renderizador }

func (r Detalhado) Gerar(d Dados) string {
	out := r.R()
	out.Titulo("Detalhado " + d.Periodo)
	out.Item("pedidos", fmt.Sprint(d.Pedidos))
	out.Item("faturamento", fmt.Sprintf("R$ %.2f", d.Faturamento))
	out.Item("cancelados", fmt.Sprint(d.Cancelados))
	out.Item("maior pedido", fmt.Sprintf("R$ %.2f", d.MaiorPedido))
	out.Item("melhor cliente", d.Cliente)
	return out.Resultado()
}
