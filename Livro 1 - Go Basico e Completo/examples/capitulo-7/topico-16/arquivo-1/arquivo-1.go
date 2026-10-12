package main

import (
	"errors"
	"fmt"
	"strings"
)

// Canal é o contrato mínimo que o serviço exige de qualquer meio de envio.
type Canal interface {
	Enviar(destinatario, mensagem string) error
}

// FuncCanal permite usar uma função comum como Canal (mesmo truque do http.HandlerFunc).
type FuncCanal func(destinatario, mensagem string) error

func (f FuncCanal) Enviar(d, m string) error { return f(d, m) }

type Email struct{}

func (Email) Enviar(destinatario, mensagem string) error {
	if !strings.Contains(destinatario, "@") {
		return fmt.Errorf("e-mail inválido: %q", destinatario)
	}
	fmt.Printf("[EMAIL] %s: %s\n", destinatario, mensagem)
	return nil
}

type SMS struct{}

func (SMS) Enviar(destinatario, mensagem string) error {
	if len([]rune(mensagem)) > 40 {
		return errors.New("SMS excede 40 caracteres")
	}
	fmt.Printf("[SMS] %s: %s\n", destinatario, mensagem)
	return nil
}

// Servico envia por todos os canais e agrega os erros, sem parar no primeiro.
type Servico struct {
	canais map[string]Canal
}

func NovoServico() *Servico { return &Servico{canais: make(map[string]Canal)} }

func (s *Servico) Registrar(nome string, c Canal) { s.canais[nome] = c }

func (s *Servico) NotificarTodos(destinatario, mensagem string) error {
	var erros []error
	for _, nome := range []string{"email", "sms", "push"} {
		c, ok := s.canais[nome]
		if !ok {
			continue
		}
		if err := c.Enviar(destinatario, mensagem); err != nil {
			erros = append(erros, fmt.Errorf("canal %s: %w", nome, err))
		}
	}
	return errors.Join(erros...)
}

func main() {
	s := NovoServico()
	s.Registrar("email", Email{})
	s.Registrar("sms", SMS{})
	s.Registrar("push", FuncCanal(func(d, m string) error {
		fmt.Printf("[PUSH] %s: %s\n", d, m)
		return nil
	}))

	fmt.Println(s.NotificarTodos("ana@exemplo.com", "Seu pedido foi confirmado!"))
	fmt.Println("---")
	err := s.NotificarTodos("11999990000", "Seu pedido #1234 foi confirmado e sai para entrega ainda hoje.")
	fmt.Println(err)
}
