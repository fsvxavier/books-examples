package main

import (
	"flag"
	"fmt"
	"io"
	"time"
)

// config reúne tudo que muda entre ambientes.
type config struct {
	endereco     string
	baseURL      string
	desligamento time.Duration
}

// lerConfig combina variáveis de ambiente e flags; a flag vence o ambiente.
// Receber getenv como parâmetro evita ler o ambiente real nos testes.
func lerConfig(args []string, getenv func(string) string, saida io.Writer) (config, error) {
	cfg := config{
		endereco:     "127.0.0.1:8080",
		baseURL:      "http://localhost:8080",
		desligamento: 5 * time.Second,
	}
	if v := getenv("ENCURTADOR_ADDR"); v != "" {
		cfg.endereco = v
	}
	if v := getenv("ENCURTADOR_BASE"); v != "" {
		cfg.baseURL = v
	}

	fs := flag.NewFlagSet("encurtador", flag.ContinueOnError)
	fs.SetOutput(saida)
	fs.StringVar(&cfg.endereco, "addr", cfg.endereco, "endereço de escuta")
	fs.StringVar(&cfg.baseURL, "base", cfg.baseURL, "prefixo dos links curtos")
	fs.DurationVar(&cfg.desligamento, "espera", cfg.desligamento, "prazo do desligamento")
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	if cfg.desligamento <= 0 {
		return config{}, fmt.Errorf("espera deve ser positiva, recebi %v", cfg.desligamento)
	}
	return cfg, nil
}
