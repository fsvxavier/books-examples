// Command e12 usa interfaces genéricas e incorpora (embedding) uma interface genérica em uma struct.
package main

import (
	"fmt"
	"sort"
)

// Armazem é uma interface genérica: o contrato depende dos tipos K e V.
type Armazem[K comparable, V any] interface {
	Obter(K) (V, bool)
	Guardar(K, V)
	Chaves() []K
}

// Mapa é uma implementação em memória.
type Mapa[K comparable, V any] struct{ m map[K]V }

// NovoMapa cria um Mapa vazio.
func NovoMapa[K comparable, V any]() *Mapa[K, V] { return &Mapa[K, V]{m: map[K]V{}} }

// Obter devolve o valor da chave.
func (a *Mapa[K, V]) Obter(k K) (V, bool) { v, ok := a.m[k]; return v, ok }

// Guardar grava o valor.
func (a *Mapa[K, V]) Guardar(k K, v V) { a.m[k] = v }

// Chaves devolve as chaves (em ordem qualquer).
func (a *Mapa[K, V]) Chaves() []K {
	s := make([]K, 0, len(a.m))
	for k := range a.m {
		s = append(s, k)
	}
	return s
}

// Contador decora qualquer Armazem: embute a interface genérica e conta as buscas.
type Contador[K comparable, V any] struct {
	Armazem[K, V]
	Buscas, Acertos int
}

// Obter intercepta a busca e delega ao armazém embutido.
func (c *Contador[K, V]) Obter(k K) (V, bool) {
	c.Buscas++
	v, ok := c.Armazem.Obter(k)
	if ok {
		c.Acertos++
	}
	return v, ok
}

// Copiar passa tudo de um armazém para outro. K e V são inferidos das duas interfaces.
func Copiar[K comparable, V any](destino, origem Armazem[K, V]) {
	for _, k := range origem.Chaves() {
		v, _ := origem.Obter(k)
		destino.Guardar(k, v)
	}
}

func main() {
	origem := NovoMapa[string, int]()
	origem.Guardar("a", 1)
	origem.Guardar("b", 2)

	destino := &Contador[string, int]{Armazem: NovoMapa[string, int]()}
	Copiar(destino, origem) // K e V são inferidos: *Contador[string,int] satisfaz Armazem[string,int]

	chaves := destino.Chaves() // promovido do Armazem embutido
	sort.Strings(chaves)
	_, _ = destino.Obter("a")
	_, _ = destino.Obter("zzz")
	fmt.Println("chaves:", chaves, "| buscas:", destino.Buscas, "| acertos:", destino.Acertos)
}
