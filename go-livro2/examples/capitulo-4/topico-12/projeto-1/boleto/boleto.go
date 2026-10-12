// Package boleto registra o meio "boleto" no pacote pagamento ao ser importado.
package boleto

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/pagamento"
)

type meio struct{}

func (meio) Cobrar(centavos int64) (string, error) {
	return fmt.Sprintf("%d centavos via boleto bancario", centavos), nil
}

func init() {
	pagamento.Registrar("boleto", func() pagamento.Meio { return meio{} })
}
