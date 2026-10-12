package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type tarefa struct {
	indice int
	valor  int
}

type resultado struct {
	indice int
	saida  int
	err    error
}

var ErrMultiplo7 = errors.New("múltiplos de 7 não são permitidos")

func trabalho(n int) (int, error) {
	time.Sleep(20 * time.Millisecond)
	if n%7 == 0 {
		return 0, fmt.Errorf("tarefa %d: %w", n, ErrMultiplo7)
	}
	return n * n, nil
}

// Pool executa os valores com `workers` goroutines e devolve os resultados NA ORDEM DE ENTRADA.
func Pool(valores []int, workers int) []resultado {
	tarefas := make(chan tarefa)
	saidas := make([]resultado, len(valores))

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for t := range tarefas {
				v, err := trabalho(t.valor)
				saidas[t.indice] = resultado{t.indice, v, err} // cada índice é escrito por UM worker
			}
		})
	}

	for i, v := range valores {
		tarefas <- tarefa{i, v}
	}
	close(tarefas)
	wg.Wait()
	return saidas
}

func main() {
	valores := make([]int, 12)
	for i := range valores {
		valores[i] = i + 1
	}

	inicio := time.Now()
	for _, v := range valores {
		trabalho(v)
	}
	seq := time.Since(inicio)

	inicio = time.Now()
	res := Pool(valores, 4)
	par := time.Since(inicio)

	var erros []error
	var soma int
	for _, r := range res {
		if r.err != nil {
			erros = append(erros, r.err)
			continue
		}
		soma += r.saida
	}
	fmt.Println("soma:", soma, "| erros:", errors.Join(erros...))
	fmt.Printf("sequencial ~%v | pool de 4 ~%v | ganho > 3x: %t\n", seq.Round(10*time.Millisecond), par.Round(10*time.Millisecond), seq > 3*par)
	fmt.Println(res[0].saida, res[1].saida, res[2].saida)
}
