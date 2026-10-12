package janela

import (
	"hash/maphash"
	"time"
)

// Fragmentada reparte as entidades entre várias janelas independentes, cada
// uma com a sua trava. Duas entidades em fragmentos diferentes não disputam
// o mesmo Mutex.
type Fragmentada struct {
	semente maphash.Seed
	partes  []Janela
}

func Fragmentar(n int, nova func() Janela) *Fragmentada {
	f := &Fragmentada{semente: maphash.MakeSeed(), partes: make([]Janela, n)}
	for i := range f.partes {
		f.partes[i] = nova()
	}
	return f
}

func (f *Fragmentada) parte(entidade string) Janela {
	return f.partes[maphash.String(f.semente, entidade)%uint64(len(f.partes))]
}

func (f *Fragmentada) Adicionar(entidade string, quando time.Time, centav int64) {
	f.parte(entidade).Adicionar(entidade, quando, centav)
}

func (f *Fragmentada) Consultar(entidade string, agora time.Time) Agregado {
	return f.parte(entidade).Consultar(entidade, agora)
}
