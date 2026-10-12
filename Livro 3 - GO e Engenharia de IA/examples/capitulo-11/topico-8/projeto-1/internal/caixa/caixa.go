// Package caixa implementa agentes como atores: cada um tem uma caixa de
// entrada (um canal) e uma goroutine que a lê. Os agentes só conversam por
// mensagens, e nunca compartilham memória.
package caixa

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Mensagem é a unidade de comunicação.
type Mensagem struct {
	Conversa int
	De, Para string
	Tipo     string
	Corpo    string
}

// Agente trata uma mensagem e devolve as mensagens a enviar em resposta.
type Agente interface {
	Tratar(ctx context.Context, m Mensagem) ([]Mensagem, error)
}

// ErrDestino indica uma mensagem para quem não existe.
var ErrDestino = errors.New("destino desconhecido")

// Barramento entrega mensagens entre os agentes registrados.
type Barramento struct {
	caixas        map[string]chan Mensagem
	agentes       map[string]Agente
	trabalhadores map[string]int
}

// Novo cria um barramento vazio.
func Novo() *Barramento {
	return &Barramento{caixas: map[string]chan Mensagem{}, agentes: map[string]Agente{}, trabalhadores: map[string]int{}}
}

// Registrar cria a caixa de entrada de nome, com a capacidade dada.
func (b *Barramento) Registrar(nome string, capacidade int, a Agente) {
	b.caixas[nome] = make(chan Mensagem, capacidade)
	b.agentes[nome] = a
}

// Trabalhadores define quantas goroutines leem a caixa de nome (padrão: uma).
// Com mais de uma, o agente passa a tratar mensagens ao mesmo tempo e o seu
// estado interno precisa ser protegido.
func (b *Barramento) Trabalhadores(nome string, n int) { b.trabalhadores[nome] = n }

// Caixa devolve o canal de entrada de um nome (útil para ler as saídas).
func (b *Barramento) Caixa(nome string) <-chan Mensagem { return b.caixas[nome] }

// Enviar entrega m na caixa do destino. Espera se a caixa estiver cheia (é a
// contrapressão), mas desiste se o contexto acabar.
func (b *Barramento) Enviar(ctx context.Context, m Mensagem) error {
	c, ok := b.caixas[m.Para]
	if !ok {
		return fmt.Errorf("%q: %w", m.Para, ErrDestino)
	}
	select {
	case c <- m:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Executar inicia uma goroutine por agente (menos os que não têm Agente,
// como a caixa de saída) e só retorna quando o contexto acaba e todas terminam.
func (b *Barramento) Executar(ctx context.Context) error {
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		primeiro error
	)
	for nome, a := range b.agentes {
		if a == nil {
			continue
		}
		for range max(b.trabalhadores[nome], 1) {
			wg.Go(func() {
				for {
					select {
					case <-ctx.Done():
						return
					case m := <-b.caixas[nome]:
						saidas, err := a.Tratar(ctx, m)
						if err != nil {
							mu.Lock()
							if primeiro == nil {
								primeiro = fmt.Errorf("agente %q: %w", nome, err)
							}
							mu.Unlock()
							continue
						}
						for _, s := range saidas {
							if err := b.Enviar(ctx, s); err != nil {
								return
							}
						}
					}
				}
			})
		}
	}
	wg.Wait()
	return primeiro
}
