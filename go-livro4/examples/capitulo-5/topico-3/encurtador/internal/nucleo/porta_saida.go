package nucleo

import (
	"context"
	"time"

	"exemplos/livro4/capitulo-5/topico-3/encurtador/internal/nucleo/dominio"
)

// Repositorio guarda e recupera o agregado Link, inteiro. Os erros devolvidos
// são os do domínio (ErrCodigoEmUso, ErrNaoEncontrado).
type Repositorio interface {
	// Salvar guarda um link novo, ou devolve dominio.ErrCodigoEmUso.
	Salvar(ctx context.Context, l dominio.Link) error
	// Buscar reconstitui o link, ou devolve dominio.ErrNaoEncontrado.
	Buscar(ctx context.Context, codigo dominio.Codigo) (dominio.Link, error)
}

// Contador guarda os totais de acessos. Fica fora do agregado: somar um é uma
// operação atômica do armazenamento, que não precisa carregar o Link.
type Contador interface {
	// Incrementar soma um acesso ao código, de forma atômica.
	Incrementar(ctx context.Context, codigo dominio.Codigo) error
	// Total devolve os acessos do código (zero se ninguém acessou ainda).
	Total(ctx context.Context, codigo dominio.Codigo) (int, error)
}

// Armazenamento reúne as duas portas, para os adaptadores que implementam as
// duas (a memória e o PostgreSQL).
type Armazenamento interface {
	Repositorio
	Contador
}

// Publicador é a porta de NOTIFICAÇÃO: o núcleo entrega os eventos de domínio
// ao mundo e não espera resposta. Quem reage a eles (auditoria, contagem,
// um dia uma fila) é decidido na composição, e não aqui.
type Publicador interface {
	Publicar(ctx context.Context, eventos ...dominio.Evento)
}

// Gerador sorteia um código novo a cada chamada.
type Gerador func() string

// Relogio informa o instante atual.
type Relogio func() time.Time
