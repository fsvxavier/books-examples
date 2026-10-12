package main

import (
	"errors"
	"fmt"
	"slices"
)

type Conta struct {
	saldo   int // centavos
	extrato []string
}

type fotoConta struct {
	saldo   int
	extrato []string
}

func (c *Conta) salvar() fotoConta {
	return fotoConta{c.saldo, slices.Clone(c.extrato)}
}

func (c *Conta) restaurar(f fotoConta) {
	c.saldo = f.saldo
	c.extrato = slices.Clone(f.extrato)
}

var ErrSaldo = errors.New("saldo insuficiente")

func (c *Conta) Depositar(v int) error {
	c.saldo += v
	c.extrato = append(c.extrato, fmt.Sprintf("+%d", v))
	return nil
}

func (c *Conta) Sacar(v int) error {
	c.saldo -= v // aplica primeiro, valida depois: deixa a conta suja
	c.extrato = append(c.extrato, fmt.Sprintf("-%d", v))
	if c.saldo < 0 {
		return ErrSaldo
	}
	return nil
}

// Transacao aplica as operações em ordem; se uma falhar (erro ou panic),
// a conta volta ao instantâneo tirado antes da primeira.
func (c *Conta) Transacao(ops ...func(*Conta) error) (err error) {
	foto := c.salvar()
	defer func() {
		if r := recover(); r != nil {
			c.restaurar(foto)
			panic(r) // restaura e deixa o panic seguir
		}
		if err != nil {
			c.restaurar(foto)
		}
	}()
	for i, op := range ops {
		if err = op(c); err != nil {
			return fmt.Errorf("operação %d: %w", i+1, err)
		}
	}
	return nil
}

func main() {
	c := &Conta{saldo: 10000}

	err := c.Transacao(
		func(c *Conta) error { return c.Depositar(2500) },
		func(c *Conta) error { return c.Sacar(5000) },
	)
	fmt.Println("tx1:", err, c.saldo, c.extrato)

	err = c.Transacao(
		func(c *Conta) error { return c.Depositar(100) },
		func(c *Conta) error { return c.Sacar(99999) },
	)
	fmt.Println("tx2:", err, errors.Is(err, ErrSaldo), c.saldo, c.extrato)

	func() {
		defer func() { fmt.Println("tx3 panic recuperado:", recover(), c.saldo, c.extrato) }()
		_ = c.Transacao(
			func(c *Conta) error { return c.Depositar(1) },
			func(c *Conta) error { var m map[string]int; m["x"] = 1; return nil },
		)
	}()

	// O que acontece sem o instantâneo:
	sem := &Conta{saldo: 10000}
	_ = sem.Depositar(100)
	if err := sem.Sacar(99999); err != nil {
		fmt.Println("sem memento:", err, sem.saldo, sem.extrato)
	}
}
