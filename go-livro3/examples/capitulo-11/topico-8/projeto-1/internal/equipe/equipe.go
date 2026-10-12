// Package equipe é a equipe de redação: um pesquisador, um redator e um revisor
// ligados a um modelo, mais o grafo que os conduz.
package equipe

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-11/topico-8/projeto-1/internal/grafo"
	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-11/topico-8/projeto-1/internal/modelo"
)

// ErrReprovado indica que quem aprova (um humano) recusou o resultado.
var ErrReprovado = errors.New("reprovado na aprovação")

// Opcoes ajusta a execução do grafo da equipe.
type Opcoes struct {
	Tentativas int           // tentativas por nó diante de erro transitório
	Espera     time.Duration // espera entre tentativas
	MaxTokens  int           // orçamento de tokens (zero: sem teto)
	Aprovacao  <-chan bool   // se não for nulo, um humano aprova pelo canal
}

// BaseDemo é a base de fatos que o modelo falso consulta: tema -> "texto|fonte".
func BaseDemo() map[string][]string {
	return map[string][]string{
		"goroutines": {
			"Goroutines são threads leves gerenciadas pelo runtime de Go|go.dev/doc/effective_go",
			"Uma goroutine começa com cerca de 2 KiB de pilha|go.dev/doc/faq",
			"O escalonador multiplexa goroutines em threads do sistema|go.dev/src/runtime/proc.go",
		},
		"concorrência": {
			"Concorrência é lidar com muitas coisas ao mesmo tempo|go.dev/blog/waza-talk",
			"Canais sincronizam goroutines sem travas explícitas|go.dev/doc/effective_go",
			"O pacote context propaga cancelamento e prazos|pkg.go.dev/context",
		},
	}
}

type Estado struct {
	Tema     string
	Subtemas []string // se vazio, pesquisa só o Tema
	Fatos    []string
	Rascunho string
	Critica  string
	Rodadas  int
	Aprovado bool
	Tokens   int
}

type Equipe struct {
	Modelo      modelo.Modelo
	Paralelismo int // pesquisas simultâneas
	// Pesquisar permite trocar o pesquisador local por um remoto (A2A).
	Pesquisar func(ctx context.Context, tema string) (string, error)
}

func (e *Equipe) Pesquisador(ctx context.Context, s Estado) (Estado, error) {
	temas := s.Subtemas
	if len(temas) == 0 {
		temas = []string{s.Tema}
	}
	buscar := e.Pesquisar
	if buscar == nil {
		buscar = e.pesquisarLocal
	}
	blocos, err := grafo.Mapear(ctx, temas, e.Paralelismo, func(ctx context.Context, t string) (string, error) {
		return buscar(ctx, t)
	})
	if err != nil {
		return s, err
	}
	s.Fatos = nil
	for _, b := range blocos {
		if b == "SEM RESULTADOS" {
			continue
		}
		s.Fatos = append(s.Fatos, strings.Split(b, "\n")...)
	}
	if len(s.Fatos) == 0 {
		return s, fmt.Errorf("nenhum fato sobre %q", s.Tema)
	}
	return s, nil
}

func (e *Equipe) Redator(ctx context.Context, s Estado) (Estado, error) {
	apoio := append([]string{}, s.Fatos...)
	if s.Critica != "" {
		apoio = append(apoio, "CRITICA: "+s.Critica)
	}
	r, err := e.Modelo.Gerar(ctx, modelo.Pedido{Papel: "redator", Instrucao: "escreva sobre " + s.Tema, Contexto: apoio})
	if err != nil {
		return s, err
	}
	s.Rascunho = r.Texto
	s.Rodadas++
	s.Tokens += r.TokensEntrada + r.TokensSaida
	return s, nil
}

func (e *Equipe) Revisor(ctx context.Context, s Estado) (Estado, error) {
	r, err := e.Modelo.Gerar(ctx, modelo.Pedido{Papel: "revisor", Instrucao: "revise", Contexto: []string{s.Rascunho}})
	if err != nil {
		return s, err
	}
	s.Tokens += r.TokensEntrada + r.TokensSaida
	s.Aprovado = r.Texto == "APROVADO"
	s.Critica = ""
	if !s.Aprovado {
		s.Critica = strings.TrimPrefix(r.Texto, "REVISAR: ")
	}
	return s, nil
}

func (e *Equipe) Grafo(o ...Opcoes) *grafo.Grafo[Estado] {
	var op Opcoes
	if len(o) > 0 {
		op = o[0]
	}
	envolver := func(no grafo.No[Estado]) grafo.No[Estado] {
		if op.MaxTokens > 0 {
			no = grafo.ComOrcamento(op.MaxTokens, func(s Estado) int { return s.Tokens }, no)
		}
		if op.Tentativas > 1 {
			no = grafo.Tentar(op.Tentativas, op.Espera, func(err error) bool { return errors.Is(err, modelo.ErrIndisponivel) }, no)
		}
		return no
	}
	fim := grafo.Fim
	if op.Aprovacao != nil {
		fim = "aprovar"
	}
	g := grafo.Novo[Estado]("pesquisar").
		No("pesquisar", envolver(e.Pesquisador)).
		No("redigir", envolver(e.Redator)).
		No("revisar", envolver(e.Revisor)).
		Aresta("pesquisar", "redigir").
		Aresta("redigir", "revisar").
		Condicional("revisar", func(s Estado) string {
			if s.Aprovado {
				return fim
			}
			return "redigir"
		})
	if op.Aprovacao != nil {
		g.No("aprovar", e.aprovar(op.Aprovacao)).Aresta("aprovar", grafo.Fim)
	}
	return g
}

func (e *Equipe) aprovar(canal <-chan bool) grafo.No[Estado] {
	return func(ctx context.Context, s Estado) (Estado, error) {
		select {
		case ok := <-canal:
			if !ok {
				return s, ErrReprovado
			}
			return s, nil
		case <-ctx.Done():
			return s, fmt.Errorf("esperando aprovação: %w", context.Cause(ctx))
		}
	}
}

// pesquisarLocal consulta o modelo no papel de pesquisador, quando a equipe
// não recebe uma função de busca própria (Pesquisar).
func (e *Equipe) pesquisarLocal(ctx context.Context, tema string) (string, error) {
	r, err := e.Modelo.Gerar(ctx, modelo.Pedido{Papel: "pesquisador", Instrucao: tema})
	if err != nil {
		return "", err
	}
	return r.Texto, nil
}
