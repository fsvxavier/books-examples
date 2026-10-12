package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
)

type Chamado struct {
	Assunto    string
	Gravidade  int // 1 (leve) a 5 (critica)
	TratadoPor string
}

// Tratador decide se trata. Devolve true quando tratou (a cadeia para) e false para repassar.
type Tratador interface {
	Nome() string
	Tratar(ctx context.Context, c *Chamado) (tratou bool, err error)
}

// TratadorFunc adapta uma funcao a Tratador (o truque do http.HandlerFunc).
type TratadorFunc struct {
	nome string
	f    func(ctx context.Context, c *Chamado) (bool, error)
}

func (t TratadorFunc) Nome() string { return t.nome }
func (t TratadorFunc) Tratar(ctx context.Context, c *Chamado) (bool, error) {
	return t.f(ctx, c)
}

func Novo(nome string, f func(ctx context.Context, c *Chamado) (bool, error)) Tratador {
	return TratadorFunc{nome, f}
}

var ErrNaoTratado = errors.New("nenhum tratador assumiu o chamado")

// Cadeia e uma lista ordenada de tratadores: em Go, uma fatia costuma ser mais simples que
// uma lista encadeada de objetos que guardam o "proximo".
type Cadeia []Tratador

func (cd Cadeia) Tratar(ctx context.Context, c *Chamado) error {
	for _, t := range cd {
		if err := ctx.Err(); err != nil {
			return err
		}
		tratou, err := t.Tratar(ctx, c)
		if err != nil {
			return fmt.Errorf("tratador %s: %w", t.Nome(), err)
		}
		if tratou {
			c.TratadoPor = t.Nome()
			return nil
		}
	}
	return fmt.Errorf("%w: %q (gravidade %d)", ErrNaoTratado, c.Assunto, c.Gravidade)
}

// ---------- Tratadores ----------

func Robo() Tratador {
	respostas := []string{"senha", "fatura", "horario"}
	return Novo("robo", func(_ context.Context, c *Chamado) (bool, error) {
		a := strings.ToLower(c.Assunto)
		for _, p := range respostas {
			if strings.Contains(a, p) {
				return true, nil
			}
		}
		return false, nil
	})
}

func PorGravidade(nome string, max int) Tratador {
	return Novo(nome, func(_ context.Context, c *Chamado) (bool, error) {
		return c.Gravidade <= max, nil
	})
}

// Fabrica monta a cadeia a partir de nomes (por exemplo, lidos de configuracao).
func Fabrica(nomes ...string) (Cadeia, error) {
	catalogo := map[string]Tratador{
		"robo":         Robo(),
		"atendente":    PorGravidade("atendente", 2),
		"especialista": PorGravidade("especialista", 4),
		"coordenacao":  PorGravidade("coordenacao", 5),
	}
	var cd Cadeia
	for _, n := range nomes {
		t, ok := catalogo[n]
		if !ok {
			return nil, fmt.Errorf("tratador desconhecido: %q", n)
		}
		cd = append(cd, t)
	}
	return cd, nil
}

// Inserir devolve uma NOVA cadeia com o tratador na posicao pos (a original nao muda).
func (cd Cadeia) Inserir(pos int, t Tratador) Cadeia {
	return slices.Insert(slices.Clone(cd), pos, t)
}
