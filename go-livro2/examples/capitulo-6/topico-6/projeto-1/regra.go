package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Regra e o COMPONENTE: tanto uma regra simples quanto uma combinacao de regras.
type Regra[T any] interface {
	Validar(v T) error
}

// RegraFunc adapta uma funcao a interface (folha rapida).
type RegraFunc[T any] func(T) error

func (f RegraFunc[T]) Validar(v T) error { return f(v) }

// ---------- Containers (compostos) ----------

type todas[T any] struct{ regras []Regra[T] }

// Todas exige que TODAS passem e reune todos os erros (nao para na primeira).
func Todas[T any](regras ...Regra[T]) Regra[T] { return todas[T]{regras} }

func (t todas[T]) Validar(v T) error {
	var erros []error
	for _, r := range t.regras {
		if err := r.Validar(v); err != nil {
			erros = append(erros, err)
		}
	}
	return errors.Join(erros...)
}

type alguma[T any] struct{ regras []Regra[T] }

// Alguma exige que PELO MENOS UMA passe.
func Alguma[T any](regras ...Regra[T]) Regra[T] { return alguma[T]{regras} }

func (a alguma[T]) Validar(v T) error {
	var erros []error
	for _, r := range a.regras {
		err := r.Validar(v)
		if err == nil {
			return nil
		}
		erros = append(erros, err)
	}
	return fmt.Errorf("nenhuma alternativa passou: %w", errors.Join(erros...))
}

type nao[T any] struct {
	regra Regra[T]
	msg   string
}

// Nao inverte uma regra. Tem UM filho (e ja se parece com um Decorator).
func Nao[T any](r Regra[T], msg string) Regra[T] { return nao[T]{r, msg} }

func (n nao[T]) Validar(v T) error {
	if n.regra.Validar(v) == nil {
		return errors.New(n.msg)
	}
	return nil
}

// ---------- Folhas (regras simples) ----------

func NaoVazio(campo string) Regra[string] {
	return RegraFunc[string](func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s: obrigatorio", campo)
		}
		return nil
	})
}

func Tamanho(campo string, min, max int) Regra[string] {
	return RegraFunc[string](func(s string) error {
		if n := utf8.RuneCountInString(s); n < min || n > max {
			return fmt.Errorf("%s: tamanho %d fora de [%d, %d]", campo, n, min, max)
		}
		return nil
	})
}

func Contem(campo, trecho string) Regra[string] {
	return RegraFunc[string](func(s string) error {
		if !strings.Contains(s, trecho) {
			return fmt.Errorf("%s: deve conter %q", campo, trecho)
		}
		return nil
	})
}
