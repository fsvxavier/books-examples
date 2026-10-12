package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

type Conta struct {
	mu    sync.Mutex
	id    int
	saldo int
}

// Transferir trava as duas contas SEMPRE na mesma ordem (pelo menor id primeiro).
// Como todos seguem a mesma ordem, não existe espera circular, logo não há deadlock.
func Transferir(de, para *Conta, valor int) bool {
	if de == para {
		return false
	}
	primeira, segunda := de, para
	if para.id < de.id {
		primeira, segunda = para, de
	}
	primeira.mu.Lock()
	defer primeira.mu.Unlock()
	segunda.mu.Lock()
	defer segunda.mu.Unlock()

	if de.saldo < valor {
		return false
	}
	de.saldo -= valor
	para.saldo += valor
	return true
}

func main() {
	contas := make([]*Conta, 5)
	for i := range contas {
		contas[i] = &Conta{id: i, saldo: 1000}
	}

	concluiu := make(chan struct{})
	var feitas, recusadas int
	var mu sync.Mutex
	go func() {
		var wg sync.WaitGroup
		for range 50 {
			wg.Go(func() {
				for range 200 {
					de, para := contas[rand.IntN(5)], contas[rand.IntN(5)]
					ok := Transferir(de, para, rand.IntN(300))
					mu.Lock()
					if ok {
						feitas++
					} else {
						recusadas++
					}
					mu.Unlock()
				}
			})
		}
		wg.Wait()
		close(concluiu)
	}()

	select {
	case <-concluiu:
	case <-time.After(5 * time.Second):
		fmt.Println("deadlock!")
		return
	}

	total := 0
	for _, c := range contas {
		total += c.saldo
	}
	fmt.Println("tentativas:", feitas+recusadas)
	fmt.Println("total preservado:", total == 5000)
}
