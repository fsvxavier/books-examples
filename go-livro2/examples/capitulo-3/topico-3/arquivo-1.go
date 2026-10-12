package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type cliente struct {
	url     string
	timeout time.Duration
	cert    string
	tls     bool
	log     []string
}

// ---- Opcao como INTERFACE: pode ser descrita, comparada e estendida ----

type Opcao interface {
	aplicar(*cliente) error
	String() string
}

type opcaoFunc struct {
	nome string
	f    func(*cliente) error
}

func (o opcaoFunc) aplicar(c *cliente) error { return o.f(c) }
func (o opcaoFunc) String() string           { return o.nome }

func ComTimeout(d time.Duration) Opcao {
	return opcaoFunc{fmt.Sprintf("ComTimeout(%v)", d), func(c *cliente) error {
		if d <= 0 {
			return errors.New("timeout deve ser positivo")
		}
		c.timeout = d
		return nil
	}}
}

func ComCertificado(caminho string) Opcao {
	return opcaoFunc{fmt.Sprintf("ComCertificado(%q)", caminho), func(c *cliente) error { c.cert = caminho; return nil }}
}

// ComTLS DEPENDE de ComCertificado: a ordem das opcoes importaria se checassemos aqui.
func ComTLS() Opcao {
	return opcaoFunc{"ComTLS()", func(c *cliente) error { c.tls = true; return nil }}
}

// Novo aplica as opcoes e DEPOIS valida as relacoes entre elas (ordem-independente).
func Novo(url string, opts ...Opcao) (*cliente, error) {
	c := &cliente{url: url, timeout: 30 * time.Second}
	var aplicadas []string
	for _, o := range opts {
		if err := o.aplicar(c); err != nil {
			return nil, fmt.Errorf("%v: %w", o, err)
		}
		aplicadas = append(aplicadas, o.String())
	}
	if c.tls && c.cert == "" {
		return nil, errors.New("ComTLS exige ComCertificado")
	}
	c.log = aplicadas
	return c, nil
}

func main() {
	c, err := Novo("https://x", ComTLS(), ComCertificado("/etc/cert.pem"), ComTimeout(5*time.Second))
	fmt.Println("ordem invertida funciona:", err == nil, "| aplicadas:", strings.Join(c.log, ", "))

	_, err = Novo("https://x", ComTLS())
	fmt.Println("relacao invalida:", err)

	_, err = Novo("https://x", ComTimeout(0))
	fmt.Println("opcao invalida:  ", err)

	// duplicata: a ultima vence (documente isso)
	c, _ = Novo("https://x", ComTimeout(1*time.Second), ComTimeout(9*time.Second))
	fmt.Println("duplicada, vale a ultima:", c.timeout)
}
