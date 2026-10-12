// Package uteis calcula dias uteis considerando fins de semana e feriados nacionais.
package uteis

import "time"

// Pascoa devolve o domingo de Pascoa do ano (algoritmo de Meeus/Jones/Butcher).
func Pascoa(ano int) time.Time {
	a := ano % 19
	b, c := ano/100, ano%100
	d, e := b/4, b%4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i, k := c/4, c%4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	mes := (h + l - 7*m + 114) / 31
	dia := (h+l-7*m+114)%31 + 1
	return time.Date(ano, time.Month(mes), dia, 0, 0, 0, 0, time.UTC)
}

type Calendario struct {
	feriados map[time.Time]string
}

// NovoCalendario reune os feriados nacionais fixos e a Sexta-feira Santa dos anos pedidos.
// Carnaval e Corpus Christi sao "ponto facultativo": cada instituicao decide; acrescente
// com Adicionar se as suas regras os tratarem como nao uteis.
func NovoCalendario(anos ...int) *Calendario {
	c := &Calendario{feriados: map[time.Time]string{}}
	fixos := []struct {
		mes  time.Month
		dia  int
		nome string
	}{
		{time.January, 1, "Confraternizacao Universal"},
		{time.April, 21, "Tiradentes"},
		{time.May, 1, "Dia do Trabalho"},
		{time.September, 7, "Independencia"},
		{time.October, 12, "Nossa Senhora Aparecida"},
		{time.November, 2, "Finados"},
		{time.November, 15, "Proclamacao da Republica"},
		{time.November, 20, "Consciencia Negra"},
		{time.December, 25, "Natal"},
	}
	for _, ano := range anos {
		for _, f := range fixos {
			if f.mes == time.November && f.dia == 20 && ano < 2024 {
				continue // feriado nacional so a partir de 2024
			}
			c.Adicionar(time.Date(ano, f.mes, f.dia, 0, 0, 0, 0, time.UTC), f.nome)
		}
		c.Adicionar(Pascoa(ano).AddDate(0, 0, -2), "Sexta-feira Santa")
	}
	return c
}

func dia(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (c *Calendario) Adicionar(t time.Time, nome string) { c.feriados[dia(t)] = nome }

func (c *Calendario) EhUtil(t time.Time) bool {
	if w := t.Weekday(); w == time.Saturday || w == time.Sunday {
		return false
	}
	_, feriado := c.feriados[dia(t)]
	return !feriado
}

// ProximoUtil devolve o proprio dia se for util, ou o primeiro dia util seguinte.
func (c *Calendario) ProximoUtil(t time.Time) time.Time {
	t = dia(t)
	for !c.EhUtil(t) {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

// SomarUteis soma n dias uteis (n >= 0) a partir de t, sem contar t.
func (c *Calendario) SomarUteis(t time.Time, n int) time.Time {
	t = dia(t)
	for n > 0 {
		t = t.AddDate(0, 0, 1)
		if c.EhUtil(t) {
			n--
		}
	}
	return t
}
