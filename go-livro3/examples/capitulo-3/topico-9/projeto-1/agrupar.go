// Package lote agrupa itens de um canal em lotes por tamanho ou por tempo.
package lote

import (
	"context"
	"time"
)

// Agrupar lê de entrada e emite lotes de até tam itens. Um lote parcial sai
// quando passa espera desde o primeiro item dele, para que poucos itens não
// fiquem parados esperando companhia. Ao fechar a entrada, o resto sai.
// O canal de saída é fechado quando a entrada acaba ou o contexto é cancelado.
func Agrupar[T any](ctx context.Context, entrada <-chan T, tam int, espera time.Duration) <-chan []T {
	saida := make(chan []T)
	go func() {
		defer close(saida)
		var (
			lote    = make([]T, 0, tam)
			prazo   <-chan time.Time
			relogio *time.Timer
		)
		envia := func() bool {
			if len(lote) == 0 {
				return true
			}
			select {
			case saida <- lote:
			case <-ctx.Done():
				return false
			}
			lote = make([]T, 0, tam)
			if relogio != nil {
				relogio.Stop()
			}
			prazo = nil
			return true
		}
		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-entrada:
				if !ok {
					envia()
					return
				}
				if len(lote) == 0 {
					if relogio == nil {
						relogio = time.NewTimer(espera)
					} else {
						relogio.Reset(espera)
					}
					prazo = relogio.C
				}
				lote = append(lote, item)
				if len(lote) >= tam && !envia() {
					return
				}
			case <-prazo:
				if !envia() {
					return
				}
			}
		}
	}()
	return saida
}
