// Package janela cuida da memória de curto prazo: o que cabe no contexto do modelo.
//
// O modelo só enxerga o que o prompt contém, e o prompt tem um tamanho máximo
// (medido em tokens). Quando a conversa cresce além disso, é preciso escolher o
// que ficar: cortar o mais antigo, ou trocá-lo por um resumo.
package janela

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Papéis usados nas mensagens.
const (
	Sistema    = "system"
	Usuario    = "user"
	Assistente = "assistant"
)

// Mensagem é uma fala da conversa.
type Mensagem struct {
	Papel string `json:"papel"`
	Texto string `json:"texto"`
}

// Tokens estima o tamanho de um texto em tokens. Aproximação por palavras:
// um tokenizador real conta diferente, mas as proporções entre estratégias se mantêm.
func Tokens(texto string) int { return len(strings.Fields(texto)) }

// Total soma os tokens de uma lista de mensagens (mais 2 por mensagem, de moldura).
func Total(msgs []Mensagem) int {
	n := 0
	for _, m := range msgs {
		n += Tokens(m.Texto) + 2
	}
	return n
}

// ErrOrcamento indica que nem as mensagens fixas cabem no orçamento.
var ErrOrcamento = errors.New("as mensagens fixas já estouram o orçamento")

// Truncar mantém todas as mensagens de sistema e, das demais, as MAIS RECENTES
// que couberem em orcamento tokens. A ordem original é preservada.
func Truncar(msgs []Mensagem, orcamento int) ([]Mensagem, error) {
	fixo := 0
	manter := make([]bool, len(msgs))
	for i, m := range msgs {
		if m.Papel == Sistema {
			fixo += Tokens(m.Texto) + 2
			manter[i] = true
		}
	}
	if fixo > orcamento {
		return nil, fmt.Errorf("fixas=%d orçamento=%d: %w", fixo, orcamento, ErrOrcamento)
	}
	livre := orcamento - fixo
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Papel == Sistema {
			continue
		}
		custo := Tokens(msgs[i].Texto) + 2
		if custo > livre {
			break // a partir daqui só há mensagens mais antigas: pare, não pule
		}
		livre -= custo
		manter[i] = true
	}
	var saida []Mensagem
	for i, m := range msgs {
		if manter[i] {
			saida = append(saida, m)
		}
	}
	return saida, nil
}

// Resumidor condensa mensagens antigas em um texto curto.
type Resumidor interface {
	Resumir(ctx context.Context, msgs []Mensagem) (string, error)
}

// Resumir devolve msgs com o trecho mais antigo trocado por UMA mensagem de
// resumo, até caber em orcamento. recentes é o número mínimo de mensagens finais
// que nunca entram no resumo. O resumo anterior (se houver) entra no novo.
func Resumir(ctx context.Context, msgs []Mensagem, orcamento, recentes int, r Resumidor) ([]Mensagem, error) {
	if Total(msgs) <= orcamento {
		return msgs, nil
	}
	var fixas, corpo []Mensagem
	for _, m := range msgs {
		if m.Papel == Sistema && !strings.HasPrefix(m.Texto, PrefixoResumo) {
			fixas = append(fixas, m)
		} else {
			corpo = append(corpo, m)
		}
	}
	if len(corpo) <= recentes {
		return Truncar(msgs, orcamento)
	}
	antigas, novas := corpo[:len(corpo)-recentes], corpo[len(corpo)-recentes:]
	texto, err := r.Resumir(ctx, antigas)
	if err != nil {
		return nil, fmt.Errorf("resumir: %w", err)
	}
	// O resumo ganha no máximo um terço do orçamento: sem esse teto, um resumo
	// que cresce a cada rodada acabaria comendo o espaço das mensagens recentes.
	// Se passar, descarta o começo (o mais antigo do próprio resumo).
	teto := max(orcamento/3-Tokens(PrefixoResumo), 1)
	if palavras := strings.Fields(texto); len(palavras) > teto {
		texto = strings.Join(palavras[len(palavras)-teto:], " ")
	}
	saida := append(append([]Mensagem{}, fixas...), Mensagem{Papel: Sistema, Texto: PrefixoResumo + texto})
	saida = append(saida, novas...)
	// Se mesmo assim não coube, corta o excesso pelo mais antigo.
	return Truncar(saida, orcamento)
}

// PrefixoResumo marca a mensagem que contém o resumo da conversa antiga.
const PrefixoResumo = "Resumo da conversa até aqui: "
