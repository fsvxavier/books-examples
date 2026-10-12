// models/pedido.go
package models

import "exemplo.com/ciclo/services"

type Pedido struct {
	ID    int
	Valor int64
}

// Frete usa um serviço: a regra de cálculo "vazou" para a camada de modelos.
func (p Pedido) Frete() int64 { return services.CalcularFrete(p.Valor) }
