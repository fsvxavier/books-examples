package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Pedido struct {
	ID    int
	Valor int64 // centavos
}

type Resultado struct {
	PedidoID int
	Status   string
}

// processar consome pedidos até o canal fechar. Vários workers podem rodar esta mesma função.
func processar(id int, pedidos <-chan Pedido, resultados chan<- Resultado, wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range pedidos {
		time.Sleep(time.Duration(p.ID%3+1) * 10 * time.Millisecond) // trabalho de duração variável
		status := "aprovado"
		if p.Valor > 100_000 {
			status = "em análise"
		}
		resultados <- Resultado{p.ID, status}
	}
}

func main() {
	pedidos := make(chan Pedido)
	resultados := make(chan Resultado)

	var wg sync.WaitGroup
	const workers = 3
	for id := 1; id <= workers; id++ {
		wg.Add(1)
		go processar(id, pedidos, resultados, &wg)
	}

	// quem envia pedidos fecha o canal de pedidos; outra goroutine fecha os resultados
	go func() {
		for i := 1; i <= 8; i++ {
			pedidos <- Pedido{ID: i, Valor: int64(i) * 30_000}
		}
		close(pedidos)
	}()
	go func() {
		wg.Wait()
		close(resultados)
	}()

	var todos []Resultado
	for r := range resultados {
		todos = append(todos, r)
	}
	sort.Slice(todos, func(i, j int) bool { return todos[i].PedidoID < todos[j].PedidoID })
	for _, r := range todos {
		fmt.Printf("pedido %d: %s\n", r.PedidoID, r.Status)
	}
}
