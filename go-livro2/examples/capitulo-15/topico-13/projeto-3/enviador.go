package bom

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrMensagemVazia = errors.New("mensagem vazia")
	ErrFilaCheia     = errors.New("fila cheia")
)

// Enviador é a dependência: quem consome declara o que precisa.
type Enviador interface {
	Enviar(ctx context.Context, msg string) error
}

type Opcoes struct {
	Capacidade int           // tamanho da fila
	Tentativas int           // tentativas por mensagem
	Espera     time.Duration // espera entre tentativas
}

type Notificador struct {
	env  Enviador
	op   Opcoes
	fila chan string
	wg   sync.WaitGroup

	mu       sync.Mutex
	enviados int
	falhas   []error
}

func Novo(env Enviador, op Opcoes) *Notificador {
	return &Notificador{env: env, op: op, fila: make(chan string, op.Capacidade)}
}

// Enfileirar nunca bloqueia: fila cheia é um erro que o chamador decide como tratar.
func (n *Notificador) Enfileirar(msg string) error {
	if msg == "" {
		return ErrMensagemVazia
	}
	select {
	case n.fila <- msg:
		return nil
	default:
		return ErrFilaCheia
	}
}

// Iniciar começa o trabalhador. Ele termina quando ctx é cancelado, depois de esvaziar a fila.
func (n *Notificador) Iniciar(ctx context.Context) {
	n.wg.Go(func() {
		for {
			select {
			case m := <-n.fila:
				n.processar(ctx, m)
			case <-ctx.Done():
				n.esvaziar()
				return
			}
		}
	})
}

// Esperar bloqueia até o trabalhador terminar (use após cancelar o contexto).
func (n *Notificador) Esperar() { n.wg.Wait() }

func (n *Notificador) esvaziar() {
	// ctx já foi cancelado: o fluxo de encerramento usa um contexto novo e curto.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for {
		select {
		case m := <-n.fila:
			n.processar(ctx, m)
		default:
			return
		}
	}
}

func (n *Notificador) processar(ctx context.Context, m string) {
	var err error
	for i := range n.op.Tentativas {
		if err = n.env.Enviar(ctx, m); err == nil {
			n.mu.Lock()
			n.enviados++
			n.mu.Unlock()
			return
		}
		if i < n.op.Tentativas-1 {
			select {
			case <-time.After(n.op.Espera):
			case <-ctx.Done():
			}
		}
	}
	n.mu.Lock()
	n.falhas = append(n.falhas, fmt.Errorf("mensagem %q: %w", m, err))
	n.mu.Unlock()
}

func (n *Notificador) Enviados() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.enviados
}

func (n *Notificador) Falhas() []error {
	n.mu.Lock()
	defer n.mu.Unlock()
	return append([]error(nil), n.falhas...)
}
