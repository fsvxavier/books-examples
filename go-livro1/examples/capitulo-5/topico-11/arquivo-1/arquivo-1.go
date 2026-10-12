package main

import (
	"errors"
	"fmt"
	"strings"
)

// Centavos guarda dinheiro como inteiro, como vimos no Capítulo 2.
type Centavos int64

func (c Centavos) String() string {
	sinal := ""
	if c < 0 {
		sinal, c = "-", -c
	}
	return fmt.Sprintf("%sR$ %d,%02d", sinal, c/100, c%100)
}

var (
	ErrValorInvalido     = errors.New("valor deve ser positivo")
	ErrSaldoInsuficiente = errors.New("saldo insuficiente")
)

// Transacao é um registro imutável do extrato.
type Transacao struct {
	Tipo  string
	Valor Centavos
	Saldo Centavos
}

// Conta protege o saldo: os campos são privados, e só os métodos o alteram.
type Conta struct {
	titular   string
	saldo     Centavos
	historico []Transacao
}

// NovaConta é o construtor: garante que toda Conta nasce válida.
func NovaConta(titular string, saldoInicial Centavos) (*Conta, error) {
	if strings.TrimSpace(titular) == "" {
		return nil, errors.New("titular é obrigatório")
	}
	if saldoInicial < 0 {
		return nil, errors.New("saldo inicial não pode ser negativo")
	}
	return &Conta{titular: titular, saldo: saldoInicial}, nil
}

func (c *Conta) Titular() string { return c.titular }
func (c *Conta) Saldo() Centavos { return c.saldo }

func (c *Conta) registrar(tipo string, valor Centavos) {
	c.historico = append(c.historico, Transacao{tipo, valor, c.saldo})
}

func (c *Conta) Depositar(valor Centavos) error {
	if valor <= 0 {
		return ErrValorInvalido
	}
	c.saldo += valor
	c.registrar("depósito", valor)
	return nil
}

func (c *Conta) Sacar(valor Centavos) error {
	if valor <= 0 {
		return ErrValorInvalido
	}
	if valor > c.saldo {
		return fmt.Errorf("sacar %v com saldo %v: %w", valor, c.saldo, ErrSaldoInsuficiente)
	}
	c.saldo -= valor
	c.registrar("saque", valor)
	return nil
}

// Transferir move dinheiro entre contas; se o saque falhar, nada é alterado.
func (c *Conta) Transferir(destino *Conta, valor Centavos) error {
	if err := c.Sacar(valor); err != nil {
		return fmt.Errorf("transferência: %w", err)
	}
	return destino.Depositar(valor)
}

// Extrato devolve uma cópia do histórico, para que ninguém altere o original.
func (c *Conta) Extrato() []Transacao {
	return append([]Transacao(nil), c.historico...)
}

func main() {
	ana, _ := NovaConta("Ana", 100_00)
	bia, _ := NovaConta("Bia", 0)

	fmt.Println(ana.Depositar(50_00), ana.Saldo())
	fmt.Println(ana.Sacar(500_00))
	err := ana.Sacar(500_00)
	fmt.Println(errors.Is(err, ErrSaldoInsuficiente))
	fmt.Println(ana.Depositar(-5))

	fmt.Println(ana.Transferir(bia, 30_00), ana.Saldo(), bia.Saldo())
	fmt.Println(ana.Transferir(bia, 1000_00))

	for _, t := range ana.Extrato() {
		fmt.Printf("%-9s %10v -> saldo %v\n", t.Tipo, t.Valor, t.Saldo)
	}

	_, err = NovaConta("  ", 0)
	fmt.Println(err)
}
