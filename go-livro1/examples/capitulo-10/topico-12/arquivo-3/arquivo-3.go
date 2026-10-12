package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Resposta struct {
	ID    int
	Dados string
	Err   error
}

var ErrServidor = errors.New("servidor indisponível")

// chamarAPI simula uma API externa que falha para alguns IDs.
func chamarAPI(id int) (string, error) {
	time.Sleep(5 * time.Millisecond)
	if id%6 == 0 {
		return "", fmt.Errorf("id %d: %w", id, ErrServidor)
	}
	return fmt.Sprintf("dados-%d", id), nil
}

// ProcessarLote consulta os IDs com `workers` goroutines e no máximo 1 chamada por `intervalo`
// (limite de taxa COMPARTILHADO por todos os workers). Os resultados voltam na ordem dos IDs.
func ProcessarLote(ids []int, workers int, intervalo time.Duration) []Resposta {
	var limite <-chan time.Time
	if intervalo > 0 {
		t := time.NewTicker(intervalo)
		defer t.Stop()
		limite = t.C
	}

	tarefas := make(chan int)
	respostas := make([]Resposta, len(ids))
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for i := range tarefas {
				if limite != nil {
					<-limite // espera a próxima "vaga" de chamada, dividida entre todos
				}
				dados, err := chamarAPI(ids[i])
				respostas[i] = Resposta{ids[i], dados, err}
			}
		})
	}
	for i := range ids {
		tarefas <- i
	}
	close(tarefas)
	wg.Wait()
	return respostas
}

func main() {
	ids := make([]int, 12)
	for i := range ids {
		ids[i] = i + 1
	}

	inicio := time.Now()
	livre := ProcessarLote(ids, 4, 0)
	tLivre := time.Since(inicio)

	inicio = time.Now()
	limitado := ProcessarLote(ids, 4, 20*time.Millisecond)
	tLimitado := time.Since(inicio)

	var erros []error
	for _, r := range limitado {
		if r.Err != nil {
			erros = append(erros, r.Err)
		}
	}
	fmt.Println("mesmos resultados:", fmt.Sprint(livre) == fmt.Sprint(limitado))
	fmt.Println("erros:", errors.Join(erros...))
	fmt.Println("primeiros:", limitado[0].Dados, limitado[1].Dados, "| ordem preservada:", limitado[11].ID == 12)
	fmt.Printf("sem limite: ~%v | com 1 chamada a cada 20ms: >= 200ms: %t (sem limite < 50ms: %t)\n",
		tLivre.Round(10*time.Millisecond), tLimitado >= 200*time.Millisecond, tLivre < 50*time.Millisecond)
}
