package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Regra valida um valor de texto e devolve um erro descritivo, ou nil se estiver tudo certo.
type Regra func(valor string) error

func Obrigatorio() Regra {
	return func(valor string) error {
		if strings.TrimSpace(valor) == "" {
			return errors.New("é obrigatório")
		}
		return nil
	}
}

func TamanhoMinimo(minimo int) Regra {
	return func(valor string) error {
		if n := utf8.RuneCountInString(strings.TrimSpace(valor)); n < minimo {
			return fmt.Errorf("deve ter ao menos %d caracteres (tem %d)", minimo, n)
		}
		return nil
	}
}

func Contem(trecho string) Regra {
	return func(valor string) error {
		if !strings.Contains(valor, trecho) {
			return fmt.Errorf("deve conter %q", trecho)
		}
		return nil
	}
}

// Campo associa um nome, o valor recebido e as regras que ele precisa cumprir.
type Campo struct {
	Nome   string
	Valor  string
	Regras []Regra
}

// Validar executa todas as regras de todos os campos e devolve os problemas encontrados.
func Validar(campos ...Campo) []error {
	var erros []error
	for _, c := range campos {
		for _, regra := range c.Regras {
			if err := regra(c.Valor); err != nil {
				erros = append(erros, fmt.Errorf("%s: %w", c.Nome, err))
			}
		}
	}
	return erros
}

func main() {
	erros := Validar(
		Campo{"nome", "Jo", []Regra{Obrigatorio(), TamanhoMinimo(3)}},
		Campo{"email", "ana.exemplo.com", []Regra{Obrigatorio(), Contem("@")}},
		Campo{"cidade", "", []Regra{Obrigatorio()}},
		Campo{"apelido", "Zé", []Regra{TamanhoMinimo(2)}},
	)

	if len(erros) == 0 {
		fmt.Println("formulário válido")
		return
	}
	fmt.Printf("%d problema(s):\n", len(erros))
	for _, err := range erros {
		fmt.Println(" -", err)
	}
	fmt.Println(errors.Join(erros...) != nil)
}
