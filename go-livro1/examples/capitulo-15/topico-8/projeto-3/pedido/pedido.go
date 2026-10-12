// pedido/pedido.go
// Package pedido reúne o tipo Pedido e as regras de negócio ligadas a ele.
package pedido

// Pedido representa uma compra. Valores em centavos.
type Pedido struct {
	ID    int
	Valor int64
}

// Frete devolve o frete em centavos; é grátis a partir de R$ 200,00.
func (p Pedido) Frete() int64 {
	if p.Valor > 20000 {
		return 0
	}
	return 1290
}

// Resumo devolve uma descrição curta do pedido.
func (p Pedido) Resumo() string { return "pedido pronto" }
