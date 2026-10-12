package dominio_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-6/topico-4/encurtador/internal/dominio"
)

var instante = time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC)

func novoLink(t *testing.T, validade time.Duration) dominio.Link {
	t.Helper()
	c, err := dominio.NovoCodigo("abc234")
	if err != nil {
		t.Fatal(err)
	}
	d, err := dominio.NovoDestino("https://go.dev/doc/")
	if err != nil {
		t.Fatal(err)
	}
	l, err := dominio.NovoLink(c, d, instante, validade)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestNovoLinkNasceComEventoPendente(t *testing.T) {
	l := novoLink(t, 0)
	eventos := l.RecolherEventos()
	if len(eventos) != 1 {
		t.Fatalf("%d eventos, esperado 1", len(eventos))
	}
	criado, ok := eventos[0].(dominio.LinkCriado)
	if !ok || criado.Codigo != l.Codigo() || criado.Nome() != "link_criado" || !criado.Quando().Equal(instante) {
		t.Errorf("evento = %#v", eventos[0])
	}
	if again := l.RecolherEventos(); len(again) != 0 {
		t.Errorf("os eventos deveriam ser entregues uma vez só, vieram %d de novo", len(again))
	}
}

func TestNovoLinkRecusaValoresZeroEValidadeNegativa(t *testing.T) {
	c, _ := dominio.NovoCodigo("abc234")
	d, _ := dominio.NovoDestino("https://go.dev")

	if _, err := dominio.NovoLink(dominio.Codigo{}, d, instante, 0); !errors.Is(err, dominio.ErrCodigoInvalido) {
		t.Errorf("código zero: %v", err)
	}
	if _, err := dominio.NovoLink(c, dominio.Destino{}, instante, 0); !errors.Is(err, dominio.ErrDestinoInvalido) {
		t.Errorf("destino zero: %v", err)
	}
	if _, err := dominio.NovoLink(c, d, instante, -time.Hour); !errors.Is(err, dominio.ErrValidadeInvalida) {
		t.Errorf("validade negativa: %v", err)
	}
}

func TestSemPrazoNuncaExpira(t *testing.T) {
	l := novoLink(t, 0)
	if _, tem := l.Expira(); tem {
		t.Error("um link sem validade não deveria ter prazo")
	}
	if l.Expirado(instante.AddDate(100, 0, 0)) {
		t.Error("um link sem validade não expira, nem em cem anos")
	}
}

func TestPrazoEIntervaloSemiaberto(t *testing.T) {
	l := novoLink(t, time.Hour)
	fim, tem := l.Expira()
	if !tem || !fim.Equal(instante.Add(time.Hour)) {
		t.Fatalf("expira = %v, %v", fim, tem)
	}
	// vale em [criado, fim): no instante exato do fim, já expirou.
	casos := []struct {
		quando   time.Time
		expirado bool
	}{
		{instante, false},
		{fim.Add(-time.Nanosecond), false},
		{fim, true},
		{fim.Add(time.Nanosecond), true},
	}
	for _, c := range casos {
		if got := l.Expirado(c.quando); got != c.expirado {
			t.Errorf("Expirado(%v) = %v, esperado %v", c.quando, got, c.expirado)
		}
	}
}

func TestAcessarRegistraEventoOuRecusa(t *testing.T) {
	l := novoLink(t, time.Hour)
	l.RecolherEventos() // descarta o LinkCriado

	if err := l.Acessar(instante.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	eventos := l.RecolherEventos()
	if len(eventos) != 1 || eventos[0].Nome() != "link_acessado" {
		t.Fatalf("eventos = %#v", eventos)
	}

	if err := l.Acessar(instante.Add(2 * time.Hour)); !errors.Is(err, dominio.ErrExpirado) {
		t.Fatalf("erro = %v, esperado ErrExpirado", err)
	}
	if eventos := l.RecolherEventos(); len(eventos) != 0 {
		t.Errorf("um acesso recusado não pode gerar evento: %#v", eventos)
	}
}

func TestReconstituirNaoGeraEventosEReaplicaAsRegras(t *testing.T) {
	l, err := dominio.Reconstituir("abc234", "https://go.dev", instante, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if eventos := l.RecolherEventos(); len(eventos) != 0 {
		t.Errorf("reconstituir não é criar: %#v", eventos)
	}
	if _, err := dominio.Reconstituir("curto", "https://go.dev", instante, time.Time{}); !errors.Is(err, dominio.ErrCodigoInvalido) {
		t.Errorf("dados corrompidos no armazenamento deveriam ser recusados: %v", err)
	}
	if _, err := dominio.Reconstituir("abc234", "javascript:x", instante, time.Time{}); !errors.Is(err, dominio.ErrDestinoInvalido) {
		t.Errorf("destino corrompido deveria ser recusado: %v", err)
	}
}

// TestEncapsulamento confere, por reflexão, que nenhum tipo do modelo expõe
// campos: é o que impede construir um Link ou um valor inválido pela porta dos
// fundos, com um literal de struct. Se alguém exportar um campo, o teste avisa.
func TestEncapsulamento(t *testing.T) {
	for _, v := range []any{dominio.Link{}, dominio.Codigo{}, dominio.Destino{}} {
		tipo := reflect.TypeOf(v)
		for i := range tipo.NumField() {
			if campo := tipo.Field(i); campo.IsExported() {
				t.Errorf("%s expõe o campo %s", tipo.Name(), campo.Name)
			}
		}
	}
}
