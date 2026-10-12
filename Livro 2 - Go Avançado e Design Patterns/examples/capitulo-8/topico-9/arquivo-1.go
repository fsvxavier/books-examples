package main

import (
	"bytes"
	"fmt"
	"sync"
)

// ---------- Flyweight: UMA instancia, imutavel, COMPARTILHADA AO MESMO TEMPO ----------

type Moeda struct {
	codigo  string
	simbolo string
	casas   int
}

func (m *Moeda) Formatar(centavos int64) string {
	return fmt.Sprintf("%s %d,%0*d", m.simbolo, centavos/100, m.casas, centavos%100)
}

var brl = &Moeda{"BRL", "R$", 2}

// ---------- sync.Pool: instancias MUTAVEIS, de USO EXCLUSIVO, devolvidas depois ----------

var buffers = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func relatorio(id int, valor int64) string {
	b := buffers.Get().(*bytes.Buffer) // uso exclusivo: ninguem mais enxerga este buffer agora
	defer func() {
		b.Reset() // limpar ANTES de devolver
		buffers.Put(b)
	}()
	fmt.Fprintf(b, "pedido %d: %s", id, brl.Formatar(valor)) // o flyweight e lido por todos, sem trava
	return b.String()
}

func main() {
	var wg sync.WaitGroup
	saidas := make([]string, 6)
	for i := range saidas {
		wg.Go(func() { saidas[i] = relatorio(i+1, int64(1000*(i+1)+5)) })
	}
	wg.Wait()
	for _, s := range saidas {
		fmt.Println(s)
	}
	fmt.Println("o Flyweight (Moeda) foi lido por 6 goroutines ao mesmo tempo, sem trava")
	fmt.Println("cada bytes.Buffer do Pool foi usado por UMA goroutine de cada vez")
}
