package memoria_test

import (
	"testing"

	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/adaptadores/saida/memoria"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo/repositoriotest"
)

func TestContrato(t *testing.T) {
	repositoriotest.Executar(t, func(*testing.T) nucleo.Repositorio { return memoria.Novo() })
}
