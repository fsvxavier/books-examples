package pedido

import (
	"errors"
	"testing"
)

// Precificacao testada SEM banco, SEM e-mail, SEM nada: so dados.
func TestPrecificador(t *testing.T) {
	pr := Precificador{Precos: map[string]int64{"A": 1000, "B": 2500}}
	casos := []struct {
		nome  string
		itens map[string]int
		total int64
	}{
		{"sem desconto", map[string]int{"A": 2}, 2000},
		{"no limite (10000) sem desconto", map[string]int{"A": 10}, 10000},
		{"acima do limite, 10% off", map[string]int{"A": 4, "B": 4}, 12600},
	}
	for _, c := range casos {
		if got := pr.Total(Pedido{Cliente: "x", Itens: c.itens}); got != c.total {
			t.Errorf("%s: got %d, want %d", c.nome, got, c.total)
		}
	}
}

// fakes minimos, escritos no proprio teste
type repoFake struct {
	salvos []int64
	erro   error
}

func (r *repoFake) Salvar(_ Pedido, total int64) error {
	if r.erro != nil {
		return r.erro
	}
	r.salvos = append(r.salvos, total)
	return nil
}

type avisoFake struct{ chamadas int }

func (a *avisoFake) Avisar(string, int64) error { a.chamadas++; return errors.New("smtp fora do ar") }

func TestServicoFechar(t *testing.T) {
	repo, aviso := &repoFake{}, &avisoFake{}
	s := Servico{Precificador{map[string]int64{"A": 1000}}, repo, aviso}

	total, err := s.Fechar(Pedido{Cliente: "ana", Itens: map[string]int{"A": 3}})
	if err != nil || total != 3000 || len(repo.salvos) != 1 || aviso.chamadas != 1 {
		t.Fatalf("total=%d err=%v salvos=%v avisos=%d", total, err, repo.salvos, aviso.chamadas)
	}
	// o e-mail falhou, mas o pedido foi salvo e o fechamento deu certo
	if _, err := s.Fechar(Pedido{}); err == nil {
		t.Fatal("pedido vazio devia falhar")
	}
	repo.erro = errors.New("banco caiu")
	if _, err := s.Fechar(Pedido{Cliente: "ana", Itens: map[string]int{"A": 1}}); !errors.Is(err, repo.erro) {
		t.Fatalf("esperava erro do banco, veio %v", err)
	}
}
