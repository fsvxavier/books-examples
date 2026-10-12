// Package config lê a configuração do servidor: padrões, variáveis de ambiente e flags.
package config

import (
	"flag"
	"fmt"
	"io"
	"time"
)

// Config é a configuração do servidor.
type Config struct {
	Endereco      string
	BaseURL       string
	Desligamento  time.Duration
	MostrarVersao bool
}

// Ler combina os padrões, o ambiente (ENCURTADOR_ADDR e ENCURTADOR_BASE) e as
// flags, nessa ordem de precedência crescente. As mensagens de uso vão para saida.
func Ler(args []string, getenv func(string) string, saida io.Writer) (Config, error) {
	cfg := Config{
		Endereco:     "127.0.0.1:8080",
		BaseURL:      "http://localhost:8080",
		Desligamento: 5 * time.Second,
	}
	if v := getenv("ENCURTADOR_ADDR"); v != "" {
		cfg.Endereco = v
	}
	if v := getenv("ENCURTADOR_BASE"); v != "" {
		cfg.BaseURL = v
	}

	fs := flag.NewFlagSet("encurtador", flag.ContinueOnError)
	fs.SetOutput(saida)
	fs.StringVar(&cfg.Endereco, "addr", cfg.Endereco, "endereço de escuta")
	fs.StringVar(&cfg.BaseURL, "base", cfg.BaseURL, "prefixo dos links curtos")
	fs.DurationVar(&cfg.Desligamento, "espera", cfg.Desligamento, "prazo do desligamento")
	fs.BoolVar(&cfg.MostrarVersao, "versao", false, "mostra a versão e sai")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if cfg.Desligamento <= 0 {
		return Config{}, fmt.Errorf("espera deve ser positiva, recebi %v", cfg.Desligamento)
	}
	return cfg, nil
}
