package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unique"
	"unsafe"
)

const nRegistros = 1_000_000

var cidades = func() []string {
	c := make([]string, 100)
	for i := range c {
		c[i] = "Municipio de Exemplo numero " + strconv.Itoa(i) + " do Estado Federado"
	}
	return c
}()

// Cada registro "lido do arquivo" recebe uma copia propria do texto (como acontece ao decodificar CSV/JSON).
func lerCidade(i int) string { return strings.Clone(cidades[i%100]) }

type RegistroTexto struct {
	ID     int32
	Cidade string
}

type RegistroUnique struct {
	ID     int32
	Cidade unique.Handle[string]
}

type Internador struct {
	mu sync.Mutex
	m  map[string]string
}

func (in *Internador) Interna(s string) string {
	in.mu.Lock()
	defer in.mu.Unlock()
	if c, ok := in.m[s]; ok {
		return c
	}
	in.m[s] = s
	return s
}

type RegistroManual = RegistroTexto

func medir(f func() any) (any, uint64) {
	var a, d runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&a)
	v := f()
	runtime.GC()
	runtime.ReadMemStats(&d)
	return v, d.HeapAlloc - a.HeapAlloc
}

func main() {
	v1, b1 := medir(func() any {
		r := make([]RegistroTexto, nRegistros)
		for i := range r {
			r[i] = RegistroTexto{int32(i), lerCidade(i)}
		}
		return r
	})
	v2, b2 := medir(func() any {
		r := make([]RegistroUnique, nRegistros)
		for i := range r {
			r[i] = RegistroUnique{int32(i), unique.Make(lerCidade(i))}
		}
		return r
	})
	in := &Internador{m: map[string]string{}}
	v3, b3 := medir(func() any {
		r := make([]RegistroManual, nRegistros)
		for i := range r {
			r[i] = RegistroManual{int32(i), in.Interna(lerCidade(i))}
		}
		return r
	})

	fmt.Printf("%d registros, 100 cidades distintas de %d bytes\n", nRegistros, len(cidades[0]))
	fmt.Printf("string em cada registro:  %7.1f MiB  (struct de %d bytes)\n", float64(b1)/(1<<20), unsafe.Sizeof(RegistroTexto{}))
	fmt.Printf("unique.Handle[string]:    %7.1f MiB  (struct de %d bytes)\n", float64(b2)/(1<<20), unsafe.Sizeof(RegistroUnique{}))
	fmt.Printf("internador com mapa:      %7.1f MiB  (struct de %d bytes)\n", float64(b3)/(1<<20), unsafe.Sizeof(RegistroManual{}))

	// igualdade
	h1, h2 := v2.([]RegistroUnique)[0].Cidade, v2.([]RegistroUnique)[100].Cidade
	fmt.Println("Handles da mesma cidade sao iguais (comparacao de ponteiro):", h1 == h2, "| valor:", h1.Value()[:20]+"...")
	_ = v1
	_ = v3
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
	runtime.KeepAlive(v3)
}
