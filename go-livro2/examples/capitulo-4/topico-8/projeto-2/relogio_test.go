package relogio

import (
	"sync"
	"testing"
	"time"
)

type fixo time.Time

func (f fixo) Agora() time.Time { return time.Time(f) }

// Teste 1: o singleton com Once NAO pode ser trocado: o teste depende do relogio real.
func TestSingletonDeOnceNaoSeTroca(t *testing.T) {
	r := Obter()
	if _, ok := r.(sistema); !ok {
		t.Fatalf("esperava o relogio do sistema, veio %T", r)
	}
}

// Teste 2: com Definir, o teste controla o tempo e restaura ao final.
func TestPadraoTrocavel(t *testing.T) {
	data := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	t.Cleanup(Definir(fixo(data)))

	if got := Padrao().Agora(); !got.Equal(data) {
		t.Fatalf("got %v, quero %v", got, data)
	}
}

// Teste 3: prova que o Cleanup restaurou o relogio real.
func TestRestaurado(t *testing.T) {
	if _, ok := Padrao().(sistema); !ok {
		t.Fatalf("o relogio nao foi restaurado: %T", Padrao())
	}
}

// Teste 4: Definir em testes paralelos e uma armadilha: os dois brigam pelo global.
// As duas barreiras so existem para tornar o resultado reproduzivel: sem elas, a
// falha aparece "de vez em quando", que e pior.
func TestParalelosBrigam(t *testing.T) {
	var trocaram, leram sync.WaitGroup
	trocaram.Add(2)
	leram.Add(2)
	for _, nome := range []string{"a", "b"} {
		t.Run(nome, func(t *testing.T) {
			t.Parallel()
			d := time.Date(2026, 1, 1, int(nome[0]-'a'), 0, 0, 0, time.UTC)
			t.Cleanup(Definir(fixo(d)))
			trocaram.Done()
			trocaram.Wait() // os dois ja trocaram o relogio global
			got := Padrao().Agora()
			leram.Done()
			leram.Wait() // os dois ja leram, so entao os Cleanups rodam
			if !got.Equal(d) {
				t.Errorf("subteste %s viu %v, esperava %v", nome, got, d)
			}
		})
	}
}
