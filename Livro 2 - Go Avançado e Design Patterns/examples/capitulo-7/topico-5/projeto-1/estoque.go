package main

import (
	"context"
	"fmt"
	"log/slog"
)

// As interfaces ficam no lado da FACHADA (quem consome), e so com o que ela usa.
type estoque interface {
	Reservar(sku string, qtd int) error
	Liberar(sku string, qtd int)
}
type frete interface {
	Calcular(cep string, itens int) (int64, error)
}
type pagamento interface {
	Cobrar(ctx context.Context, cartao string, centavos int64) (string, error)
	Estornar(id string) error
}
type fiscal interface {
	EmitirNota(cliente string, centavos int64) (string, error)
}
type notificacao interface {
	Avisar(destino, texto string) error
}

type Linha struct {
	SKU        string
	Quantidade int
	Centavos   int64 // preco unitario
}

type Pedido struct {
	Cliente string
	Email   string
	CEP     string
	Cartao  string
	Linhas  []Linha
}

type Recibo struct {
	Cobranca string
	Nota     string
	Total    int64
	Frete    int64
}

// Loja e a FACHADA: um metodo para o caso de uso "finalizar a compra".
type Loja struct {
	estoque estoque
	frete   frete
	pgto    pagamento
	fiscal  fiscal
	aviso   notificacao
	log     *slog.Logger
}

func NovaLoja(e estoque, f frete, p pagamento, fi fiscal, n notificacao, l *slog.Logger) *Loja {
	return &Loja{e, f, p, fi, n, l}
}

// Finalizar coordena os cinco servicos na ordem certa e DESFAZ o que ja foi feito se algo falhar.
func (l *Loja) Finalizar(ctx context.Context, p Pedido) (rec Recibo, err error) {
	var desfazer []func()
	defer func() {
		if err != nil {
			for i := len(desfazer) - 1; i >= 0; i-- { // desfaz em ordem inversa
				desfazer[i]()
			}
		}
	}()

	var subtotal int64
	itens := 0
	for _, ln := range p.Linhas {
		if err := l.estoque.Reservar(ln.SKU, ln.Quantidade); err != nil {
			return Recibo{}, fmt.Errorf("reservando %s: %w", ln.SKU, err)
		}
		desfazer = append(desfazer, func() { l.estoque.Liberar(ln.SKU, ln.Quantidade) })
		subtotal += ln.Centavos * int64(ln.Quantidade)
		itens += ln.Quantidade
	}

	fr, err := l.frete.Calcular(p.CEP, itens)
	if err != nil {
		return Recibo{}, fmt.Errorf("calculando frete: %w", err)
	}
	total := subtotal + fr

	cobranca, err := l.pgto.Cobrar(ctx, p.Cartao, total)
	if err != nil {
		return Recibo{}, fmt.Errorf("cobrando: %w", err)
	}
	desfazer = append(desfazer, func() {
		if e := l.pgto.Estornar(cobranca); e != nil {
			l.log.Error("estorno falhou", "cobranca", cobranca, "erro", e)
		}
	})

	nota, err := l.fiscal.EmitirNota(p.Cliente, total)
	if err != nil {
		return Recibo{}, fmt.Errorf("emitindo nota: %w", err)
	}

	// Aviso e "melhor esforco": falhar aqui NAO desfaz uma compra ja paga e faturada.
	if e := l.aviso.Avisar(p.Email, "pedido confirmado: "+nota); e != nil {
		l.log.Warn("aviso nao enviado", "cliente", p.Cliente, "erro", e)
	}
	return Recibo{Cobranca: cobranca, Nota: nota, Total: total, Frete: fr}, nil
}
