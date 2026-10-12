package main

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Evento struct {
	Nivel    string
	Servico  string
	Mensagem string
}

const logs = `2026-10-04T10:00:01Z INFO pagamentos autorizado pedido 1001
2026-10-04T10:00:02Z ERROR pagamentos timeout ao chamar gateway
linha corrompida sem formato
2026-10-04T10:00:03Z WARN estoque reposição atrasada
2026-10-04T10:00:04Z ERROR estoque divergência no item 77
2026-10-04T10:00:05Z ERROR pagamentos cartão recusado
2026-10-04T10:00:06Z INFO entrega rota calculada
2026-10-04T10:00:07Z ERROR entrega endereço inválido
2026-10-04T10:00:08Z ERROR pagamentos timeout ao chamar gateway`

// Estágio 1: lê linhas. Para se done for fechado.
func lerLinhas(done <-chan struct{}, r io.Reader) <-chan string {
	saida := make(chan string)
	go func() {
		defer close(saida)
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			select {
			case saida <- sc.Text():
			case <-done:
				return
			}
		}
	}()
	return saida
}

// Estágio 2: interpreta cada linha. Linhas inválidas são contadas e descartadas.
func interpretar(done <-chan struct{}, linhas <-chan string, invalidas *atomic.Int64) <-chan Evento {
	saida := make(chan Evento)
	go func() {
		defer close(saida)
		for l := range linhas {
			campos := strings.SplitN(l, " ", 4)
			if len(campos) < 4 {
				invalidas.Add(1)
				continue
			}
			if _, err := time.Parse(time.RFC3339, campos[0]); err != nil {
				invalidas.Add(1) // o primeiro campo precisa ser um horário válido
				continue
			}
			ev := Evento{Nivel: campos[1], Servico: campos[2], Mensagem: campos[3]}
			select {
			case saida <- ev:
			case <-done:
				return
			}
		}
	}()
	return saida
}

// Estágio 3: filtro.
func soErros(done <-chan struct{}, eventos <-chan Evento) <-chan Evento {
	saida := make(chan Evento)
	go func() {
		defer close(saida)
		for ev := range eventos {
			if ev.Nivel != "ERROR" {
				continue
			}
			select {
			case saida <- ev:
			case <-done:
				return
			}
		}
	}()
	return saida
}

// Estágio 4: enriquecimento com vários workers (fan-out) e junção (fan-in).
func enriquecer(done <-chan struct{}, erros <-chan Evento, workers int) <-chan string {
	saida := make(chan string)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for ev := range erros {
				time.Sleep(5 * time.Millisecond) // simula consulta a um catálogo de serviços
				alerta := fmt.Sprintf("[%s] %s", strings.ToUpper(ev.Servico), ev.Mensagem)
				select {
				case saida <- alerta:
				case <-done:
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(saida)
	}()
	return saida
}

func main() {
	base := runtime.NumGoroutine()
	done := make(chan struct{})

	var invalidas atomic.Int64
	alertas := enriquecer(done, soErros(done, interpretar(done, lerLinhas(done, strings.NewReader(logs)), &invalidas)), 3)

	contagem := map[string]int{}
	var todos []string
	for a := range alertas {
		todos = append(todos, a)
		servico := a[1:strings.Index(a, "]")]
		contagem[servico]++
	}
	close(done)

	sort.Strings(todos)
	for _, a := range todos {
		fmt.Println(a)
	}
	fmt.Println("linhas inválidas:", invalidas.Load())
	var servicos []string
	for s := range contagem {
		servicos = append(servicos, s)
	}
	sort.Strings(servicos)
	for _, s := range servicos {
		fmt.Printf("%-10s %d erro(s)\n", s, contagem[s])
	}

	time.Sleep(20 * time.Millisecond)
	fmt.Println("goroutines sobrando:", runtime.NumGoroutine()-base)
}
