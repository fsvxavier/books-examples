package janela

import (
	"sort"
	"sync"
	"time"
)

// evento é uma compra de uma entidade em um instante.
type evento struct {
	quando time.Time
	centav int64
}

// Ingenua guarda todos os eventos de cada entidade e reconstrói a lista a
// cada evento novo.
type Ingenua struct {
	mu      sync.Mutex
	janela  time.Duration
	eventos map[string][]evento
}

// NovaIngenua cria a janela ingênua de duração j.
func NovaIngenua(j time.Duration) *Ingenua {
	return &Ingenua{janela: j, eventos: map[string][]evento{}}
}

func (w *Ingenua) Adicionar(entidade string, quando time.Time, centav int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	corte := quando.Add(-w.janela)
	validos := make([]evento, 0, len(w.eventos[entidade])+1)
	for _, ev := range w.eventos[entidade] {
		if ev.quando.After(corte) {
			validos = append(validos, ev)
		}
	}
	validos = append(validos, evento{quando, centav})
	w.eventos[entidade] = validos
}

func (w *Ingenua) Consultar(entidade string, agora time.Time) Agregado {
	w.mu.Lock()
	defer w.mu.Unlock()

	corte := agora.Add(-w.janela)
	var a Agregado
	for _, ev := range w.eventos[entidade] {
		if ev.quando.After(corte) && !ev.quando.After(agora) {
			a.Contagem++
			a.Centav += ev.centav
		}
	}
	return a
}

// fila é a lista de eventos ordenada por tempo; inicio marca o primeiro que ainda conta.
type fila struct {
	ev     []evento
	inicio int
}

// Exata guarda os eventos em uma fila ordenada e descarta do começo o que saiu da janela.
type Exata struct {
	mu     sync.Mutex
	janela time.Duration
	filas  map[string]*fila
}

// NovaExata cria a janela exata de duração j.
func NovaExata(j time.Duration) *Exata {
	return &Exata{janela: j, filas: map[string]*fila{}}
}

func (w *Exata) Adicionar(entidade string, quando time.Time, centav int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	f := w.filas[entidade]
	if f == nil {
		f = &fila{}
		w.filas[entidade] = f
	}
	ev := evento{quando, centav}
	n := len(f.ev)
	if n == f.inicio || !quando.Before(f.ev[n-1].quando) {
		f.ev = append(f.ev, ev) // caso comum: chegou em ordem
	} else { // chegou atrasado: acha o lugar
		i := f.inicio + sort.Search(n-f.inicio, func(k int) bool {
			return f.ev[f.inicio+k].quando.After(quando)
		})
		f.ev = append(f.ev, evento{})
		copy(f.ev[i+1:], f.ev[i:])
		f.ev[i] = ev
	}

	// descarta o que saiu da janela em relação a este evento
	corte := quando.Add(-w.janela)
	for f.inicio < len(f.ev) && !f.ev[f.inicio].quando.After(corte) {
		f.inicio++
	}
	// compacta quando metade da fatia é lixo, para não vazar memória
	if f.inicio > 32 && f.inicio > len(f.ev)/2 {
		f.ev = append(f.ev[:0], f.ev[f.inicio:]...)
		f.inicio = 0
	}
}

func (w *Exata) Consultar(entidade string, agora time.Time) Agregado {
	w.mu.Lock()
	defer w.mu.Unlock()

	f := w.filas[entidade]
	if f == nil {
		return Agregado{}
	}
	corte := agora.Add(-w.janela)
	var a Agregado
	// os eventos já são ordenados: percorre só os que podem contar
	i := f.inicio + sort.Search(len(f.ev)-f.inicio, func(k int) bool {
		return f.ev[f.inicio+k].quando.After(corte)
	})
	for ; i < len(f.ev) && !f.ev[i].quando.After(agora); i++ {
		a.Contagem++
		a.Centav += f.ev[i].centav
	}
	return a
}

// NumBaldes é a quantidade de baldes em que cada janela é dividida.
const NumBaldes = 60

// anel guarda os baldes de uma entidade; tag registra o índice absoluto de cada posição.
type anel struct {
	tag   [NumBaldes]int64
	cont  [NumBaldes]int32
	centv [NumBaldes]int64
}

// Baldes divide a janela em NumBaldes baldes de largura fixa, em um anel por entidade.
type Baldes struct {
	mu      sync.Mutex
	largura int64 // largura de um balde, em nanossegundos
	aneis   map[string]*anel
}

// NovaBaldes cria a janela de baldes de duração j.
func NovaBaldes(j time.Duration) *Baldes {
	return &Baldes{largura: int64(j) / NumBaldes, aneis: map[string]*anel{}}
}

func (w *Baldes) Adicionar(entidade string, quando time.Time, centav int64) {
	idx := quando.UnixNano() / w.largura
	pos := int(((idx % NumBaldes) + NumBaldes) % NumBaldes)

	w.mu.Lock()
	defer w.mu.Unlock()
	a := w.aneis[entidade]
	if a == nil {
		a = &anel{}
		w.aneis[entidade] = a
	}
	switch {
	case a.tag[pos] > idx:
		return // a posição já tem um balde mais novo: o evento é velho demais
	case a.tag[pos] < idx:
		a.tag[pos], a.cont[pos], a.centv[pos] = idx, 0, 0 // reaproveita o balde vencido
	}
	a.cont[pos]++
	a.centv[pos] += centav
}

func (w *Baldes) Consultar(entidade string, agora time.Time) Agregado {
	idxAgora := agora.UnixNano() / w.largura

	w.mu.Lock()
	defer w.mu.Unlock()
	a := w.aneis[entidade]
	if a == nil {
		return Agregado{}
	}
	var r Agregado
	for pos := 0; pos < NumBaldes; pos++ {
		if t := a.tag[pos]; t > idxAgora-NumBaldes && t <= idxAgora {
			r.Contagem += int64(a.cont[pos])
			r.Centav += a.centv[pos]
		}
	}
	return r
}
