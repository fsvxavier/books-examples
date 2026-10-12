package pix

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-2/topico-2/projeto-1/proc"
)

type processador struct{ chave string }

func (p processador) Cobrar(_ context.Context, c int64) (string, error) {
	return fmt.Sprintf("PIX de R$ %.2f para a chave %s", float64(c)/100, p.chave), nil
}

func init() {
	proc.Registrar("pix", func(cfg map[string]string) (proc.Processador, error) {
		if cfg["chave"] == "" {
			return nil, errors.New("pix: configuracao 'chave' obrigatoria")
		}
		return processador{chave: cfg["chave"]}, nil
	})
}
