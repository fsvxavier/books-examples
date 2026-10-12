package memoria_test

import (
	"testing"

	"exemplos/livro4/capitulo-3/topico-2/encurtador/internal/adaptadores/memoria"
	"exemplos/livro4/capitulo-3/topico-2/encurtador/internal/aplicacao"
	"exemplos/livro4/capitulo-3/topico-2/encurtador/internal/aplicacao/repositoriotest"
)

func TestContrato(t *testing.T) {
	repositoriotest.Executar(t, func(*testing.T) aplicacao.Repositorio { return memoria.Novo() })
}
