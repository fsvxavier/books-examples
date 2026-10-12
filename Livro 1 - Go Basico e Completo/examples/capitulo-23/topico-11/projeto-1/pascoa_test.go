package uteis

import (
	"testing"
	"time"
)

func d(a, m, dia int) time.Time { return time.Date(a, time.Month(m), dia, 0, 0, 0, 0, time.UTC) }

func TestPascoa(t *testing.T) {
	casos := map[int]time.Time{ // datas conferidas em calendarios oficiais
		2024: d(2024, 3, 31),
		2025: d(2025, 4, 20),
		2026: d(2026, 4, 5),
		2027: d(2027, 3, 28),
	}
	for ano, esperada := range casos {
		if got := Pascoa(ano); !got.Equal(esperada) {
			t.Errorf("Pascoa(%d) = %s, esperava %s", ano, got.Format(time.DateOnly), esperada.Format(time.DateOnly))
		}
	}
}

func TestUteis(t *testing.T) {
	c := NovoCalendario(2026, 2027)
	casos := []struct {
		nome string
		got  time.Time
		want time.Time
	}{
		{"5 uteis a partir de sexta 2/10/2026", c.SomarUteis(d(2026, 10, 2), 5), d(2026, 10, 9)},
		{"pula 12/10 (segunda, feriado)", c.SomarUteis(d(2026, 10, 9), 1), d(2026, 10, 13)},
		{"pula Sexta-feira Santa 3/4/2026", c.SomarUteis(d(2026, 4, 2), 1), d(2026, 4, 6)},
		{"vencimento em sabado vai para segunda", c.ProximoUtil(d(2026, 10, 3)), d(2026, 10, 5)},
		{"Natal numa sexta, proximo util", c.ProximoUtil(d(2026, 12, 25)), d(2026, 12, 28)},
		{"cruza o ano", c.SomarUteis(d(2026, 12, 30), 2), d(2027, 1, 4)},
	}
	for _, x := range casos {
		if !x.got.Equal(x.want) {
			t.Errorf("%s: got %s, want %s", x.nome, x.got.Format(time.DateOnly), x.want.Format(time.DateOnly))
		}
	}
	if c.EhUtil(d(2026, 11, 20)) {
		t.Error("20/11/2026 deveria ser feriado")
	}
	if !NovoCalendario(2023).EhUtil(d(2023, 11, 20)) {
		t.Error("20/11/2023 ainda nao era feriado nacional")
	}
}
