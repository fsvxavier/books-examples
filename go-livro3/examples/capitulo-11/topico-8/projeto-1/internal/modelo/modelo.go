// Package modelo define a fronteira entre os agentes e o modelo de linguagem.
//
// O capítulo usa um modelo falso e determinístico: sem rede, sem custo e com
// o mesmo resultado a cada execução. Um cliente real implementa a mesma interface.
package modelo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// Pedido é o que um agente envia ao modelo.
type Pedido struct {
	Papel     string   // pesquisador, redator ou revisor
	Instrucao string   // o que fazer
	Contexto  []string // material de apoio (fatos, rascunho, crítica)
}

// Resposta traz o texto e a contagem de tokens (aproximada em palavras).
type Resposta struct {
	Texto         string
	TokensEntrada int
	TokensSaida   int
}

// Modelo é a interface que todo agente recebe.
type Modelo interface {
	Gerar(ctx context.Context, p Pedido) (Resposta, error)
}

// ErrIndisponivel simula uma falha transitória do provedor.
var ErrIndisponivel = errors.New("modelo indisponível")

// Falso é um modelo determinístico guiado por regras.
type Falso struct {
	Latencia   time.Duration       // tempo de cada chamada
	Base       map[string][]string // tema -> fatos, cada um no formato "texto|fonte"
	MinFontes  int                 // fontes que o revisor exige
	FalharEm   int64               // se > 0, a chamada de número FalharEm falha uma vez
	chamadas   atomic.Int64
	TotalToken atomic.Int64
}

// Chamadas devolve quantas vezes o modelo foi chamado.
func (f *Falso) Chamadas() int64 { return f.chamadas.Load() }

// Gerar implementa Modelo.
func (f *Falso) Gerar(ctx context.Context, p Pedido) (Resposta, error) {
	n := f.chamadas.Add(1)
	if f.FalharEm > 0 && n == f.FalharEm {
		return Resposta{}, fmt.Errorf("chamada %d: %w", n, ErrIndisponivel)
	}
	if f.Latencia > 0 {
		select {
		case <-time.After(f.Latencia):
		case <-ctx.Done():
			return Resposta{}, ctx.Err()
		}
	}
	var texto string
	switch p.Papel {
	case "pesquisador":
		texto = f.pesquisar(p.Instrucao)
	case "redator":
		texto = f.redigir(p)
	case "revisor":
		texto = f.revisar(p)
	default:
		return Resposta{}, fmt.Errorf("papel desconhecido: %q", p.Papel)
	}
	r := Resposta{
		Texto:         texto,
		TokensEntrada: contar(p.Instrucao) + contar(strings.Join(p.Contexto, " ")),
		TokensSaida:   contar(texto),
	}
	f.TotalToken.Add(int64(r.TokensEntrada + r.TokensSaida))
	return r, nil
}

func (f *Falso) pesquisar(tema string) string {
	fatos, ok := f.Base[tema]
	if !ok {
		return "SEM RESULTADOS"
	}
	return strings.Join(fatos, "\n")
}

// redigir usa só a primeira fonte no primeiro rascunho; depois de uma crítica
// que menciona fontes, cita todas.
func (f *Falso) redigir(p Pedido) string {
	var fatos []string
	critica := false
	for _, c := range p.Contexto {
		switch {
		case strings.HasPrefix(c, "CRITICA:"):
			critica = strings.Contains(c, "fontes")
		default:
			fatos = append(fatos, c)
		}
	}
	var b strings.Builder
	for i, fato := range fatos {
		texto, fonte, _ := strings.Cut(fato, "|")
		b.WriteString(texto)
		if critica || i == 0 {
			b.WriteString(" [fonte: " + fonte + "]")
		}
		b.WriteString(". ")
	}
	return strings.TrimSpace(b.String())
}

func (f *Falso) revisar(p Pedido) string {
	rascunho := strings.Join(p.Contexto, " ")
	if strings.Count(rascunho, "[fonte:") >= f.MinFontes {
		return "APROVADO"
	}
	return fmt.Sprintf("REVISAR: citar pelo menos %d fontes", f.MinFontes)
}

func contar(s string) int { return len(strings.Fields(s)) }
