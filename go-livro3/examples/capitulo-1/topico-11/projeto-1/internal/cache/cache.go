// Package cache guarda respostas já geradas, para não pagar duas vezes pelo
// mesmo trabalho. Há dois níveis: o exato (mesma pergunta, no Valkey) e o
// semântico (pergunta parecida, em memória).
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-1/topico-11/projeto-1/internal/embed"
	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-1/topico-11/projeto-1/internal/texto"
)

// Entrada é uma resposta guardada.
type Entrada struct {
	Resposta      string   `json:"resposta"`
	Fontes        []string `json:"fontes"`
	TokensEntrada int      `json:"tokens_entrada"`
	TokensSaida   int      `json:"tokens_saida"`
	Pergunta      string   `json:"pergunta,omitempty"`
}

// Chave identifica a pergunta e a VERSÃO do corpus. A versão entra na chave
// para que reindexar os documentos invalide as respostas antigas: uma
// resposta em cache baseada em documentos que mudaram é uma resposta velha.
func Chave(pergunta, versaoCorpus string) string {
	h := sha256.Sum256([]byte(versaoCorpus + "|" + texto.Canonico(pergunta)))
	return "rag:resp:" + hex.EncodeToString(h[:16])
}

// Valkey é o cache exato, compartilhado entre todas as instâncias da API.
type Valkey struct {
	Cliente valkey.Client
	TTL     time.Duration
	// Prazo limita cada chamada ao Valkey. Um cache lento é pior do que
	// nenhum cache: sem prazo, a pergunta fica presa esperando uma otimização.
	Prazo time.Duration
}

func (v *Valkey) comPrazo(ctx context.Context) (context.Context, context.CancelFunc) {
	if v.Prazo <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, v.Prazo)
}

func (v *Valkey) Get(ctx context.Context, chave string) (Entrada, bool, error) {
	ctx, cancel := v.comPrazo(ctx)
	defer cancel()
	s, err := v.Cliente.Do(ctx, v.Cliente.B().Get().Key(chave).Build()).ToString()
	if valkey.IsValkeyNil(err) {
		return Entrada{}, false, nil
	}
	if err != nil {
		return Entrada{}, false, fmt.Errorf("cache get: %w", err)
	}
	var e Entrada
	if err := json.Unmarshal([]byte(s), &e); err != nil {
		return Entrada{}, false, fmt.Errorf("entrada de cache corrompida: %w", err)
	}
	return e, true, nil
}

func (v *Valkey) Set(ctx context.Context, chave string, e Entrada) error {
	ctx, cancel := v.comPrazo(ctx)
	defer cancel()
	dados, err := json.Marshal(e)
	if err != nil {
		return err
	}
	cmd := v.Cliente.B().Set().Key(chave).Value(string(dados)).ExSeconds(int64(v.TTL.Seconds())).Build()
	if err := v.Cliente.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	return nil
}

// Semantico devolve a resposta de uma pergunta ANTERIOR parecida com a nova,
// quando o cosseno entre elas passa do Limiar. Economiza tokens em perguntas
// reescritas ("como renovar?" e "como faço a renovação?"), mas é arriscado:
// perguntas parecidas nem sempre têm a mesma resposta.
type Semantico struct {
	Emb    embed.Embedder
	Limiar float64
	Max    int // entradas guardadas; as mais antigas saem primeiro

	mu       sync.RWMutex
	vetores  [][]float32
	entradas []Entrada
}

// Acerto descreve uma resposta devolvida pelo cache semântico.
type Acerto struct {
	Entrada      Entrada
	Similaridade float64
}

func (s *Semantico) Buscar(ctx context.Context, pergunta string) (Acerto, bool, error) {
	q, err := s.Emb.Embed(ctx, []string{pergunta})
	if err != nil {
		return Acerto{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	melhor, pos := -1.0, -1
	for i, v := range s.vetores {
		if c := embed.Cosseno(q[0], v); c > melhor {
			melhor, pos = c, i
		}
	}
	if pos < 0 || melhor < s.Limiar {
		return Acerto{Similaridade: melhor}, false, nil
	}
	return Acerto{Entrada: s.entradas[pos], Similaridade: melhor}, true, nil
}

func (s *Semantico) Guardar(ctx context.Context, pergunta string, e Entrada) error {
	q, err := s.Emb.Embed(ctx, []string{pergunta})
	if err != nil {
		return err
	}
	e.Pergunta = pergunta
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Max > 0 && len(s.entradas) >= s.Max {
		s.vetores = s.vetores[1:]
		s.entradas = s.entradas[1:]
	}
	s.vetores = append(s.vetores, q[0])
	s.entradas = append(s.entradas, e)
	return nil
}

// Tamanho devolve quantas entradas há.
func (s *Semantico) Tamanho() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entradas)
}
