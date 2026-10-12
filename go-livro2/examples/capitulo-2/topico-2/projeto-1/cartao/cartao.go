package cartao

import (
	"context"
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-2/topico-2/projeto-1/proc"
)

type processador struct{ parcelas int }

func (p processador) Cobrar(_ context.Context, c int64) (string, error) {
	return fmt.Sprintf("cartao em %dx de R$ %.2f", p.parcelas, float64(c)/100/float64(p.parcelas)), nil
}

func init() {
	proc.Registrar("cartao", func(cfg map[string]string) (proc.Processador, error) {
		n := 1
		if v := cfg["parcelas"]; v != "" {
			if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n < 1 || n > 12 {
				return nil, fmt.Errorf("cartao: parcelas invalidas: %q", v)
			}
		}
		return processador{parcelas: n}, nil
	})
}
