// Package indice traz os dois jeitos clássicos de buscar texto: por vetores
// (semântico) e por palavras (lexical, com BM25). Os dois escondem-se atrás
// da mesma interface, Buscador, o que permite combiná-los.
package indice

import (
	"context"
	"math"
	"sort"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-1/topico-11/projeto-1/internal/embed"
	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-1/topico-11/projeto-1/internal/texto"
)

// Doc é um trecho indexado.
type Doc struct {
	ID    string
	Fonte string
	Texto string
}

// Resultado é um documento com a pontuação que a fonte deu a ele.
// Pontuações de fontes diferentes NÃO são comparáveis entre si.
type Resultado struct {
	Doc    Doc
	Pontos float64
}

// Buscador é qualquer coisa que responde "quais os k melhores documentos
// para esta consulta?": um índice em memória, o pgvector, o Qdrant...
type Buscador interface {
	Nome() string
	Buscar(ctx context.Context, consulta string, k int) ([]Resultado, error)
}

// antes diz se a deve vir antes de b: pontuação maior primeiro e, no empate,
// o identificador menor. Sem o desempate, a ordem de documentos com a mesma
// pontuação dependeria da ordem de iteração de um map, que o Go sorteia de
// propósito a cada execução: a mesma consulta daria respostas diferentes.
func antes(a, b Resultado) bool {
	if a.Pontos != b.Pontos {
		return a.Pontos > b.Pontos
	}
	return a.Doc.ID < b.Doc.ID
}

// melhores guarda os k melhores resultados vistos, em ordem. Para k pequeno,
// a inserção ordenada vence um heap e um sort completo.
type melhores struct {
	k   int
	res []Resultado
}

func (m *melhores) oferecer(r Resultado) {
	if len(m.res) == m.k && !antes(r, m.res[len(m.res)-1]) {
		return
	}
	i := sort.Search(len(m.res), func(i int) bool { return antes(r, m.res[i]) })
	if len(m.res) < m.k {
		m.res = append(m.res, Resultado{})
	}
	copy(m.res[i+1:], m.res[i:])
	m.res[i] = r
}

// Vetorial guarda um vetor por documento e compara por cosseno, varrendo
// todos (busca exata, ou "força bruta"). É imutável depois de construído,
// então várias goroutines podem buscar ao mesmo tempo sem trava.
type Vetorial struct {
	// Minimo é o piso de similaridade: abaixo dele o documento não entra na
	// resposta. Sem piso, a busca por vizinhos SEMPRE devolve k documentos,
	// mesmo para uma pergunta que não tem nada a ver com a base.
	Minimo float64

	emb  embed.Embedder
	docs []Doc
	vecs [][]float32
}

// NovoVetorial calcula os embeddings de todos os documentos, em lotes.
func NovoVetorial(ctx context.Context, emb embed.Embedder, docs []Doc, lote int) (*Vetorial, error) {
	v := &Vetorial{emb: emb, docs: docs, vecs: make([][]float32, 0, len(docs))}
	for ini := 0; ini < len(docs); ini += lote {
		fim := min(ini+lote, len(docs))
		textos := make([]string, 0, fim-ini)
		for _, d := range docs[ini:fim] {
			textos = append(textos, d.Texto)
		}
		vs, err := emb.Embed(ctx, textos)
		if err != nil {
			return nil, err
		}
		v.vecs = append(v.vecs, vs...)
	}
	return v, nil
}

func (v *Vetorial) Nome() string { return "vetorial" }

func (v *Vetorial) Buscar(ctx context.Context, consulta string, k int) ([]Resultado, error) {
	q, err := v.emb.Embed(ctx, []string{consulta})
	if err != nil {
		return nil, err
	}
	m := &melhores{k: k}
	for i, vec := range v.vecs {
		if p := embed.Cosseno(q[0], vec); p > v.Minimo {
			m.oferecer(Resultado{Doc: v.docs[i], Pontos: p})
		}
	}
	return m.res, nil
}

// Lexical é um índice invertido com a pontuação BM25: quanto mais vezes uma
// palavra rara da consulta aparece no documento, mais ele sobe.
type Lexical struct {
	docs    []Doc
	tamanho []int                // palavras de cada documento
	post    map[string][]posting // palavra -> onde aparece
	media   float64              // tamanho médio dos documentos
}

type posting struct{ doc, tf int }

const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

func NovoLexical(docs []Doc) *Lexical {
	l := &Lexical{docs: docs, tamanho: make([]int, len(docs)), post: map[string][]posting{}}
	total := 0
	for i, d := range docs {
		freq := map[string]int{}
		toks := texto.Tokens(d.Texto)
		for _, t := range toks {
			freq[t]++
		}
		l.tamanho[i] = len(toks)
		total += len(toks)
		for t, n := range freq {
			l.post[t] = append(l.post[t], posting{doc: i, tf: n})
		}
	}
	if len(docs) > 0 {
		l.media = float64(total) / float64(len(docs))
	}
	return l
}

func (l *Lexical) Nome() string { return "lexical" }

func (l *Lexical) Buscar(ctx context.Context, consulta string, k int) ([]Resultado, error) {
	n := float64(len(l.docs))
	pontos := map[int]float64{}
	vistos := map[string]bool{}
	for _, t := range texto.Tokens(consulta) {
		if vistos[t] {
			continue
		}
		vistos[t] = true
		ps := l.post[t]
		df := float64(len(ps))
		idf := math.Log(1 + (n-df+0.5)/(df+0.5))
		for _, p := range ps {
			tf := float64(p.tf)
			norma := bm25K1 * (1 - bm25B + bm25B*float64(l.tamanho[p.doc])/l.media)
			pontos[p.doc] += idf * tf * (bm25K1 + 1) / (tf + norma)
		}
	}
	m := &melhores{k: k}
	for i, p := range pontos {
		m.oferecer(Resultado{Doc: l.docs[i], Pontos: p})
	}
	return m.res, nil
}
