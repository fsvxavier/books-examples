package varredura

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

func TestSomaCompleta(t *testing.T) {
	dados := make([]byte, 1000)
	for i := range dados {
		dados[i] = byte(i % 7)
	}
	var visto []int
	soma, err := Somar(context.Background(), dados, func(p int) { visto = append(visto, p) })
	if err != nil {
		t.Fatal(err)
	}
	var quer int64
	for _, b := range dados {
		quer += int64(b)
	}
	if soma != quer {
		t.Errorf("soma = %d, quer %d", soma, quer)
	}
	if want := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}; !slices.Equal(visto, want) {
		t.Errorf("progresso = %v", visto)
	}
}

func TestCancelaNoMeio(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	var visto []int
	inicio := time.Now()
	_, err := Somar(ctx, make([]byte, 1000), func(p int) {
		visto = append(visto, p)
		if p == 30 {
			cancel(ErrCancelado)
		}
	})
	if !errors.Is(err, ErrCancelado) {
		t.Fatalf("err = %v, quer ErrCancelado", err)
	}
	t.Logf("progresso visto: %v em %v", visto, time.Since(inicio).Round(10*time.Millisecond))
	if last := visto[len(visto)-1]; last != 30 {
		t.Errorf("deveria parar em 30, parou em %d", last)
	}
}
