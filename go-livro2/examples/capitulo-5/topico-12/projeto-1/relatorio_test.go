package main

import (
	"encoding/json"
	"strings"
	"testing"
)

var dados = Dados{Periodo: "2026-09", Pedidos: 128, Faturamento: 15432.9, Cancelados: 7, MaiorPedido: 980, Cliente: "Tom & Jerry <ltda>"}

var renderizadores = map[string]func() Renderizador{
	"texto": func() Renderizador { return &Texto{} },
	"json":  func() Renderizador { return &JSON{} },
	"html":  func() Renderizador { return &HTML{} },
}

var abstracoes = map[string]func(func() Renderizador) Relatorio{
	"resumo":    func(f func() Renderizador) Relatorio { return Resumo{f} },
	"detalhado": func(f func() Renderizador) Relatorio { return Detalhado{f} },
}

// A matriz abstracao x implementacao e testada com UMA tabela: 2 x 3 = 6 combinacoes.
func TestMatriz(t *testing.T) {
	for an, a := range abstracoes {
		for rn, r := range renderizadores {
			t.Run(an+"/"+rn, func(t *testing.T) {
				saida := a(r).Gerar(dados)
				if !strings.Contains(saida, "128") || !strings.Contains(saida, "15432.90") {
					t.Fatalf("faltam os numeros:\n%s", saida)
				}
				if an == "detalhado" && !strings.Contains(saida, "980.00") {
					t.Fatalf("detalhado sem maior pedido:\n%s", saida)
				}
				if an == "resumo" && strings.Contains(saida, "980.00") {
					t.Fatalf("resumo com detalhe demais:\n%s", saida)
				}
			})
		}
	}
}

func TestJSONEhValido(t *testing.T) {
	saida := Detalhado{renderizadores["json"]}.Gerar(dados)
	var v struct {
		Titulo string
		Itens  []map[string]string
	}
	if err := json.Unmarshal([]byte(saida), &v); err != nil {
		t.Fatalf("JSON invalido: %v\n%s", err, saida)
	}
	if len(v.Itens) != 5 {
		t.Fatalf("itens = %d", len(v.Itens))
	}
}

func TestHTMLEscapa(t *testing.T) {
	saida := Detalhado{renderizadores["html"]}.Gerar(dados)
	if strings.Contains(saida, "<ltda>") || !strings.Contains(saida, "Tom &amp; Jerry &lt;ltda&gt;") {
		t.Fatalf("HTML sem escape:\n%s", saida)
	}
}

// Um Renderizador falso testa a ABSTRACAO sem formatar nada.
type espiao struct{ chamadas []string }

func (e *espiao) Titulo(s string)   { e.chamadas = append(e.chamadas, "titulo:"+s) }
func (e *espiao) Item(k, v string)  { e.chamadas = append(e.chamadas, "item:"+k) }
func (e *espiao) Resultado() string { return strings.Join(e.chamadas, ",") }

func TestAbstracaoComEspiao(t *testing.T) {
	got := Resumo{func() Renderizador { return &espiao{} }}.Gerar(dados)
	want := "titulo:Resumo 2026-09,item:pedidos,item:faturamento"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
