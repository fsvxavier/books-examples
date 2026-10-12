package nucleo

import (
	"context"

	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

// Encurtador é a porta primária (de entrada): tudo que o mundo de fora pode
// pedir à aplicação. Quem a usa são os atores que DIRIGEM a aplicação: o
// navegador por HTTP, a pessoa no terminal, um script em lote, um teste.
// A aplicação a declara e a oferece; os adaptadores de entrada a consomem.
type Encurtador interface {
	// Criar encurta um destino e devolve o link criado.
	Criar(ctx context.Context, destino string) (dominio.Link, error)
	// Resolver devolve o link de um código e conta um acesso.
	Resolver(ctx context.Context, codigo string) (dominio.Link, error)
	// Consultar devolve o link de um código, sem contar acesso.
	Consultar(ctx context.Context, codigo string) (dominio.Link, error)
}
