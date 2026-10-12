package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// aguardar repete a verificacao ate ela ser verdadeira (so para tornar a demonstracao determinista).
func aguardar(cond func() bool) {
	for !cond() {
		time.Sleep(time.Millisecond)
	}
}

func main() {
	torre := NovaTorre()
	defer torre.Encerrar()
	ctx := context.Background()

	var (
		mu    sync.Mutex
		ordem []string
		wg    sync.WaitGroup
	)
	pousou := func() int { mu.Lock(); defer mu.Unlock(); return len(ordem) }
	soltar := make(chan struct{}) // A só libera a pista depois que B, C e D estão na fila

	pouso := func(id string, prioridade int) {
		defer wg.Done()
		if err := torre.Solicitar(ctx, id, prioridade); err != nil {
			fmt.Println(id, "erro:", err)
			return
		}
		mu.Lock()
		ordem = append(ordem, id)
		mu.Unlock()
		if id == "A" {
			<-soltar
		}
		time.Sleep(5 * time.Millisecond) // "pousando"
		torre.Liberar(id)
	}

	// A chega primeiro e ocupa a pista.
	wg.Add(1)
	go pouso("A", 1)
	aguardar(func() bool { return pousou() == 1 })

	// B, C (rotina) e D (emergencia) pedem enquanto A usa a pista. Esperamos cada um entrar na fila.
	for i, p := range []struct {
		id   string
		prio int
	}{{"B", 1}, {"C", 1}, {"D", 9}} {
		wg.Add(1)
		go pouso(p.id, p.prio)
		aguardar(func() bool { return torre.Esperando() == i+1 })
	}
	fmt.Println("aguardando na fila:", torre.Esperando())
	close(soltar)
	wg.Wait()
	fmt.Println("ordem de pouso:", ordem)
}
