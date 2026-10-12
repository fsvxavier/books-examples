package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

// Componente é uma parte do serviço com ciclo de vida.
type Componente struct {
	Nome string
	// Iniciar bloqueia enquanto o componente roda; devolve quando para.
	Iniciar func(ctx context.Context) error
	// Parar pede um encerramento ordenado, respeitando o prazo do contexto.
	Parar func(ctx context.Context) error
}

// Orquestrar inicia todos, espera o cancelamento de ctx (um sinal, ou a
// falha de um componente) e então os PARA EM ORDEM INVERSA, cada um com
// o seu prazo. Devolve o primeiro erro de início, se houve, junto com os
// erros de parada.
func Orquestrar(ctx context.Context, prazo time.Duration, log func(string, ...any), comps ...Componente) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, c := range comps {
		g.Go(func() error {
			log("iniciando %s", c.Nome)
			err := c.Iniciar(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				log("%s falhou: %v", c.Nome, err)
				return fmt.Errorf("%s: %w", c.Nome, err)
			}
			return nil
		})
	}

	var errosParada error
	paradaFeita := make(chan struct{})
	go func() {
		defer close(paradaFeita)
		<-ctx.Done() // sinal de fora OU falha de um componente
		// O contexto de parada NÃO herda o cancelamento: senão o prazo
		// de encerramento nasceria já vencido.
		pctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), prazo)
		defer cancel()
		for i := len(comps) - 1; i >= 0; i-- {
			c := comps[i]
			log("parando %s", c.Nome)
			if err := c.Parar(pctx); err != nil {
				errosParada = errors.Join(errosParada, fmt.Errorf("parar %s: %w", c.Nome, err))
			}
		}
	}()

	err := g.Wait()
	<-paradaFeita
	return errors.Join(err, errosParada)
}
