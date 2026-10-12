package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Estado int

const (
	Fechado    Estado = iota // tudo passa; as falhas são contadas
	Aberto                   // nada passa; falha na hora
	MeioAberto               // deixa passar UMA sonda para testar
)

func (e Estado) String() string { return [...]string{"fechado", "aberto", "meio-aberto"}[e] }

var ErrAberto = errors.New("disjuntor aberto")

// Disjuntor protege uma dependência que pode falhar ou ficar lenta.
type Disjuntor struct {
	LimiteFalhas int           // falhas seguidas que abrem o disjuntor
	Espera       time.Duration // quanto fica aberto antes de testar de novo
	// Conta decide se um erro conta como falha da dependência. Nil: todo erro conta.
	Conta func(error) bool

	mu       sync.Mutex
	estado   Estado
	falhas   int
	abertoEm time.Time
	sondando bool
}

func (d *Disjuntor) Estado() Estado {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.estado
}

func (d *Disjuntor) conta(err error) bool {
	if d.Conta == nil {
		return true
	}
	return d.Conta(err)
}

// permitir decide se a chamada passa e se ela é a sonda.
func (d *Disjuntor) permitir() (ok, sonda bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	switch d.estado {
	case Aberto:
		if time.Since(d.abertoEm) < d.Espera {
			return false, false
		}
		d.estado = MeioAberto
		d.sondando = true
		return true, true
	case MeioAberto:
		if d.sondando {
			return false, false // já há uma sonda em andamento
		}
		d.sondando = true
		return true, true
	}
	return true, false
}

func (d *Disjuntor) registrar(err error, sonda bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	falhou := err != nil && d.conta(err)
	if sonda {
		d.sondando = false
		if falhou {
			d.estado, d.abertoEm = Aberto, time.Now()
		} else {
			d.estado, d.falhas = Fechado, 0
		}
		return
	}
	if d.estado != Fechado {
		return // um resultado atrasado de antes de abrir: ignora
	}
	if !falhou {
		if err == nil {
			d.falhas = 0
		}
		return
	}
	d.falhas++
	if d.falhas >= d.LimiteFalhas {
		d.estado, d.abertoEm = Aberto, time.Now()
	}
}

// Executar roda f se o disjuntor deixar; senão devolve ErrAberto na hora.
func (d *Disjuntor) Executar(ctx context.Context, f func(context.Context) error) error {
	ok, sonda := d.permitir()
	if !ok {
		return ErrAberto
	}
	err := f(ctx)
	d.registrar(err, sonda)
	return err
}
