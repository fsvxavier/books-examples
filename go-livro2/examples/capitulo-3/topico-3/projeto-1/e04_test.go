package e04

import (
	"testing"
	"time"
)

type Cliente struct {
	URL     string
	Timeout time.Duration
	Tentat  int
	Debug   bool
}

// ---- opcoes funcionais ----

type Opcao func(*Cliente)

func ComTimeout(d time.Duration) Opcao { return func(c *Cliente) { c.Timeout = d } }
func ComTentativas(n int) Opcao        { return func(c *Cliente) { c.Tentat = n } }
func ComDebug() Opcao                  { return func(c *Cliente) { c.Debug = true } }

func NovoOpcoes(url string, opts ...Opcao) *Cliente {
	c := &Cliente{URL: url, Timeout: 30 * time.Second, Tentat: 3}
	for _, o := range opts {
		o(c)
	}
	return c
}

// ---- struct de configuracao ----

type Config struct {
	Timeout time.Duration
	Tentat  int
	Debug   bool
}

func NovoConfig(url string, cfg Config) *Cliente {
	c := &Cliente{URL: url, Timeout: 30 * time.Second, Tentat: 3}
	if cfg.Timeout != 0 {
		c.Timeout = cfg.Timeout
	}
	if cfg.Tentat != 0 {
		c.Tentat = cfg.Tentat
	}
	c.Debug = cfg.Debug
	return c
}

// ---- builder fluente ----

type Builder struct{ c Cliente }

func Novo(url string) *Builder {
	return &Builder{Cliente{URL: url, Timeout: 30 * time.Second, Tentat: 3}}
}

func (b *Builder) Timeout(d time.Duration) *Builder { b.c.Timeout = d; return b }
func (b *Builder) Tentativas(n int) *Builder        { b.c.Tentat = n; return b }
func (b *Builder) Debug() *Builder                  { b.c.Debug = true; return b }
func (b *Builder) Construir() *Cliente              { c := b.c; return &c }

var sink *Cliente

func BenchmarkOpcoesFuncionais(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = NovoOpcoes("http://x", ComTimeout(5*time.Second), ComTentativas(5), ComDebug())
	}
}

func BenchmarkStructConfig(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = NovoConfig("http://x", Config{Timeout: 5 * time.Second, Tentat: 5, Debug: true})
	}
}

func BenchmarkBuilderFluente(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = Novo("http://x").Timeout(5 * time.Second).Tentativas(5).Debug().Construir()
	}
}
