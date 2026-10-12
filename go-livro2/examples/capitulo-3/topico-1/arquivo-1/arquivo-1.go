package main

import (
	"fmt"
	"time"
)

// ---- 1. Construtor com muitos parametros: o "telescopio" ----

type Servidor struct {
	Host         string
	Porta        int
	TLS          bool
	LeituraMax   time.Duration
	EscritaMax   time.Duration
	ConexoesMax  int
	NomeDoLogger string
}

func NovoServidorTelescopico(host string, porta int, tls bool, leitura, escrita time.Duration, max int, logger string) *Servidor {
	return &Servidor{host, porta, tls, leitura, escrita, max, logger}
}

// ---- 2. Struct de configuracao: o valor zero vira ambiguo ----

type Config struct {
	Host        string
	Porta       int
	TLS         bool
	LeituraMax  time.Duration // zero significa "sem limite" ou "nao informei"?
	ConexoesMax int
}

func NovoDeConfig(c Config) *Servidor {
	if c.Porta == 0 {
		c.Porta = 8080
	}
	return &Servidor{Host: c.Host, Porta: c.Porta, TLS: c.TLS, LeituraMax: c.LeituraMax, ConexoesMax: c.ConexoesMax}
}

// ---- 3. Opcoes funcionais ----

type Opcao func(*Servidor) error

func ComPorta(p int) Opcao {
	return func(s *Servidor) error {
		if p < 1 || p > 65535 {
			return fmt.Errorf("porta invalida: %d", p)
		}
		s.Porta = p
		return nil
	}
}

func ComTLS() Opcao { return func(s *Servidor) error { s.TLS = true; return nil } }

func ComTimeouts(leitura, escrita time.Duration) Opcao {
	return func(s *Servidor) error {
		if leitura < 0 || escrita < 0 {
			return fmt.Errorf("timeouts nao podem ser negativos")
		}
		s.LeituraMax, s.EscritaMax = leitura, escrita
		return nil
	}
}

func ComMaxConexoes(n int) Opcao {
	return func(s *Servidor) error {
		if n < 1 {
			return fmt.Errorf("max de conexoes deve ser >= 1")
		}
		s.ConexoesMax = n
		return nil
	}
}

func NovoServidor(host string, opts ...Opcao) (*Servidor, error) {
	s := &Servidor{Host: host, Porta: 8080, LeituraMax: 5 * time.Second, EscritaMax: 10 * time.Second, ConexoesMax: 100} // padroes
	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, fmt.Errorf("opcao invalida: %w", err)
		}
	}
	return s, nil
}

func main() {
	// 1: o que significam esses argumentos? E a ordem dos dois time.Duration?
	a := NovoServidorTelescopico("api.exemplo.com", 8443, true, 5*time.Second, 10*time.Second, 100, "")
	fmt.Printf("1 telescopico: %+v\n", *a)

	// 2: LeituraMax zero e "sem limite" ou "usar o padrao"?
	b := NovoDeConfig(Config{Host: "api.exemplo.com", TLS: true})
	fmt.Printf("2 struct:      %+v  <- LeituraMax=0: sem limite? padrao?\n", *b)

	// 3: so o que importa, com padroes sensatos e validacao
	c, err := NovoServidor("api.exemplo.com", ComPorta(8443), ComTLS())
	fmt.Printf("3 opcoes:      %+v err=%v\n", *c, err)
	_, err = NovoServidor("api.exemplo.com", ComPorta(70000), ComMaxConexoes(0))
	fmt.Println("3 invalida:   ", err)
	_, err = NovoServidor("x", ComMaxConexoes(0))
	fmt.Println("3 invalida 2: ", err)
}
