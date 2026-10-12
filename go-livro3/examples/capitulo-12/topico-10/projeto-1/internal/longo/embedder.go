package longo

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

// Dimensao é o tamanho padrão dos vetores do Hash.
const Dimensao = 256

// Embedder transforma um texto em um vetor. É a fronteira com o modelo de
// embeddings: o falso (Hash) e um real implementam a mesma interface.
type Embedder interface {
	Embed(ctx context.Context, texto string) ([]float32, error)
}

// Hash é o "truque do hash" (feature hashing): cada palavra cai em uma posição
// do vetor, escolhida por uma função de hash, com sinal + ou -. Não entende
// sentido nenhum, e é de propósito: serve para testar sem um modelo.
type Hash struct{ Dim int }

func (h Hash) Embed(_ context.Context, texto string) ([]float32, error) {
	dim := h.Dim
	if dim == 0 {
		dim = Dimensao
	}
	v := make([]float32, dim)
	palavras := strings.FieldsFunc(strings.ToLower(texto), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, p := range palavras {
		if len(p) < 3 {
			continue // ignora "de", "o", "a"...
		}
		f := fnv.New32a()
		f.Write([]byte(p))
		s := f.Sum32()
		sinal := float32(1)
		if s&1 == 1 {
			sinal = -1
		}
		v[(s>>1)%uint32(dim)] += sinal
	}
	var norma float64
	for _, x := range v {
		norma += float64(x) * float64(x)
	}
	if norma == 0 {
		v[0] = 1 // texto sem palavras úteis: um vetor qualquer, mas válido
		return v, nil
	}
	norma = math.Sqrt(norma)
	for i := range v {
		v[i] = float32(float64(v[i]) / norma)
	}
	return v, nil
}
