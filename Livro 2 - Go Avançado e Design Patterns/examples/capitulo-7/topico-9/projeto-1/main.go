package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type relogio struct {
	mu sync.Mutex
	t  time.Time
}

func (r *relogio) Agora() time.Time { r.mu.Lock(); defer r.mu.Unlock(); return r.t }
func (r *relogio) Avancar(d time.Duration) {
	r.mu.Lock()
	r.t = r.t.Add(d)
	r.mu.Unlock()
}

func main() {
	ana := Perfil{1, "Ana", "ana@x.com"}
	bia := Perfil{2, "Bia", "bia@x.com"}
	ctxAna := ComUsuario(context.Background(), Usuario{ID: 1})
	ctxBia := ComUsuario(context.Background(), Usuario{ID: 2})

	fmt.Println("== Protecao")
	banco := NovoBanco(ana, bia)
	var p Perfis = Proteger(banco)
	_, err := p.Obter(ctxAna, 1)
	fmt.Println("Ana le o proprio perfil:", err)
	_, err = p.Obter(ctxAna, 2)
	fmt.Println("Ana le o de Bia:", err)
	_, err = p.Obter(context.Background(), 1)
	fmt.Println("sem identidade:", err)
	_, err = p.Obter(ComUsuario(context.Background(), Usuario{ID: 9, Admin: true}), 2)
	fmt.Println("admin le o de Bia:", err)

	fmt.Println("== Cache (TTL de 1 min, relogio injetado)")
	banco = NovoBanco(ana, bia)
	rel := &relogio{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)}
	cache := ComCache(banco, time.Minute, rel.Agora)
	for range 3 {
		cache.Obter(ctxAna, 1)
	}
	fmt.Printf("3 leituras: banco consultado %d vez, acertos=%d falhas=%d\n", banco.Leituras.Load(), cache.Acertos, cache.Falhas)
	rel.Avancar(61 * time.Second)
	cache.Obter(ctxAna, 1)
	fmt.Printf("depois de 61 s: banco consultado %d vezes\n", banco.Leituras.Load())
	cache.Atualizar(ctxAna, Perfil{1, "Ana Maria", "ana@x.com"})
	novo, _ := cache.Obter(ctxAna, 1)
	fmt.Printf("depois de Atualizar: le %q (invalidou), banco consultado %d vezes\n", novo.Nome, banco.Leituras.Load())

	fmt.Println("== Estampida: 100 leituras simultaneas, banco com 50 ms de latencia")
	banco = NovoBanco(ana)
	banco.Latencia = 50 * time.Millisecond
	cache = ComCache(banco, time.Minute, time.Now)
	var wg sync.WaitGroup
	inicio := time.Now()
	for range 100 {
		wg.Go(func() { cache.Obter(ctxAna, 1) })
	}
	wg.Wait()
	fmt.Printf("banco consultado %d vez em ~%v\n", banco.Leituras.Load(), time.Since(inicio).Round(50*time.Millisecond))

	fmt.Println("== ORDEM das camadas: cache por FORA da protecao vaza dados")
	for _, c := range []struct {
		nome string
		mont func(*Banco) Perfis
	}{
		{"Protecao(Cache(Banco)): correto", func(b *Banco) Perfis { return Proteger(ComCache(b, time.Minute, time.Now)) }},
		{"Cache(Protecao(Banco)): VAZA   ", func(b *Banco) Perfis { return ComCache(Proteger(b), time.Minute, time.Now) }},
	} {
		s := c.mont(NovoBanco(ana, bia))
		s.Obter(ctxBia, 2) // Bia le o proprio perfil (autorizado) e aquece o cache
		_, err := s.Obter(ctxAna, 2)
		fmt.Printf("  %s -> Ana le o perfil de Bia: %v\n", c.nome, err)
	}

	fmt.Println("== Log por fora x por dentro da protecao")
	b := NovoBanco(ana, bia)
	logDentro := ComLog("dentro", b)
	logFora := ComLog("fora", Proteger(logDentro))
	logFora.Obter(ctxAna, 2) // negado
	logFora.Obter(ctxAna, 1) // permitido
	for _, l := range append(logFora.Linhas, logDentro.Linhas...) {
		fmt.Println("  ", l)
	}
}
