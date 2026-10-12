package main

import (
	"errors"
	"fmt"
)

type Comando interface {
	Executar() error
	Desfazer() error
}

// ---------- Comando por funcoes: o mais leve em Go ----------

// Par junta a acao e a sua inversa. Qualquer closure serve, sem declarar um tipo novo.
type Par struct {
	Faz, Desfaz func() error
	Nome        string
}

func (p Par) Executar() error { return p.Faz() }
func (p Par) Desfazer() error {
	if p.Desfaz == nil {
		return fmt.Errorf("%s: nao pode ser desfeito", p.Nome)
	}
	return p.Desfaz()
}

// ---------- Macro: um comando feito de comandos (Composite de Comandos) ----------

type Macro struct {
	Nome     string
	Comandos []Comando
	feitos   int // quantos executaram com sucesso
}

// Executar roda os comandos em ordem; se um falhar, DESFAZ os anteriores em ordem inversa (tudo ou nada).
func (m *Macro) Executar() error {
	m.feitos = 0
	for i, c := range m.Comandos {
		if err := c.Executar(); err != nil {
			errDesfazer := m.desfazerFeitos()
			return errors.Join(fmt.Errorf("macro %q falhou no passo %d: %w", m.Nome, i+1, err), errDesfazer)
		}
		m.feitos++
	}
	return nil
}

func (m *Macro) desfazerFeitos() error {
	var erros []error
	for i := m.feitos - 1; i >= 0; i-- {
		if err := m.Comandos[i].Desfazer(); err != nil {
			erros = append(erros, fmt.Errorf("desfazendo passo %d: %w", i+1, err))
		}
	}
	m.feitos = 0
	return errors.Join(erros...)
}

func (m *Macro) Desfazer() error { return m.desfazerFeitos() }

// ---------- Remetentes: Botao, Menu e Atalho disparam o MESMO comando, sem conhecer o receptor ----------

type Remetente struct {
	Nome    string
	Comando Comando // pode ser trocado em tempo de execucao
}

func (r *Remetente) Acionar() error {
	if r.Comando == nil {
		return fmt.Errorf("%s: nenhum comando associado", r.Nome)
	}
	return r.Comando.Executar()
}
