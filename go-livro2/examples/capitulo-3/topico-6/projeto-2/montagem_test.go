package montagem

import (
	"errors"
	"strings"
	"testing"
)

type Email struct {
	Para    []string
	Assunto string
	Corpo   string
	Anexos  []string
	Urgente bool
}

func Para(e string) Passo[Email] {
	return func(m *Email) error {
		if !strings.Contains(e, "@") {
			return errors.New("endereco invalido: " + e)
		}
		m.Para = append(m.Para, e)
		return nil
	}
}

func Assunto(s string) Passo[Email] { return func(m *Email) error { m.Assunto = s; return nil } }
func Corpo(s string) Passo[Email]   { return func(m *Email) error { m.Corpo = s; return nil } }
func Anexo(n string) Passo[Email] {
	return func(m *Email) error { m.Anexos = append(m.Anexos, n); return nil }
}
func Urgente() Passo[Email] { return func(m *Email) error { m.Urgente = true; return nil } }

func ExigeDestino() Passo[Email] {
	return Validar(func(m Email) error {
		if len(m.Para) == 0 {
			return errors.New("sem destinatario")
		}
		return nil
	})
}

func TestMontarEmail(t *testing.T) {
	temRelatorio := true
	e, err := Montar(
		Para("ana@exemplo.com"), Para("bia@exemplo.com"),
		Assunto("Fechamento"),
		Corpo("segue o resumo"),
		Quando(temRelatorio, Anexo("relatorio.pdf")),
		Quando(!temRelatorio, Anexo("nunca.pdf")),
		Quando(len("x") > 0, Urgente()),
		ExigeDestino(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Para) != 2 || len(e.Anexos) != 1 || e.Anexos[0] != "relatorio.pdf" || !e.Urgente {
		t.Fatalf("%+v", e)
	}
}

func TestErroIndicaOPasso(t *testing.T) {
	_, err := Montar(Assunto("x"), Para("sem-arroba"), Corpo("y"))
	if err == nil || !strings.HasPrefix(err.Error(), "passo 2:") {
		t.Fatalf("esperava erro no passo 2, veio %v", err)
	}
}

func TestValidacaoFinal(t *testing.T) {
	if _, err := Montar(Assunto("x"), ExigeDestino()); err == nil || !strings.Contains(err.Error(), "sem destinatario") {
		t.Fatalf("esperava recusa, veio %v", err)
	}
}
