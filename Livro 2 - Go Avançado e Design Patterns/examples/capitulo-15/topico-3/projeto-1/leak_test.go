package leak

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

func chamarComPressa(f func(context.Context, int) int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	f(ctx, 1)
	time.Sleep(time.Second) // tempo falso: dá a quem está só lento a chance de terminar
}

func TestConsultarVaza(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		chamarComPressa(Consultar)
	})
}

func TestConsultarOkNaoVaza(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		chamarComPressa(ConsultarOk)
	})
}
