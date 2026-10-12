package main

import (
	"fmt"
	"sync"
)

// Ingenuo guarda o último total recebido, na ordem em que os eventos chegam.
type Ingenuo struct {
	mu    sync.Mutex
	total int
}

func (p *Ingenuo) Atualizar(m Mudanca) {
	p.mu.Lock()
	p.total = m.Total
	p.mu.Unlock()
}

// Versionado ignora o evento que chega com versão menor que a já vista.
type Versionado struct {
	mu        sync.Mutex
	versao    uint64
	total     int
	atrasados int
}

func (p *Versionado) Atualizar(m Mudanca) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if m.Versao > p.versao {
		p.versao, p.total = m.Versao, m.Total
		return
	}
	p.atrasados++
}

func main() {
	const rodadas = 300
	const goroutines = 8
	const adicoes = 100

	errosIngenuo, errosVersionado, atrasados, rodadasComAtraso := 0, 0, 0, 0
	for range rodadas {
		o := &Orcamento{}
		ing, ver := &Ingenuo{}, &Versionado{}
		o.Observar(ing.Atualizar)
		o.Observar(ver.Atualizar)

		var wg sync.WaitGroup
		for range goroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range adicoes {
					_ = o.Adicionar("x", 1)
				}
			}()
		}
		wg.Wait()

		atrasados += ver.atrasados
		if ver.atrasados > 0 {
			rodadasComAtraso++
		}
		real := o.Total()
		if ing.total != real {
			errosIngenuo++
		}
		if ver.total != real {
			errosVersionado++
		}
	}
	fmt.Printf("%d rodadas de %d goroutines x %d adições\n", rodadas, goroutines, adicoes)
	fmt.Printf("eventos entregues fora de ordem: %d de %d (%d rodadas com pelo menos um)\n", atrasados, rodadas*goroutines*adicoes, rodadasComAtraso)
	fmt.Printf("painel ingênuo terminou ERRADO em %d rodadas\n", errosIngenuo)
	fmt.Printf("painel versionado terminou errado em %d rodadas\n", errosVersionado)
}
