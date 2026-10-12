// Package pix fornece o meio de pagamento por pix.
package pix

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-16/projeto-1/cobranca"
)

type meio struct{}

func (meio) Cobrar(centavos int64) (string, error) {
	return fmt.Sprintf("PIX-%d", centavos), nil
}

// Novo devolve o meio. Quem registra e a raiz de composicao, nao este pacote.
func Novo() cobranca.Meio { return meio{} }
