package main

import (
	"fmt"
	"slices"
	"sync"
)

// Barramento e um Mediador de eventos implementado como Observer (a variante popular do catalogo):
// os componentes se inscrevem em topicos e publicam nele, sem se conhecerem.
type Barramento struct {
	mu        sync.RWMutex
	proximo   int
	inscritos map[string][]inscricao
}

type inscricao struct {
	id     int
	tratar func(evento any)
}

func NovoBarramento() *Barramento { return &Barramento{inscritos: map[string][]inscricao{}} }

// Inscrever devolve a funcao que CANCELA a inscricao (idempotente).
func (b *Barramento) Inscrever(topico string, tratar func(evento any)) (cancelar func()) {
	b.mu.Lock()
	b.proximo++
	id := b.proximo
	b.inscritos[topico] = append(b.inscritos[topico], inscricao{id, tratar})
	b.mu.Unlock()

	var uma sync.Once
	return func() {
		uma.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			b.inscritos[topico] = slices.DeleteFunc(b.inscritos[topico], func(i inscricao) bool { return i.id == id })
		})
	}
}

// Publicar entrega o evento a todos os inscritos NO TOPICO, de forma sincrona e em ordem de inscricao.
// Copia a lista sob a trava e entrega FORA dela: assim um tratador pode se inscrever, cancelar ou
// publicar de dentro de uma entrega sem causar deadlock.
func (b *Barramento) Publicar(topico string, evento any) (entregues int, err error) {
	b.mu.RLock()
	copia := slices.Clone(b.inscritos[topico])
	b.mu.RUnlock()

	for _, i := range copia {
		func() {
			defer func() {
				if r := recover(); r != nil { // um tratador com panic nao derruba os outros
					err = fmt.Errorf("tratador %d do topico %q entrou em panic: %v", i.id, topico, r)
				}
			}()
			i.tratar(evento)
			entregues++
		}()
	}
	return entregues, err
}

func (b *Barramento) Inscritos(topico string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.inscritos[topico])
}
