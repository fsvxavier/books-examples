package caixa

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-11/topico-8/projeto-1/internal/modelo"
)

// Pesquisador responde a "tema" com "fatos".
type Pesquisador struct{ Modelo modelo.Modelo }

func (p Pesquisador) Tratar(ctx context.Context, m Mensagem) ([]Mensagem, error) {
	r, err := p.Modelo.Gerar(ctx, modelo.Pedido{Papel: "pesquisador", Instrucao: m.Corpo})
	if err != nil {
		return nil, err
	}
	return []Mensagem{{Conversa: m.Conversa, De: "pesquisador", Para: "redator", Tipo: "fatos", Corpo: r.Texto}}, nil
}

// Redator guarda os fatos de cada conversa e responde com "rascunho".
// Com uma só goroutine o mapa não precisaria de mutex; com várias, precisa.
type Redator struct {
	Modelo modelo.Modelo
	mu     sync.Mutex
	fatos  map[int][]string
}

func (r *Redator) Tratar(ctx context.Context, m Mensagem) ([]Mensagem, error) {
	r.mu.Lock()
	if r.fatos == nil {
		r.fatos = map[int][]string{}
	}
	apoio := r.fatos[m.Conversa]
	switch m.Tipo {
	case "fatos":
		apoio = strings.Split(m.Corpo, "\n")
		r.fatos[m.Conversa] = apoio
	case "critica":
		apoio = append(append([]string{}, apoio...), "CRITICA: "+m.Corpo)
	}
	r.mu.Unlock()
	resp, err := r.Modelo.Gerar(ctx, modelo.Pedido{Papel: "redator", Instrucao: "escreva", Contexto: apoio})
	if err != nil {
		return nil, err
	}
	return []Mensagem{{Conversa: m.Conversa, De: "redator", Para: "revisor", Tipo: "rascunho", Corpo: resp.Texto}}, nil
}

// Revisor devolve "critica" ao redator ou envia "pronto" para a saída.
// Desiste depois de MaxRodadas rodadas por conversa.
type Revisor struct {
	Modelo    modelo.Modelo
	MaxRodada int
	mu        sync.Mutex
	rodadas   map[int]int
}

func (v *Revisor) Tratar(ctx context.Context, m Mensagem) ([]Mensagem, error) {
	v.mu.Lock()
	if v.rodadas == nil {
		v.rodadas = map[int]int{}
	}
	v.rodadas[m.Conversa]++
	rodada := v.rodadas[m.Conversa]
	v.mu.Unlock()
	resp, err := v.Modelo.Gerar(ctx, modelo.Pedido{Papel: "revisor", Instrucao: "revise", Contexto: []string{m.Corpo}})
	if err != nil {
		return nil, err
	}
	switch {
	case resp.Texto == "APROVADO":
		return []Mensagem{{Conversa: m.Conversa, De: "revisor", Para: "saida", Tipo: "pronto", Corpo: m.Corpo}}, nil
	case rodada >= v.MaxRodada:
		return []Mensagem{{Conversa: m.Conversa, De: "revisor", Para: "saida", Tipo: "desistiu", Corpo: fmt.Sprint(rodada)}}, nil
	}
	return []Mensagem{{Conversa: m.Conversa, De: "revisor", Para: "redator", Tipo: "critica", Corpo: strings.TrimPrefix(resp.Texto, "REVISAR: ")}}, nil
}

// Montar liga os três agentes a um barramento com caixas do tamanho dado.
func Montar(m modelo.Modelo, capacidade, maxRodada int) *Barramento {
	return MontarComTrabalhadores(m, capacidade, maxRodada, 1)
}

// MontarComTrabalhadores é como Montar, com n goroutines por agente.
func MontarComTrabalhadores(m modelo.Modelo, capacidade, maxRodada, n int) *Barramento {
	b := Novo()
	b.Registrar("pesquisador", capacidade, Pesquisador{Modelo: m})
	b.Registrar("redator", capacidade, &Redator{Modelo: m})
	b.Registrar("revisor", capacidade, &Revisor{Modelo: m, MaxRodada: maxRodada})
	b.Registrar("saida", capacidade, nil)
	for _, nome := range []string{"pesquisador", "redator", "revisor"} {
		b.Trabalhadores(nome, n)
	}
	return b
}
