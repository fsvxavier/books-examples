package pedido

import (
	"errors"
	"fmt"
	"sort"
)

// ---- ANTES: uma struct que faz tudo (viola o SRP) ----

type PedidoGrande struct {
	Cliente string
	Itens   map[string]int // sku -> quantidade
	Total   int64
}

var precos = map[string]int64{"A": 1000, "B": 2500}

// Fechar valida, calcula, "grava" e "avisa": quatro motivos para mudar.
func (p *PedidoGrande) Fechar(log *[]string) error {
	if p.Cliente == "" || len(p.Itens) == 0 { // regra de validacao
		return errors.New("pedido invalido")
	}
	var total int64 // regra de precificacao
	for sku, q := range p.Itens {
		total += precos[sku] * int64(q)
	}
	if total > 10_000 {
		total = total * 90 / 100
	}
	p.Total = total
	*log = append(*log, fmt.Sprintf("INSERT pedido cliente=%s total=%d", p.Cliente, total)) // persistencia
	*log = append(*log, fmt.Sprintf("EMAIL para %s: pedido de %d", p.Cliente, total))       // notificacao
	return nil
}

// ---- DEPOIS: cada responsabilidade em seu lugar ----

type Pedido struct {
	Cliente string
	Itens   map[string]int
}

func (p Pedido) Validar() error {
	if p.Cliente == "" || len(p.Itens) == 0 {
		return errors.New("pedido invalido")
	}
	return nil
}

// Precificador so conhece preco e desconto.
type Precificador struct{ Precos map[string]int64 }

func (pr Precificador) Total(p Pedido) int64 {
	skus := make([]string, 0, len(p.Itens))
	for s := range p.Itens {
		skus = append(skus, s)
	}
	sort.Strings(skus)
	var total int64
	for _, s := range skus {
		total += pr.Precos[s] * int64(p.Itens[s])
	}
	if total > 10_000 {
		total = total * 90 / 100
	}
	return total
}

// Portas: o servico depende de comportamentos, nao de bancos nem de SMTP.
type Repositorio interface {
	Salvar(p Pedido, total int64) error
}
type Notificador interface {
	Avisar(cliente string, total int64) error
}

type Servico struct {
	Preco Precificador
	Repo  Repositorio
	Aviso Notificador
}

func (s Servico) Fechar(p Pedido) (int64, error) {
	if err := p.Validar(); err != nil {
		return 0, err
	}
	total := s.Preco.Total(p)
	if err := s.Repo.Salvar(p, total); err != nil {
		return 0, fmt.Errorf("salvar: %w", err)
	}
	// falha ao avisar nao desfaz o pedido: decisao de negocio explicita
	_ = s.Aviso.Avisar(p.Cliente, total)
	return total, nil
}
