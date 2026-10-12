// services/frete.go
package services

import "github.com/fsvxavier/books-examples/go-livro-1/capitulo-15/topico-8/projeto-2/models"

func Frete(p models.Pedido) int64 {
	if p.Valor > 20000 {
		return 0
	}
	return 1290
}
