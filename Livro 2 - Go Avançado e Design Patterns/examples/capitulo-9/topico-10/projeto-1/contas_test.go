package main

import (
	"bytes"
	"errors"
	"maps"
	"sync"
	"testing"
)

func env(t *testing.T, id string, c Comando) Envelope {
	t.Helper()
	e, err := Empacotar(id, c)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestIdaEVoltaDoEnvelope(t *testing.T) {
	e := env(t, "1", Depositar{"ana", 500})
	c, err := e.Comando()
	if err != nil {
		t.Fatal(err)
	}
	d, ok := c.(Depositar)
	if !ok || d.Conta != "ana" || d.Valor != 500 {
		t.Fatalf("c = %#v", c)
	}
}

func TestIdempotencia(t *testing.T) {
	contas := NovasContas()
	p := NovoProcessador(contas, nil)
	e := env(t, "x", Depositar{"ana", 100})
	for range 5 {
		if err := p.Processar(e); err != nil {
			t.Fatal(err)
		}
	}
	if contas.Saldo("ana") != 100 {
		t.Fatalf("saldo = %d: o comando repetido foi executado mais de uma vez", contas.Saldo("ana"))
	}
}

func TestReproduzirReconstroiOEstado(t *testing.T) {
	contas := NovasContas()
	d := &Diario{}
	p := NovoProcessador(contas, d)
	for i, c := range []Comando{Depositar{"a", 100}, Sacar{"a", 30}, Depositar{"b", 50}, Sacar{"b", 80}, Sacar{"a", 10}} {
		p.Processar(env(t, string(rune('a'+i)), c)) // o 4o falha
	}
	refeitas, err := Reproduzir(bytes.NewReader(d.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(refeitas.Snapshot(), contas.Snapshot()) {
		t.Fatalf("refeitas=%v originais=%v", refeitas.Snapshot(), contas.Snapshot())
	}
}

func TestTipoDesconhecido(t *testing.T) {
	_, err := Envelope{ID: "1", Tipo: "voar"}.Comando()
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestSaldoInsuficienteEhTipado(t *testing.T) {
	p := NovoProcessador(NovasContas(), nil)
	err := p.Processar(env(t, "1", Sacar{"ana", 1}))
	if !errors.Is(err, ErrSaldoInsuficiente) {
		t.Fatalf("err = %v", err)
	}
}

// 20 produtores enviam 500 depositos pela fila: o estado final e exato.
func TestFilaConcorrente(t *testing.T) {
	contas := NovasContas()
	p := NovoProcessador(contas, &Diario{})
	f := NovaFila(p, 8)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			for j := range 25 {
				e, _ := Empacotar(string(rune('A'+i))+string(rune('a'+j)), Depositar{"cx", 10})
				f.Enviar(e)
			}
		})
	}
	wg.Wait()
	f.Encerrar()
	if contas.Saldo("cx") != 5000 {
		t.Fatalf("saldo = %d", contas.Saldo("cx"))
	}
}
