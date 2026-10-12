package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrFechado = errors.New("lote: fechado")

type pedido[T any] struct {
	item T
	res  chan error // buffer 1: o laço nunca bloqueia ao responder
}

// Lote agrupa itens de vários chamadores e os grava de uma vez: quando
// junta MaxItens, ou quando passa MaxEspera desde o primeiro item.
type Lote[T any] struct {
	MaxItens  int
	MaxEspera time.Duration
	Gravar    func(ctx context.Context, itens []T) error

	once   sync.Once
	ch     chan pedido[T]
	fechar chan struct{}
	feito  chan struct{}
	mu     sync.RWMutex
	fechou bool
}

func (l *Lote[T]) iniciar() {
	l.once.Do(func() {
		l.ch = make(chan pedido[T])
		l.fechar = make(chan struct{})
		l.feito = make(chan struct{})
		go l.laco()
	})
}

// Enviar entrega o item e espera o resultado da gravação do lote que o
// contém. Se ctx expirar antes, devolve ctx.Err(), mas o item PODE ser
// gravado mesmo assim (já estava no lote).
func (l *Lote[T]) Enviar(ctx context.Context, item T) error {
	l.iniciar()
	l.mu.RLock()
	if l.fechou {
		l.mu.RUnlock()
		return ErrFechado
	}
	p := pedido[T]{item, make(chan error, 1)}
	select {
	case l.ch <- p:
		l.mu.RUnlock()
	case <-ctx.Done():
		l.mu.RUnlock()
		return ctx.Err()
	}
	select {
	case err := <-p.res:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Fechar grava o que está pendente e encerra o laço.
func (l *Lote[T]) Fechar() {
	l.iniciar()
	l.mu.Lock()
	if !l.fechou {
		l.fechou = true
		close(l.fechar)
	}
	l.mu.Unlock()
	<-l.feito
}

func (l *Lote[T]) laco() {
	defer close(l.feito)
	var itens []T
	var respostas []chan error
	var timer *time.Timer
	var prazo <-chan time.Time

	descarregar := func() {
		if len(itens) == 0 {
			return
		}
		err := l.gravarSeguro(itens)
		for _, r := range respostas {
			r <- err
		}
		itens, respostas = nil, nil
		if timer != nil {
			timer.Stop()
		}
		prazo = nil
	}

	for {
		select {
		case p := <-l.ch:
			if len(itens) == 0 {
				timer = time.NewTimer(l.MaxEspera)
				prazo = timer.C
			}
			itens = append(itens, p.item)
			respostas = append(respostas, p.res)
			if len(itens) >= l.MaxItens {
				descarregar()
			}
		case <-prazo:
			descarregar()
		case <-l.fechar:
			descarregar()
			return
		}
	}
}

func (l *Lote[T]) gravarSeguro(itens []T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("lote: panic ao gravar: %v", r)
		}
	}()
	return l.Gravar(context.Background(), itens)
}
