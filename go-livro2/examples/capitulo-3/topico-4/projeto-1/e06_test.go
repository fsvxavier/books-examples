package e06

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"maps"
	"slices"
	"testing"
)

type Item struct {
	SKU string
	Qtd int
}

type Modelo struct {
	Nome   string
	Tags   []string
	Metas  map[string]string
	Itens  []Item
	Padrao *Item
}

func novoModelo() *Modelo {
	return &Modelo{
		Nome:   "modelo-base",
		Tags:   []string{"a", "b", "c", "d"},
		Metas:  map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
		Itens:  []Item{{"A", 1}, {"B", 2}, {"C", 3}},
		Padrao: &Item{"Z", 9},
	}
}

func (m *Modelo) Clone() *Modelo {
	c := *m
	c.Tags = slices.Clone(m.Tags)
	c.Metas = maps.Clone(m.Metas)
	c.Itens = slices.Clone(m.Itens)
	if m.Padrao != nil {
		p := *m.Padrao
		c.Padrao = &p
	}
	return &c
}

func cloneJSON(m *Modelo) *Modelo {
	b, _ := json.Marshal(m)
	var c Modelo
	json.Unmarshal(b, &c)
	return &c
}

func cloneGob(m *Modelo) *Modelo {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(m)
	var c Modelo
	gob.NewDecoder(&buf).Decode(&c)
	return &c
}

var sink *Modelo

func BenchmarkReconstruir(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = novoModelo()
	}
}

func BenchmarkCloneManual(b *testing.B) {
	m := novoModelo()
	b.ReportAllocs()
	for b.Loop() {
		sink = m.Clone()
	}
}

func BenchmarkCloneJSON(b *testing.B) {
	m := novoModelo()
	b.ReportAllocs()
	for b.Loop() {
		sink = cloneJSON(m)
	}
}

func BenchmarkCloneGob(b *testing.B) {
	m := novoModelo()
	b.ReportAllocs()
	for b.Loop() {
		sink = cloneGob(m)
	}
}

func TestClonesSaoIndependentes(t *testing.T) {
	for nome, f := range map[string]func(*Modelo) *Modelo{
		"manual": (*Modelo).Clone, "json": cloneJSON, "gob": cloneGob,
	} {
		t.Run(nome, func(t *testing.T) {
			orig := novoModelo()
			c := f(orig)
			c.Tags[0] = "X"
			c.Metas["k1"] = "X"
			c.Itens[0].Qtd = 99
			c.Padrao.Qtd = 99
			if orig.Tags[0] != "a" || orig.Metas["k1"] != "v1" || orig.Itens[0].Qtd != 1 || orig.Padrao.Qtd != 9 {
				t.Fatalf("o clone alterou o original: %+v", orig)
			}
		})
	}
}
