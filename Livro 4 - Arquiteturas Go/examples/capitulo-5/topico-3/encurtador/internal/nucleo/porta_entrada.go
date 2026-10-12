package nucleo

import (
	"context"

	"exemplos/livro4/capitulo-5/topico-3/encurtador/internal/nucleo/dominio"
)

// Consulta reúne o link e o total de acessos, que vive fora do agregado.
type Consulta struct {
	Link    dominio.Link
	Acessos int
}

// Encurtador é a porta de ENTRADA: o que o núcleo oferece ao mundo.
type Encurtador interface {
	// Criar encurta um destino e devolve o link criado.
	Criar(ctx context.Context, destino string) (dominio.Link, error)
	// Resolver devolve o link para seguir (e registra o acesso), ou ErrExpirado.
	Resolver(ctx context.Context, codigo string) (dominio.Link, error)
	// Consultar devolve o link e o total de acessos, sem contar acesso.
	Consultar(ctx context.Context, codigo string) (Consulta, error)
}
