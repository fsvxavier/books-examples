package main

import (
	"testing"
	"testing/synctest"
	"time"
)

func TestCacheExpiracao(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NovoCache[string](time.Minute)
		c.Definir("usuario:1", "Ana")

		if v, ok := c.Obter("usuario:1"); !ok || v != "Ana" {
			t.Fatalf("logo após gravar: v=%q ok=%t", v, ok)
		}

		time.Sleep(59 * time.Second) // tempo VIRTUAL: o teste não espera de verdade
		if _, ok := c.Obter("usuario:1"); !ok {
			t.Error("aos 59s o valor ainda deveria valer")
		}

		time.Sleep(2 * time.Second)
		if _, ok := c.Obter("usuario:1"); ok {
			t.Error("aos 61s o valor deveria ter expirado")
		}
		if n := c.Tamanho(); n != 1 {
			t.Errorf("a entrada vencida ainda ocupa espaço: tamanho=%d", n)
		}
		if r := c.LimparExpirados(); r != 1 || c.Tamanho() != 0 {
			t.Errorf("limpeza removeu %d, tamanho=%d", r, c.Tamanho())
		}
	})
}

func TestCacheConcorrente(t *testing.T) {
	c := NovoCache[int](time.Hour)
	done := make(chan struct{})
	for g := range 8 {
		go func() {
			for i := range 500 {
				c.Definir("k", g*1000+i)
				c.Obter("k")
			}
			done <- struct{}{}
		}()
	}
	for range 8 {
		<-done
	}
}
