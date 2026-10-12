package main

import (
	"context"
	"errors"
	"fmt"

	_ "exemplos/livro2/capitulo-2/topico-2/projeto-1/cartao" // o "_" so existe para o init() registrar o cartao
	_ "exemplos/livro2/capitulo-2/topico-2/projeto-1/pix"
	"exemplos/livro2/capitulo-2/topico-2/projeto-1/proc"
)

func main() {
	fmt.Println("registrados:", proc.Nomes())

	casos := []struct {
		nome string
		cfg  map[string]string
	}{
		{"pix", map[string]string{"chave": "loja@exemplo.com"}},
		{"cartao", map[string]string{"parcelas": "3"}},
		{"cartao", map[string]string{"parcelas": "99"}},
		{"pix", nil},
		{"boleto", nil},
	}
	for _, c := range casos {
		p, err := proc.Novo(c.nome, c.cfg)
		if err != nil {
			fmt.Printf("Novo(%q): erro: %v | desconhecido=%v\n", c.nome, err, errors.Is(err, proc.ErrDesconhecido))
			continue
		}
		msg, _ := p.Cobrar(context.Background(), 29990)
		fmt.Printf("Novo(%q): %s\n", c.nome, msg)
	}

	defer func() { fmt.Println("registrar duplicado:", recover()) }()
	proc.Registrar("pix", func(map[string]string) (proc.Processador, error) { return nil, nil })
}
