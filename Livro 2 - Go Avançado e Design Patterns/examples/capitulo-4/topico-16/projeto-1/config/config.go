// Package config carrega a configuracao do servico a partir do ambiente.
package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Config e imutavel depois de carregada: so tem campos de leitura copiados.
type Config struct {
	Ambiente       string
	LimiteCentavos int64
	Prazo          time.Duration
}

// Carregar e a forma preferida: recebe a fonte das variaveis, nao toca em
// nenhum estado global e por isso e trivial de testar.
func Carregar(getenv func(string) string) (Config, error) {
	c := Config{Ambiente: "dev", LimiteCentavos: 500_000, Prazo: 2 * time.Second}
	if v := getenv("COBRANCA_AMBIENTE"); v != "" {
		c.Ambiente = v
	}
	if v := getenv("COBRANCA_LIMITE"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("COBRANCA_LIMITE invalido: %q", v)
		}
		c.LimiteCentavos = n
	}
	if v := getenv("COBRANCA_PRAZO"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("COBRANCA_PRAZO invalido: %q", v)
		}
		c.Prazo = d
	}
	return c, nil
}

// Padrao e o Singleton de conveniencia para codigo que nao recebe a
// configuracao por parametro (por exemplo, uma biblioteca interna antiga).
// Carrega uma vez, na primeira chamada, e guarda o erro.
var Padrao = sync.OnceValues(func() (Config, error) {
	return Carregar(os.Getenv)
})
