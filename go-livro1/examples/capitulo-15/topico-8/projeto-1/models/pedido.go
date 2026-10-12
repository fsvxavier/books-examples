// models/pedido.go
package models

import "github.com/fsvxavier/books-examples/go-livro-1/capitulo-15/topico-8/projeto-1/services"

type Pedido struct {
	ID    int
	Valor int64
}

// Frete usa um serviço: a regra de cálculo "vazou" para a camada de modelos.
func (p Pedido) Frete() int64 { return services.CalcularFrete(p.Valor) }
