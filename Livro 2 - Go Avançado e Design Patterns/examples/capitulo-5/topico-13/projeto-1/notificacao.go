package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

// ---------- IMPLEMENTACAO: canais de entrega ----------

type Canal interface {
	Nome() string
	Enviar(ctx context.Context, destino, texto string) error
}

var ErrTemporario = errors.New("falha temporaria")

type Email struct{}

func (Email) Nome() string { return "email" }
func (Email) Enviar(ctx context.Context, destino, texto string) error {
	fmt.Printf("  [email] para %s: %s\n", destino, texto)
	return nil
}

type SMS struct{}

func (SMS) Nome() string { return "sms" }
func (SMS) Enviar(ctx context.Context, destino, texto string) error {
	if len(texto) > 40 {
		return fmt.Errorf("sms: texto com %d caracteres excede o limite de 40", len(texto))
	}
	fmt.Printf("  [sms] para %s: %s\n", destino, texto)
	return nil
}

type Slack struct{}

func (Slack) Nome() string { return "slack" }
func (Slack) Enviar(ctx context.Context, destino, texto string) error {
	fmt.Printf("  [slack] #%s: %s\n", destino, texto)
	return nil
}

// ---------- ABSTRACAO: o que significa "avisar" ----------

type Notificador interface {
	Avisar(ctx context.Context, destino, texto string) error
}

// Simples tenta uma vez. O canal pode ser trocado em tempo de execucao.
type Simples struct{ canal atomic.Pointer[Canal] }

func NovoSimples(c Canal) *Simples {
	s := &Simples{}
	s.Trocar(c)
	return s
}

func (s *Simples) Trocar(c Canal) { s.canal.Store(&c) }

func (s *Simples) Avisar(ctx context.Context, destino, texto string) error {
	return (*s.canal.Load()).Enviar(ctx, destino, texto)
}

// ComRetentativa repete se o canal devolver ErrTemporario. O tempo e injetado.
type ComRetentativa struct {
	Canal      Canal
	Tentativas int
	Espera     time.Duration
	Dormir     func(context.Context, time.Duration) error
}

func (n ComRetentativa) Avisar(ctx context.Context, destino, texto string) error {
	var err error
	for i := 1; i <= n.Tentativas; i++ {
		if err = n.Canal.Enviar(ctx, destino, texto); err == nil || !errors.Is(err, ErrTemporario) {
			return err
		}
		fmt.Printf("  tentativa %d/%d falhou: %v\n", i, n.Tentativas, err)
		if i < n.Tentativas {
			if e := n.Dormir(ctx, n.Espera<<(i-1)); e != nil {
				return e
			}
		}
	}
	return fmt.Errorf("desistindo depois de %d tentativas: %w", n.Tentativas, err)
}

// Urgente envia por TODOS os canais. So devolve erro se nenhum entregar;
// falhas parciais sao avisadas por Falhou (um gancho opcional).
type Urgente struct {
	Canais []Canal
	Falhou func(canal string, err error)
}

func (n Urgente) Avisar(ctx context.Context, destino, texto string) error {
	var erros []error
	for _, c := range n.Canais {
		err := c.Enviar(ctx, destino, "URGENTE: "+texto)
		if err == nil {
			continue
		}
		erros = append(erros, err)
		if n.Falhou != nil {
			n.Falhou(c.Nome(), err)
		}
	}
	if len(erros) == len(n.Canais) {
		return errors.Join(erros...)
	}
	return nil
}

func Dormir(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
