package main

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/pagamento"
	_ "github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/pix"
)

type outroPix struct{}

func (outroPix) Cobrar(int64) (string, error) { return "outro", nil }

func main() {
	fmt.Println("antes de registrar de novo:", pagamento.Nomes())
	pagamento.Registrar("pix", func() pagamento.Meio { return outroPix{} })
}
