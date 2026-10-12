// Package config carrega a configuração do programa a partir de variáveis de ambiente.
package config

import (
	"fmt"
	"time"
)

type Config struct {
	Endereco         string
	TimeoutPagamento time.Duration
}

// Carregar lê a configuração. Recebe a função que consulta o ambiente (os.Getenv em produção),
// o que torna a função trivial de testar, sem mexer em variáveis de ambiente reais.
func Carregar(getenv func(string) string) (Config, error) {
	cfg := Config{Endereco: ":8080", TimeoutPagamento: 2 * time.Second}

	if v := getenv("ENDERECO"); v != "" {
		cfg.Endereco = v
	}
	if v := getenv("TIMEOUT_PAGAMENTO"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("TIMEOUT_PAGAMENTO inválido %q: use uma duração positiva, como 500ms", v)
		}
		cfg.TimeoutPagamento = d
	}
	return cfg, nil
}
