package main

import (
	"errors"
	"sync"
	"testing"
)

func TestAbreUmaVezFechaNoUltimo(t *testing.T) {
	c := NovoCompartilhado(func() (*Conexao, error) { return &Conexao{}, nil })
	var soltar []func()
	for range 10 {
		_, s, err := c.Adquirir()
		if err != nil {
			t.Fatal(err)
		}
		soltar = append(soltar, s)
	}
	if c.Aberturas != 1 || c.Fechadas != 0 {
		t.Fatalf("ab=%d fe=%d", c.Aberturas, c.Fechadas)
	}
	for i, s := range soltar {
		s()
		if want := 0; i < 9 && c.Fechadas != want {
			t.Fatalf("fechou cedo, i=%d", i)
		}
	}
	if c.Fechadas != 1 || c.Clientes() != 0 {
		t.Fatalf("fe=%d clientes=%d", c.Fechadas, c.Clientes())
	}
}

func TestSoltarDuasVezesNaoDesequilibra(t *testing.T) {
	c := NovoCompartilhado(func() (*Conexao, error) { return &Conexao{}, nil })
	_, s1, _ := c.Adquirir()
	_, s2, _ := c.Adquirir()
	s1()
	s1()
	if c.Fechadas != 0 || c.Clientes() != 1 {
		t.Fatalf("o segundo s1() soltou a referencia de outro cliente: fe=%d clientes=%d", c.Fechadas, c.Clientes())
	}
	s2()
	if c.Fechadas != 1 {
		t.Fatal("deveria fechar")
	}
}

func TestErroNaAbertura(t *testing.T) {
	boom := errors.New("boom")
	c := NovoCompartilhado(func() (*Conexao, error) { return nil, boom })
	if _, _, err := c.Adquirir(); !errors.Is(err, boom) || c.Clientes() != 0 {
		t.Fatalf("err=%v clientes=%d", err, c.Clientes())
	}
}

func TestConcorrente(t *testing.T) {
	c := NovoCompartilhado(func() (*Conexao, error) { return &Conexao{}, nil })
	var wg sync.WaitGroup
	for range 200 {
		wg.Go(func() {
			r, s, err := c.Adquirir()
			if err != nil || r == nil {
				t.Error("adquirir falhou")
				return
			}
			s()
		})
	}
	wg.Wait()
	if c.Clientes() != 0 || c.Aberturas != c.Fechadas {
		t.Fatalf("clientes=%d ab=%d fe=%d", c.Clientes(), c.Aberturas, c.Fechadas)
	}
}
