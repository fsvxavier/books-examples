package dominio

import "errors"

// Erros do domínio. Quem os recebe testa a categoria com errors.Is; o motivo
// detalhado vem embrulhado com %w.
var (
	ErrCodigoInvalido   = errors.New("código inválido")
	ErrDestinoInvalido  = errors.New("destino inválido")
	ErrValidadeInvalida = errors.New("validade inválida")
	ErrExpirado         = errors.New("link expirado")
	ErrNaoEncontrado    = errors.New("link não encontrado")
	ErrCodigoEmUso      = errors.New("código já está em uso")
)

// ErrSemCodigoLivre indica que todas as tentativas de sortear um código caíram
// em códigos já usados.
var ErrSemCodigoLivre = errors.New("não consegui gerar um código livre")
