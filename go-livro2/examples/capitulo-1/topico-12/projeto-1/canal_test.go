package notif

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type gravador struct {
	chamadas []string
	erros    []error // um erro por chamada; nil = sucesso
}

func (g *gravador) Enviar(_ context.Context, destino, msg string) error {
	g.chamadas = append(g.chamadas, destino+":"+msg)
	if len(g.erros) == 0 {
		return nil
	}
	e := g.erros[0]
	g.erros = g.erros[1:]
	return e
}

func TestPrimeiroCanalResolve(t *testing.T) {
	email, sms := &gravador{}, &gravador{}
	s := NovoServico().Adicionar("email", email).Adicionar("sms", sms)

	canal, err := s.Notificar(t.Context(), "ana", "oi")
	if canal != "email" || err != nil || len(sms.chamadas) != 0 {
		t.Fatalf("canal=%q err=%v sms=%v", canal, err, sms.chamadas)
	}
}

func TestFallbackParaOProximoCanal(t *testing.T) {
	email := &gravador{erros: []error{errors.New("smtp fora do ar")}}
	sms := &gravador{}
	s := NovoServico().Adicionar("email", email).Adicionar("sms", sms)

	canal, err := s.Notificar(t.Context(), "ana", "oi")
	if canal != "sms" {
		t.Fatalf("canal=%q", canal)
	}
	// o erro do canal que falhou e preservado para o log, mas o envio deu certo
	if err == nil || !strings.Contains(err.Error(), "email: smtp fora do ar") {
		t.Fatalf("esperava registrar a falha do e-mail, veio %v", err)
	}
}

func TestTodosFalham(t *testing.T) {
	boom := errors.New("boom")
	s := NovoServico().
		Adicionar("email", &gravador{erros: []error{boom}}).
		Adicionar("sms", &gravador{erros: []error{boom}})

	_, err := s.Notificar(t.Context(), "ana", "oi")
	if !errors.Is(err, ErrSemCanal) || !errors.Is(err, boom) {
		t.Fatalf("esperava ErrSemCanal e boom na cadeia, veio %v", err)
	}
}

func TestContextoCancelado(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	s := NovoServico().Adicionar("email", &gravador{})
	if _, err := s.Notificar(ctx, "ana", "oi"); !errors.Is(err, context.Canceled) {
		t.Fatalf("esperava context.Canceled, veio %v", err)
	}
}

func TestDecoratorDeTentativas(t *testing.T) {
	instavel := &gravador{erros: []error{errors.New("1"), errors.New("2"), nil}}
	s := NovoServico().Adicionar("email", ComTentativas(instavel, 3))

	canal, err := s.Notificar(t.Context(), "ana", "oi")
	if canal != "email" || err != nil || len(instavel.chamadas) != 3 {
		t.Fatalf("canal=%q err=%v chamadas=%d", canal, err, len(instavel.chamadas))
	}
}

func TestNovoCanalSemTocarNoServico(t *testing.T) {
	var recebido string
	push := CanalFunc(func(_ context.Context, destino, msg string) error {
		recebido = destino + "|" + msg
		return nil
	})
	s := NovoServico().Adicionar("push", push)
	if _, err := s.Notificar(t.Context(), "ana", "oi"); err != nil || recebido != "ana|oi" {
		t.Fatalf("recebido=%q err=%v", recebido, err)
	}
}
