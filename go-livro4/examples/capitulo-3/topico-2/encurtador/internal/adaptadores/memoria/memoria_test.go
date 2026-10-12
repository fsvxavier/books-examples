package memoria_test

import (
	"testing"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/adaptadores/memoria"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/aplicacao"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/aplicacao/repositoriotest"
)

func TestContrato(t *testing.T) {
	repositoriotest.Executar(t, func(*testing.T) aplicacao.Repositorio { return memoria.Novo() })
}
