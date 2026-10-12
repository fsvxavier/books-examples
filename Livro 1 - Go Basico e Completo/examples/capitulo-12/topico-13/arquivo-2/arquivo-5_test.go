package main

import (
	"context"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestAgendadorExecutaNoRitmo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
		defer cancel()

		var n atomic.Int32
		Agendador(ctx, 100*time.Millisecond, func(context.Context) { n.Add(1) })

		// 350ms virtuais com tique a cada 100ms: execuções em 100, 200 e 300ms
		if got := n.Load(); got != 3 {
			t.Errorf("execuções = %d, quer 3", got)
		}
	})
}

func TestAgendadorEsperaExecucaoEmAndamento(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		var concluiu atomic.Bool
		retornou := make(chan struct{})
		go func() {
			Agendador(ctx, 100*time.Millisecond, func(context.Context) {
				time.Sleep(500 * time.Millisecond) // execução longa
				concluiu.Store(true)
			})
			close(retornou)
		}()

		time.Sleep(150 * time.Millisecond) // a primeira execução (100ms) já começou
		cancel()                           // pede o encerramento no meio da execução
		<-retornou

		if !concluiu.Load() {
			t.Error("o Agendador retornou antes de a execução em andamento terminar")
		}
	})
}
