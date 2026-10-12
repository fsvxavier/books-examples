package nucleo

import (
	"context"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

// Repositorio é uma porta secundária (de saída): o que a aplicação exige do
// mundo de fora para guardar links. Os erros devolvidos são os do domínio
// (ErrCodigoEmUso, ErrNaoEncontrado); cada adaptador traduz o dialeto da sua
// tecnologia para eles.
type Repositorio interface {
	// Salvar guarda o link, ou devolve dominio.ErrCodigoEmUso.
	Salvar(ctx context.Context, l dominio.Link) error
	// Buscar devolve o link, ou dominio.ErrNaoEncontrado.
	Buscar(ctx context.Context, codigo string) (dominio.Link, error)
	// RegistrarAcesso soma um acesso, de forma atômica, e devolve o link atualizado.
	RegistrarAcesso(ctx context.Context, codigo string) (dominio.Link, error)
}

// Auditoria é uma porta secundária de NOTIFICAÇÃO: a aplicação avisa o mundo
// de que algo aconteceu e não espera resposta (na definição de Cockburn, um
// ator secundário "para obter respostas ou apenas para notificar"). Os métodos
// não devolvem erro: uma auditoria que falha não pode derrubar um redirecionamento.
type Auditoria interface {
	LinkCriado(ctx context.Context, l dominio.Link)
	LinkAcessado(ctx context.Context, l dominio.Link)
}

// Gerador sorteia um código novo a cada chamada.
type Gerador func() string

// Relogio informa o instante atual.
type Relogio func() time.Time
