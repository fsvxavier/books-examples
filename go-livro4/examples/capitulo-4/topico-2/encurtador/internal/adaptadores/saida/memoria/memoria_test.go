package memoria_test

import (
	"testing"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/adaptadores/saida/memoria"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/nucleo"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/nucleo/repositoriotest"
)

func TestContrato(t *testing.T) {
	repositoriotest.Executar(t, func(*testing.T) nucleo.Repositorio { return memoria.Novo() })
}
