package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Cada operação é protegida, e ainda assim a lógica tem corrida: verificar-e-agir (check-then-act).
type cacheRuim struct {
	mu     sync.Mutex
	m      map[string]int
	cargas atomic.Int32
}

func (c *cacheRuim) obter(chave string) int {
	c.mu.Lock()
	v, ok := c.m[chave]
	c.mu.Unlock() // a trava é solta entre a verificação e a gravação
	if ok {
		return v
	}
	v = carregar(chave, &c.cargas)
	c.mu.Lock()
	c.m[chave] = v
	c.mu.Unlock()
	return v
}

// Contador com a mesma falha, sem nenhuma corrida de memória.
type saldo struct {
	mu    sync.Mutex
	valor int
}

func (s *saldo) get() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.valor
}

func (s *saldo) set(v int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.valor = v
}

func (s *saldo) sacarRuim(n int) bool {
	if s.get() >= n { // verifica
		time.Sleep(time.Millisecond)
		s.set(s.get() - n) // age: o saldo pode ter mudado entre as duas linhas
		return true
	}
	return false
}

func (s *saldo) sacar(n int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.valor >= n {
		s.valor -= n
		return true
	}
	return false
}

func carregar(chave string, cargas *atomic.Int32) int {
	cargas.Add(1)
	time.Sleep(20 * time.Millisecond)
	return len(chave)
}

func main() {
	c := &cacheRuim{m: map[string]int{}}
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() { c.obter("produto-42") })
	}
	wg.Wait()
	fmt.Println("cache ruim: 100 pedidos da mesma chave, cargas executadas:", c.cargas.Load())

	for _, caso := range []struct {
		nome  string
		sacar func(*saldo, int) bool
	}{{"sacarRuim", (*saldo).sacarRuim}, {"sacar", (*saldo).sacar}} {
		s := &saldo{valor: 100}
		var aprovados atomic.Int32
		var wg sync.WaitGroup
		for range 10 {
			wg.Go(func() {
				if caso.sacar(s, 100) {
					aprovados.Add(1)
				}
			})
		}
		wg.Wait()
		fmt.Printf("%-10s saldo inicial 100, 10 saques de 100: aprovados=%d saldo final=%d\n", caso.nome, aprovados.Load(), s.get())
	}
}
