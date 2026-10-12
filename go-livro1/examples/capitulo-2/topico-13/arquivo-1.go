package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	idadeMinima = 18
	nomeMinimo  = 3
	nomeMaximo  = 60
	emailMaximo = 254
)

// Cadastro reúne os dados recebidos de um formulário.
type Cadastro struct {
	Nome  string
	Idade int
	Email string
}

// Validar devolve a lista de problemas encontrados. Lista vazia significa cadastro válido.
func (c Cadastro) Validar() []string {
	var problemas []string

	nome := strings.TrimSpace(c.Nome)
	tamanho := utf8.RuneCountInString(nome)
	switch {
	case tamanho < nomeMinimo:
		problemas = append(problemas, fmt.Sprintf("nome precisa de pelo menos %d caracteres", nomeMinimo))
	case tamanho > nomeMaximo:
		problemas = append(problemas, fmt.Sprintf("nome passa de %d caracteres", nomeMaximo))
	}

	if c.Idade < idadeMinima {
		problemas = append(problemas, fmt.Sprintf("idade mínima é %d anos", idadeMinima))
	}

	email := strings.TrimSpace(c.Email)
	usuario, dominio, ok := strings.Cut(email, "@")
	if !ok || usuario == "" || !strings.Contains(dominio, ".") || len(email) > emailMaximo {
		problemas = append(problemas, "e-mail inválido")
	}

	return problemas
}

func main() {
	casos := []Cadastro{
		{"Ana Souza", 25, "ana@exemplo.com"},
		{"Jo", 16, "email-sem-arroba"},
		{"  Zé  ", 30, "ze@exemplo"},
		{"José", 20, "jose@exemplo.com.br"},
	}

	for _, c := range casos {
		problemas := c.Validar()
		if len(problemas) == 0 {
			fmt.Printf("%q: válido\n", c.Nome)
			continue
		}
		fmt.Printf("%q: %d problema(s)\n", c.Nome, len(problemas))
		for _, p := range problemas {
			fmt.Println("  -", p)
		}
	}
}
