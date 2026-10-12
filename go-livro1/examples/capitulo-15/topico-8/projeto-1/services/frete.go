// services/frete.go
package services

import "github.com/fsvxavier/books-examples/go-livro-1/capitulo-15/topico-8/projeto-1/models"

func CalcularFrete(valor int64) int64 {
	if valor > 20000 {
		return 0
	}
	return 1290
}

func ResumoDoPedido(p models.Pedido) string { return "pedido pronto" }
