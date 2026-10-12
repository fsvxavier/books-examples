package main

import (
	"fmt"
	"io"
	"os"

	"exemplos/livro2/capitulo-4/topico-14/projeto-1"
	"exemplos/livro2/capitulo-4/topico-16/projeto-1/cobranca"
	"exemplos/livro2/capitulo-4/topico-16/projeto-1/config"
	"exemplos/livro2/capitulo-4/topico-16/projeto-1/meios/cartao"
	"exemplos/livro2/capitulo-4/topico-16/projeto-1/meios/pix"
)

// montar e a raiz de composicao: o unico lugar que conhece todas as pecas.
func montar(getenv func(string) string) (*cobranca.Servico, error) {
	cfg, err := config.Carregar(getenv)
	if err != nil {
		return nil, err
	}
	cat := registro.Novo[cobranca.Meio]("meios")
	cat.Registrar("cartao", cartao.Novo())
	cat.Registrar("pix", pix.Novo())
	cat.Congelar()
	return cobranca.NovoServico(cfg, cat), nil
}

func run(getenv func(string) string, stdout io.Writer) error {
	svc, err := montar(getenv)
	if err != nil {
		return err
	}
	for _, c := range []struct {
		meio  string
		valor int64
	}{{"pix", 15_000}, {"cartao", 99_90}, {"boleto", 1_000}, {"pix", 900_000}} {
		comp, err := svc.Cobrar(c.meio, c.valor)
		if err != nil {
			fmt.Fprintf(stdout, "%-6s %7d -> erro: %v\n", c.meio, c.valor, err)
			continue
		}
		fmt.Fprintf(stdout, "%-6s %7d -> %s\n", c.meio, c.valor, comp)
	}
	return nil
}

func main() {
	if err := run(os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}
