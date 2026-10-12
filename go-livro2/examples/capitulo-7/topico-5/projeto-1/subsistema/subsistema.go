// Package subsistema simula cinco servicos independentes, cada um com a sua API.
package subsistema

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ---------- Estoque ----------

type Estoque struct {
	mu     sync.Mutex
	saldos map[string]int
}

func NovoEstoque(saldos map[string]int) *Estoque { return &Estoque{saldos: saldos} }

var ErrSemEstoque = errors.New("sem estoque")

func (e *Estoque) Reservar(sku string, qtd int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.saldos[sku] < qtd {
		return fmt.Errorf("%w: %s (pedido %d, saldo %d)", ErrSemEstoque, sku, qtd, e.saldos[sku])
	}
	e.saldos[sku] -= qtd
	return nil
}

func (e *Estoque) Liberar(sku string, qtd int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.saldos[sku] += qtd
}

func (e *Estoque) Saldo(sku string) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.saldos[sku]
}

// ---------- Frete ----------

type Frete struct{}

func (Frete) Calcular(cep string, itens int) (centavos int64, err error) {
	if len(cep) != 8 {
		return 0, errors.New("CEP invalido")
	}
	return 1500 + int64(itens)*200, nil
}

// ---------- Pagamento ----------

type Pagamento struct {
	mu        sync.Mutex
	Recusar   bool
	cobrancas map[string]int64
	proximo   int
}

func (p *Pagamento) Cobrar(ctx context.Context, cartao string, centavos int64) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if p.Recusar {
		return "", errors.New("cartao recusado")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cobrancas == nil {
		p.cobrancas = map[string]int64{}
	}
	p.proximo++
	id := fmt.Sprintf("PG-%03d", p.proximo)
	p.cobrancas[id] = centavos
	return id, nil
}

func (p *Pagamento) Estornar(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.cobrancas[id]; !ok {
		return fmt.Errorf("cobranca %s inexistente", id)
	}
	delete(p.cobrancas, id)
	return nil
}

func (p *Pagamento) Ativas() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.cobrancas)
}

// ---------- Fiscal ----------

type Fiscal struct{ Fora bool }

func (f Fiscal) EmitirNota(cliente string, centavos int64) (string, error) {
	if f.Fora {
		return "", errors.New("SEFAZ indisponivel")
	}
	return fmt.Sprintf("NF-%s-%d", cliente, centavos), nil
}

// ---------- Notificacao ----------

type Notificacao struct {
	mu     sync.Mutex
	Falhar bool
	Enviou []string
}

func (n *Notificacao) Avisar(destino, texto string) error {
	if n.Falhar {
		return errors.New("servico de e-mail fora do ar")
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Enviou = append(n.Enviou, destino+": "+texto)
	return nil
}
