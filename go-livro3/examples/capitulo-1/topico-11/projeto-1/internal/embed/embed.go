// Package embed define o que é um gerador de embeddings e traz um falso,
// determinístico, para os experimentos.
package embed

import (
	"context"
	"hash/fnv"
	"math"

	"exemplo.com/rag/internal/texto"
)

// Embedder transforma textos em vetores de tamanho fixo. Em produção, é uma
// chamada de rede a um modelo; por isso recebe um contexto e aceita lotes.
type Embedder interface {
	Embed(ctx context.Context, textos []string) ([][]float32, error)
	Dim() int
}

// Hash é o "truque do hash": cada palavra cai em uma posição do vetor,
// escolhida por uma função de hash, com sinal + ou -. Textos que dividem
// palavras ficam parecidos; textos com palavras diferentes, mesmo que
// signifiquem o mesmo, não. Não entende sentido nenhum, e é de propósito.
type Hash struct{ Dimensao int }

func (h Hash) Dim() int { return h.Dimensao }

func (h Hash) Embed(ctx context.Context, textos []string) ([][]float32, error) {
	saida := make([][]float32, len(textos))
	for i, t := range textos {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		saida[i] = h.um(t)
	}
	return saida, nil
}

func (h Hash) um(t string) []float32 {
	v := make([]float32, h.Dimensao)
	for _, tok := range texto.Tokens(t) {
		f := fnv.New64a()
		f.Write([]byte(tok))
		s := f.Sum64()
		pos := int(s % uint64(h.Dimensao))
		if (s>>63)&1 == 1 {
			v[pos]--
		} else {
			v[pos]++
		}
	}
	Normalizar(v)
	return v
}

// Normalizar deixa o vetor com comprimento 1 (norma L2). Com vetores
// normalizados, o produto interno é igual ao cosseno.
func Normalizar(v []float32) {
	var soma float64
	for _, x := range v {
		soma += float64(x) * float64(x)
	}
	if soma == 0 {
		return
	}
	n := float32(1 / math.Sqrt(soma))
	for i := range v {
		v[i] *= n
	}
}

// Cosseno devolve o produto interno de dois vetores já normalizados.
func Cosseno(a, b []float32) float64 {
	var s float32
	for i := range a {
		s += a[i] * b[i]
	}
	return float64(s)
}
