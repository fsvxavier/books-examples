// Package portao é a esteira de verificação do código gerado: cada etapa
// examina a solução e devolve achados.
package portao

import (
	"context"
	"os"
	"path/filepath"

	"exemplo.com/guarda/internal/achado"
)

// Estado é o que as etapas compartilham: o fonte, que as etapas podem normalizar.
type Estado struct {
	Fonte string
	Dir   string // pasta de trabalho onde a solução é gravada (vazio: não grava)
}

// gravar escreve o fonte atual em solucao.go, na pasta de trabalho.
func (st *Estado) gravar() error {
	if st.Dir == "" {
		return nil
	}
	return os.WriteFile(filepath.Join(st.Dir, "solucao.go"), []byte(st.Fonte), 0o644)
}

// Etapa é um portão da esteira.
type Etapa struct {
	Nome  string
	Rodar func(ctx context.Context, st *Estado) ([]achado.Achado, error)
}
