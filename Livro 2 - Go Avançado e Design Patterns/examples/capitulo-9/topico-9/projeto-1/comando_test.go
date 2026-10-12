package main

import (
	"errors"
	"testing"
)

func TestMacroTudoOuNada(t *testing.T) {
	var log []string
	p := func(n string, falha bool) Comando {
		return Par{Nome: n,
			Faz: func() error {
				if falha {
					return errors.New("x")
				}
				log = append(log, "+"+n)
				return nil
			},
			Desfaz: func() error { log = append(log, "-"+n); return nil }}
	}
	m := &Macro{Nome: "m", Comandos: []Comando{p("a", false), p("b", false), p("c", true), p("d", false)}}
	if err := m.Executar(); err == nil {
		t.Fatal("esperava erro")
	}
	want := []string{"+a", "+b", "-b", "-a"}
	if len(log) != len(want) {
		t.Fatalf("log = %v", log)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("log = %v, quero %v", log, want)
		}
	}
}

func TestMacroJuntaErrosDoDesfazer(t *testing.T) {
	falhaAoDesfazer := Par{Nome: "p1", Faz: func() error { return nil }, Desfaz: func() error { return errors.New("nao consegui") }}
	falha := Par{Nome: "p2", Faz: func() error { return errors.New("original") }, Desfaz: func() error { return nil }}
	err := (&Macro{Nome: "m", Comandos: []Comando{falhaAoDesfazer, falha}}).Executar()
	if err == nil || !contem(err.Error(), "original") || !contem(err.Error(), "nao consegui") {
		t.Fatalf("err = %v: os dois erros deveriam aparecer", err)
	}
}

func contem(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestParSemInversa(t *testing.T) {
	if err := (Par{Nome: "x", Faz: func() error { return nil }}).Desfazer(); err == nil {
		t.Fatal("esperava erro")
	}
}

func TestMacroAninhada(t *testing.T) {
	n := 0
	inc := Par{Nome: "inc", Faz: func() error { n++; return nil }, Desfaz: func() error { n--; return nil }}
	nova := func() *Macro { return &Macro{Nome: "i", Comandos: []Comando{inc, inc}} } // uma instancia por uso
	externa := &Macro{Nome: "e", Comandos: []Comando{nova(), inc, nova()}}
	if err := externa.Executar(); err != nil || n != 5 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	externa.Desfazer()
	if n != 0 {
		t.Fatalf("n=%d depois de desfazer", n)
	}
}

// Um comando que guarda o estado do "desfazer" e de USO UNICO por execucao: reaproveitar a MESMA
// instancia em dois lugares da macro faz o segundo uso sobrescrever o estado do primeiro.
func TestMacroCompartilhadaEsqueceOPrimeiroUso(t *testing.T) {
	n := 0
	inc := Par{Nome: "inc", Faz: func() error { n++; return nil }, Desfaz: func() error { n--; return nil }}
	interna := &Macro{Nome: "i", Comandos: []Comando{inc, inc}}
	externa := &Macro{Nome: "e", Comandos: []Comando{interna, interna}}
	externa.Executar()
	externa.Desfazer()
	if n != 2 {
		t.Fatalf("n = %d: esperava 2 (o primeiro uso nao foi desfeito), a premissa mudou", n)
	}
}

func TestRemetenteTrocaDeComando(t *testing.T) {
	var vistos []string
	mk := func(s string) Comando {
		return Par{Nome: s, Faz: func() error { vistos = append(vistos, s); return nil }}
	}
	r := &Remetente{Nome: "r", Comando: mk("a")}
	r.Acionar()
	r.Comando = mk("b")
	r.Acionar()
	if len(vistos) != 2 || vistos[0] != "a" || vistos[1] != "b" {
		t.Fatalf("vistos = %v", vistos)
	}
}
