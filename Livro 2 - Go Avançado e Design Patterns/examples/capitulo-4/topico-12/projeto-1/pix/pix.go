// Package pix registra o meio "pix" no pacote pagamento ao ser importado.
package pix

import (
	"fmt"

	"exemplos/livro2/capitulo-4/topico-12/projeto-1/pagamento"
)

type meio struct{}

func (meio) Cobrar(centavos int64) (string, error) {
	return fmt.Sprintf("%d centavos via pix", centavos), nil
}

func init() {
	pagamento.Registrar("pix", func() pagamento.Meio { return meio{} })
}
