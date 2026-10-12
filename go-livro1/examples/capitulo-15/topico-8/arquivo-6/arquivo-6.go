// pedido/pedido.go
package pedido

// Cobrador é o que o pedido exige de quem cobra. É definido AQUI, junto de quem o usa.
type Cobrador interface {
	Cobrar(ref string, valor int64) error
}

type Pedido struct {
	ID    int
	Valor int64
}

// Servico depende da abstração, e não do pacote pagamento.
type Servico struct {
	cobrador Cobrador
}

func NovoServico(c Cobrador) *Servico { return &Servico{cobrador: c} }

func (s *Servico) Pagar(p Pedido) error {
	return s.cobrador.Cobrar(refDe(p), p.Valor)
}
