package main

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-14/projeto-1"
	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-16/projeto-1/cobranca"
	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-16/projeto-1/config"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestRunSaida(t *testing.T) {
	var out bytes.Buffer
	if err := run(env(nil), &out); err != nil {
		t.Fatal(err)
	}
	for _, quero := range []string{"PIX-15000", "CARTAO-9990", `meio de pagamento desconhecido: "boleto"`, "valor acima do limite"} {
		if !strings.Contains(out.String(), quero) {
			t.Errorf("saida sem %q:\n%s", quero, out.String())
		}
	}
}

func TestLimiteVemDoAmbiente(t *testing.T) {
	var out bytes.Buffer
	if err := run(env(map[string]string{"COBRANCA_LIMITE": "1000000"}), &out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "acima do limite") {
		t.Errorf("com limite de 1.000.000 nada deveria passar do limite:\n%s", out.String())
	}
}

func TestConfigInvalida(t *testing.T) {
	err := run(env(map[string]string{"COBRANCA_LIMITE": "abc"}), &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "COBRANCA_LIMITE invalido") {
		t.Fatalf("err = %v", err)
	}
}

type falso struct {
	mu     sync.Mutex
	chamou []int64
}

func (f *falso) Cobrar(c int64) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.chamou = append(f.chamou, c)
	return "FALSO", nil
}

// Cada teste monta o SEU catalogo: nao ha estado global para limpar.
func TestServicoComMeioFalso(t *testing.T) {
	f := &falso{}
	cat := registro.Novo[cobranca.Meio]("teste")
	cat.Registrar("falso", f)
	cat.Congelar()
	svc := cobranca.NovoServico(config.Config{LimiteCentavos: 100}, cat)

	if _, err := svc.Cobrar("falso", 50); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Cobrar("falso", 500); !errors.Is(err, cobranca.ErrAcimaDoLimite) {
		t.Fatalf("err = %v", err)
	}
	if _, err := svc.Cobrar("x", 1); !errors.Is(err, cobranca.ErrMeioDesconhecido) {
		t.Fatalf("err = %v", err)
	}
	if len(f.chamou) != 1 || f.chamou[0] != 50 {
		t.Fatalf("chamadas ao meio: %v", f.chamou)
	}
}

func TestCobrancasConcorrentes(t *testing.T) {
	svc, err := montar(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			if _, err := svc.Cobrar("pix", int64(i+1)); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
}
