package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-7/topico-5/projeto-1/subsistema"
)

// diario registra a ordem de TODAS as chamadas aos servicos falsos.
type diario struct {
	mu      sync.Mutex
	eventos []string
}

func (d *diario) add(e string) { d.mu.Lock(); d.eventos = append(d.eventos, e); d.mu.Unlock() }

type estoqueFalso struct {
	d       *diario
	falhaEm string
}

func (e estoqueFalso) Reservar(sku string, q int) error {
	if sku == e.falhaEm {
		return subsistema.ErrSemEstoque
	}
	e.d.add("reservar:" + sku)
	return nil
}
func (e estoqueFalso) Liberar(sku string, q int) { e.d.add("liberar:" + sku) }

type freteFalso struct{}

func (freteFalso) Calcular(string, int) (int64, error) { return 100, nil }

type pgtoFalso struct {
	d     *diario
	falha bool
}

func (p pgtoFalso) Cobrar(_ context.Context, _ string, c int64) (string, error) {
	if p.falha {
		return "", errors.New("recusado")
	}
	p.d.add("cobrar")
	return "PG-1", nil
}
func (p pgtoFalso) Estornar(id string) error { p.d.add("estornar:" + id); return nil }

type fiscalFalso struct{ falha bool }

func (f fiscalFalso) EmitirNota(string, int64) (string, error) {
	if f.falha {
		return "", errors.New("fora")
	}
	return "NF-1", nil
}

type avisoFalso struct{ chamado *atomic.Bool }

func (a avisoFalso) Avisar(string, string) error { a.chamado.Store(true); return errors.New("falhou") }

func nova(d *diario, e estoqueFalso, p pgtoFalso, f fiscalFalso) (*Loja, *atomic.Bool) {
	chamado := &atomic.Bool{}
	return NovaLoja(e, freteFalso{}, p, f, avisoFalso{chamado}, slog.New(slog.NewTextHandler(io.Discard, nil))), chamado
}

var pedido = Pedido{Cliente: "ana", Email: "a@x", CEP: "01310100", Cartao: "4",
	Linhas: []Linha{{"A", 1, 10}, {"B", 1, 10}, {"C", 1, 10}}}

func TestSucessoNaOrdemCerta(t *testing.T) {
	d := &diario{}
	l, avisou := nova(d, estoqueFalso{d: d}, pgtoFalso{d: d}, fiscalFalso{})
	rec, err := l.Finalizar(context.Background(), pedido)
	if err != nil || rec.Nota != "NF-1" {
		t.Fatalf("rec=%+v err=%v", rec, err)
	}
	want := []string{"reservar:A", "reservar:B", "reservar:C", "cobrar"}
	if len(d.eventos) != len(want) {
		t.Fatalf("eventos = %v", d.eventos)
	}
	for i := range want {
		if d.eventos[i] != want[i] {
			t.Fatalf("eventos = %v", d.eventos)
		}
	}
	if !avisou.Load() {
		t.Fatal("o aviso deveria ter sido tentado")
	}
}

func TestDesfazEmOrdemInversa(t *testing.T) {
	d := &diario{}
	l, avisou := nova(d, estoqueFalso{d: d}, pgtoFalso{d: d}, fiscalFalso{falha: true})
	_, err := l.Finalizar(context.Background(), pedido)
	if err == nil {
		t.Fatal("esperava erro")
	}
	want := []string{"reservar:A", "reservar:B", "reservar:C", "cobrar", "estornar:PG-1", "liberar:C", "liberar:B", "liberar:A"}
	if len(d.eventos) != len(want) {
		t.Fatalf("eventos = %v", d.eventos)
	}
	for i := range want {
		if d.eventos[i] != want[i] {
			t.Fatalf("eventos = %v", d.eventos)
		}
	}
	if avisou.Load() {
		t.Fatal("nao deveria avisar de uma compra que falhou")
	}
}

func TestFalhaNaSegundaReservaLiberaSoAPrimeira(t *testing.T) {
	d := &diario{}
	l, _ := nova(d, estoqueFalso{d: d, falhaEm: "B"}, pgtoFalso{d: d}, fiscalFalso{})
	_, err := l.Finalizar(context.Background(), pedido)
	if !errors.Is(err, subsistema.ErrSemEstoque) {
		t.Fatalf("err = %v", err)
	}
	if len(d.eventos) != 2 || d.eventos[1] != "liberar:A" {
		t.Fatalf("eventos = %v", d.eventos)
	}
}

func TestContextoCancelado(t *testing.T) {
	pg := &subsistema.Pagamento{}
	est := subsistema.NovoEstoque(map[string]int{"A": 5})
	l := NovaLoja(est, subsistema.Frete{}, pg, subsistema.Fiscal{}, &subsistema.Notificacao{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := l.Finalizar(ctx, Pedido{CEP: "01310100", Linhas: []Linha{{"A", 1, 10}}})
	if !errors.Is(err, context.Canceled) || est.Saldo("A") != 5 {
		t.Fatalf("err=%v saldo=%d", err, est.Saldo("A"))
	}
}

// 50 clientes disputam 10 unidades: exatamente 10 compram, e o saldo fecha em zero.
func TestDisputaPeloUltimoEstoque(t *testing.T) {
	est := subsistema.NovoEstoque(map[string]int{"RARO": 10})
	pg := &subsistema.Pagamento{}
	l := NovaLoja(est, subsistema.Frete{}, pg, subsistema.Fiscal{}, &subsistema.Notificacao{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	var ok, semEstoque atomic.Int64
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			_, err := l.Finalizar(context.Background(), Pedido{Cliente: "c", CEP: "01310100", Linhas: []Linha{{"RARO", 1, 100}}})
			switch {
			case err == nil:
				ok.Add(1)
			case errors.Is(err, subsistema.ErrSemEstoque):
				semEstoque.Add(1)
			default:
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if ok.Load() != 10 || semEstoque.Load() != 40 || est.Saldo("RARO") != 0 || pg.Ativas() != 10 {
		t.Fatalf("ok=%d sem=%d saldo=%d cobrancas=%d", ok.Load(), semEstoque.Load(), est.Saldo("RARO"), pg.Ativas())
	}
}
