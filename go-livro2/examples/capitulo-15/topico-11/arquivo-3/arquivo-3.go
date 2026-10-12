package main

import (
	"fmt"
	"sync"
	"syscall"
	"time"
)

func cpuUsada() time.Duration {
	var ru syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
	return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
}

// esperaOcupada: select com default dentro de for vira laço que consome um núcleo inteiro.
func esperaOcupada(ch <-chan int, parar <-chan struct{}) {
	for {
		select {
		case <-ch:
		case <-parar:
			return
		default:
			// nada a fazer: volta imediatamente
		}
	}
}

func esperaBloqueada(ch <-chan int, parar <-chan struct{}) {
	for {
		select {
		case <-ch:
		case <-parar:
			return
		}
	}
}

func medirEspera(nome string, f func(<-chan int, <-chan struct{})) {
	ch := make(chan int)
	parar := make(chan struct{})
	antes := cpuUsada()
	done := make(chan struct{})
	go func() { f(ch, parar); close(done) }()
	time.Sleep(500 * time.Millisecond) // a goroutine não tem nada para fazer durante meio segundo
	close(parar)
	<-done
	fmt.Printf("%-18s CPU gasta em 500 ms ociosos: %v\n", nome, (cpuUsada() - antes).Round(10*time.Millisecond))
}

// ---- mutex segurado durante chamada lenta ----

type cache struct {
	mu sync.Mutex
	m  map[int]int
}

func chamadaLenta(k int) int {
	time.Sleep(20 * time.Millisecond) // rede, disco, banco...
	return k * k
}

// RUIM: a trava cobre a chamada lenta; todos esperam em fila.
func (c *cache) obterRuim(k int) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.m[k]; ok {
		return v
	}
	v := chamadaLenta(k)
	c.m[k] = v
	return v
}

// BOM: trava só o acesso ao mapa; a chamada lenta roda sem trava.
func (c *cache) obter(k int) int {
	c.mu.Lock()
	v, ok := c.m[k]
	c.mu.Unlock()
	if ok {
		return v
	}
	v = chamadaLenta(k)
	c.mu.Lock()
	c.m[k] = v
	c.mu.Unlock()
	return v
}

func medirCache(nome string, f func(c *cache, k int) int) {
	c := &cache{m: map[int]int{}}
	var wg sync.WaitGroup
	inicio := time.Now()
	for k := range 10 {
		wg.Go(func() { f(c, k) })
	}
	wg.Wait()
	fmt.Printf("%-24s 10 chaves distintas em %v\n", nome, time.Since(inicio).Round(10*time.Millisecond))
}

func main() {
	medirEspera("select com default", esperaOcupada)
	medirEspera("select bloqueante", esperaBloqueada)
	medirCache("mutex na chamada lenta", (*cache).obterRuim)
	medirCache("mutex só no mapa", (*cache).obter)
}
