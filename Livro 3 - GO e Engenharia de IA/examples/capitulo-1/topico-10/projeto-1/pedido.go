// Package llm define a interface de um provedor de modelo de linguagem e
// as peças de resiliência em volta dele: retentativa, disjuntor e fallback.
package llm

import (
	"context"
	"errors"
	"time"
)

// Pedido é o que se envia ao modelo.
type Pedido struct {
	Sistema   string
	Usuario   string
	MaxTokens int
}

// Resposta é o que volta.
type Resposta struct {
	Texto         string
	TokensEntrada int
	TokensSaida   int
	Modelo        string
}

// Modelo é qualquer provedor. A aplicação depende só desta interface, e é
// ela que permite trocar de provedor, testar com um falso e encadear
// fallbacks sem mexer no resto do código.
type Modelo interface {
	Nome() string
	Gerar(ctx context.Context, p Pedido) (Resposta, error)
	// Transmitir entrega a resposta em pedaços, chamando emitir a cada um.
	// Se emitir devolver erro (o cliente foi embora), a geração para.
	Transmitir(ctx context.Context, p Pedido, emitir func(pedaco string) error) (Resposta, error)
}

// ErroTemporario marca falhas que valem outra tentativa: limite de taxa
// (429), sobrecarga (503), queda de conexão. Depois: tente de novo.
type ErroTemporario struct {
	Causa      error
	EsperarAte time.Duration // sugestão do provedor (Retry-After); 0 = sem sugestão
}

func (e *ErroTemporario) Error() string { return "erro temporário: " + e.Causa.Error() }
func (e *ErroTemporario) Unwrap() error { return e.Causa }

// ErroPermanente marca falhas que nenhuma nova tentativa resolve: chave
// inválida (401), pedido malformado (400), conteúdo recusado.
type ErroPermanente struct{ Causa error }

func (e *ErroPermanente) Error() string { return "erro permanente: " + e.Causa.Error() }
func (e *ErroPermanente) Unwrap() error { return e.Causa }

// Temporario diz se vale tentar de novo.
func Temporario(err error) bool {
	var t *ErroTemporario
	return errors.As(err, &t)
}
