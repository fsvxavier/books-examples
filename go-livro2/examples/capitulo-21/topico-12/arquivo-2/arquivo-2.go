package pool

import (
	"bytes"
	"encoding/json"
	"sync"
)

type Resposta struct {
	ID    int      `json:"id"`
	Nome  string   `json:"nome"`
	Tags  []string `json:"tags"`
	Preco float64  `json:"preco"`
}

var exemplo = Resposta{ID: 7, Nome: "caderno universitario", Tags: []string{"escola", "papel", "oferta"}, Preco: 19.9}

// SemPool aloca um buffer novo a cada chamada.
func SemPool() int {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(exemplo)
	return buf.Len()
}

var poolBuffer = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// ComPool reaproveita *bytes.Buffer: o ponteiro cabe na interface sem alocar.
func ComPool() int {
	buf := poolBuffer.Get().(*bytes.Buffer)
	buf.Reset()
	json.NewEncoder(buf).Encode(exemplo)
	n := buf.Len()
	poolBuffer.Put(buf)
	return n
}

var poolSlice = sync.Pool{New: func() any { return make([]byte, 0, 512) }}

var conteudo = []byte("conteudo de exemplo com algum tamanho, para encher o buffer um pouco")

// SlicePorValor guarda um []byte no pool: cada Put precisa embrulhar o
// cabecalho do slice (3 palavras) em uma interface, e isso aloca.
func SlicePorValor() int {
	b := poolSlice.Get().([]byte)[:0]
	b = append(b, conteudo...)
	n := len(b)
	poolSlice.Put(b)
	return n
}

var poolPtr = sync.Pool{New: func() any { b := make([]byte, 0, 512); return &b }}

// SlicePorPonteiro guarda um *[]byte: o ponteiro cabe na interface sem alocar.
func SlicePorPonteiro() int {
	p := poolPtr.Get().(*[]byte)
	b := append((*p)[:0], conteudo...)
	n := len(b)
	*p = b
	poolPtr.Put(p)
	return n
}
