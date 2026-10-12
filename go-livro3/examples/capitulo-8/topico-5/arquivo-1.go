// Package anel implementa all-reduce em anel (reduce-scatter + all-gather)
// sobre qualquer transporte que entregue mensagens em ordem.
package anel

import (
	"fmt"
	"sync/atomic"
)

// Elo é a visão que um worker tem do anel: manda para o vizinho da direita
// e recebe do vizinho da esquerda.
type Elo interface {
	Enviar(v []float64) error
	Receber(dst []float64) error // preenche dst com exatamente len(dst) valores
}

// Limites dos pedaços: o vetor de n valores é dividido em w pedaços quase iguais.
func limites(n, w, i int) (ini, fim int) {
	base, resto := n/w, n%w
	ini = i*base + min(i, resto)
	fim = ini + base
	if i < resto {
		fim++
	}
	return ini, fim
}

// Somar substitui v, em todos os workers, pela soma elemento a elemento dos
// vetores de todos. rank vai de 0 a w-1. Cada worker envia 2(w-1)/w do vetor.
func Somar(elo Elo, rank, w int, v []float64) error {
	if w == 1 {
		return nil
	}
	tmp := make([]float64, (len(v)+w-1)/w)
	// Fase 1, reduce-scatter: no passo s o worker envia o pedaço (rank-s) e
	// acumula o pedaço (rank-s-1) que chega. Ao final, o worker "rank" é dono
	// da soma completa do pedaço (rank+1) mod w.
	for s := 0; s < w-1; s++ {
		ie, fe := limites(len(v), w, ((rank-s)%w+w)%w)
		ir, fr := limites(len(v), w, ((rank-s-1)%w+w)%w)
		if err := trocar(elo, v[ie:fe], tmp[:fr-ir]); err != nil {
			return fmt.Errorf("reduce-scatter passo %d: %w", s, err)
		}
		for i, x := range tmp[:fr-ir] {
			v[ir+i] += x
		}
	}
	// Fase 2, all-gather: os pedaços já somados circulam e substituem.
	for s := 0; s < w-1; s++ {
		ie, fe := limites(len(v), w, ((rank+1-s)%w+w)%w)
		ir, fr := limites(len(v), w, ((rank-s)%w+w)%w)
		if err := trocar(elo, v[ie:fe], v[ir:fr]); err != nil {
			return fmt.Errorf("all-gather passo %d: %w", s, err)
		}
	}
	return nil
}

// trocar envia e recebe ao mesmo tempo. Se todos enviassem de forma síncrona
// antes de receber, um transporte sem buffer travaria; por isso o envio vai
// em uma goroutine.
func trocar(elo Elo, env, rec []float64) error {
	erro := make(chan error, 1)
	go func() { erro <- elo.Enviar(env) }()
	if err := elo.Receber(rec); err != nil {
		<-erro
		return err
	}
	return <-erro
}

// Memoria é um anel de w workers ligados por canais em memória.
type Memoria struct {
	canais []chan []float64
	Bytes  []atomic.Int64 // bytes enviados por cada rank
}

// NovoMemoria cria os w elos. buffer é a capacidade de cada canal.
func NovoMemoria(w, buffer int) *Memoria {
	m := &Memoria{canais: make([]chan []float64, w), Bytes: make([]atomic.Int64, w)}
	for i := range m.canais {
		m.canais[i] = make(chan []float64, buffer)
	}
	return m
}

// Elo devolve a visão do worker rank.
func (m *Memoria) Elo(rank int) Elo { return &eloMem{m: m, rank: rank} }

type eloMem struct {
	m    *Memoria
	rank int
}

func (e *eloMem) Enviar(v []float64) error {
	c := make([]float64, len(v)) // copia: o remetente continua dono do original
	copy(c, v)
	e.m.Bytes[e.rank].Add(int64(8 * len(v)))
	e.m.canais[(e.rank+1)%len(e.m.canais)] <- c
	return nil
}

func (e *eloMem) Receber(dst []float64) error {
	msg := <-e.m.canais[e.rank]
	if len(msg) != len(dst) {
		return fmt.Errorf("mensagem de %d valores, esperava %d", len(msg), len(dst))
	}
	copy(dst, msg)
	return nil
}
