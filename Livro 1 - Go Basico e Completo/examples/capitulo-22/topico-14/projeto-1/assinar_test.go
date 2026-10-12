package webhook

import (
	"errors"
	"testing"
	"time"
)

func TestVerificar(t *testing.T) {
	segredo := []byte("segredo-do-webhook-com-32-bytes!!")
	t0 := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	corpo := []byte(`{"evento":"pagamento.aprovado","id":"p_123"}`)

	casos := []struct {
		nome     string
		cab      func() string
		corpo    []byte
		id       string
		agora    time.Time
		esperado error
	}{
		{"valido", func() string { return Assinar(segredo, corpo, t0) }, corpo, "e1", t0.Add(2 * time.Second), nil},
		{"corpo adulterado", func() string { return Assinar(segredo, corpo, t0) }, []byte(`{"evento":"pagamento.aprovado","id":"p_999"}`), "e2", t0, ErrAssinatura},
		{"segredo errado", func() string { return Assinar([]byte("outro-segredo-com-32-bytes-......"), corpo, t0) }, corpo, "e3", t0, ErrAssinatura},
		{"timestamp antigo (replay tardio)", func() string { return Assinar(segredo, corpo, t0) }, corpo, "e4", t0.Add(10 * time.Minute), ErrAntigo},
		{"cabecalho malformado", func() string { return "lixo" }, corpo, "e5", t0, ErrFormato},
		{"sem assinatura", func() string { return "t=123" }, corpo, "e6", t0, ErrFormato},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			v := NovoVerificador(segredo, 5*time.Minute)
			v.Agora = func() time.Time { return c.agora }
			if err := v.Verificar(c.cab(), c.corpo, c.id); !errors.Is(err, c.esperado) {
				t.Fatalf("esperava %v, veio %v", c.esperado, err)
			}
		})
	}
}

func TestRepeticaoImediata(t *testing.T) {
	segredo := []byte("segredo-do-webhook-com-32-bytes!!")
	t0 := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	corpo := []byte(`{"id":"p_123"}`)
	v := NovoVerificador(segredo, 5*time.Minute)
	v.Agora = func() time.Time { return t0 }
	cab := Assinar(segredo, corpo, t0)

	if err := v.Verificar(cab, corpo, "evt_1"); err != nil {
		t.Fatalf("primeira entrega: %v", err)
	}
	if err := v.Verificar(cab, corpo, "evt_1"); !errors.Is(err, ErrRepetido) {
		t.Fatalf("segunda entrega devia ser repeticao, veio %v", err)
	}
	// passado o prazo, o registro do evento expira (e o timestamp tambem)
	v.Agora = func() time.Time { return t0.Add(6 * time.Minute) }
	if err := v.Verificar(cab, corpo, "evt_1"); !errors.Is(err, ErrAntigo) {
		t.Fatalf("depois do prazo devia ser ErrAntigo, veio %v", err)
	}
}
