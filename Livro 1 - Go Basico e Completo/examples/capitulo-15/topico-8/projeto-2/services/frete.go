// services/frete.go
package services

import "exemplo.com/camadas/models"

func Frete(p models.Pedido) int64 {
	if p.Valor > 20000 {
		return 0
	}
	return 1290
}
