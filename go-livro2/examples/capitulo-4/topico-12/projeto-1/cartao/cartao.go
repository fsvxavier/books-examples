// Package cartao registra o meio "cartao" no pacote pagamento ao ser importado.
package cartao

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/pagamento"
)

type meio struct{}

func (meio) Cobrar(centavos int64) (string, error) {
	return fmt.Sprintf("%d centavos via cartao de credito", centavos), nil
}

func init() {
	pagamento.Registrar("cartao", func() pagamento.Meio { return meio{} })
}
