package main

import (
	"fmt"
	"time"
)

func carregar(nome string) *time.Location {
	loc, err := time.LoadLocation(nome)
	if err != nil {
		panic(err)
	}
	return loc
}

func main() {
	ny := carregar("America/New_York")
	sp := carregar("America/Sao_Paulo")

	// Nova York: em 8 de marco de 2026 os relogios pulam de 02:00 para 03:00
	antes := time.Date(2026, 3, 7, 12, 0, 0, 0, ny)
	fmt.Println("antes do horario de verao:", antes)
	fmt.Println("Add(24h):       ", antes.Add(24*time.Hour), "<- 24h exatas (13h no relogio)")
	fmt.Println("AddDate(0,0,1): ", antes.AddDate(0, 0, 1), "<- mesmo horario no relogio")
	fmt.Println("dia de 8/3 tem: ", time.Date(2026, 3, 9, 0, 0, 0, 0, ny).Sub(time.Date(2026, 3, 8, 0, 0, 0, 0, ny)), "de duracao")

	// hora que nao existe: 02:30 de 8/3 em NY
	inexistente := time.Date(2026, 3, 8, 2, 30, 0, 0, ny)
	fmt.Println("02:30 (nao existe): ", inexistente)
	// hora que existe duas vezes: 01:30 de 1/11 em NY
	ambigua := time.Date(2026, 11, 1, 1, 30, 0, 0, ny)
	fmt.Println("01:30 (ocorre 2x):  ", ambigua, "<- Go escolhe uma")

	// Sao Paulo teve horario de verao ate 2019 e nao tem mais
	for _, ano := range []int{2018, 2019, 2020} {
		jan := time.Date(ano, 1, 15, 12, 0, 0, 0, sp)
		dez := time.Date(ano, 12, 15, 12, 0, 0, 0, sp)
		fmt.Printf("Sao Paulo %d: 15/jan %s | 15/dez %s\n", ano, jan.Format("-07"), dez.Format("-07"))
	}

	// AddDate normaliza o estouro de mes
	jan31 := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	fmt.Println("\n31/jan + 1 mes:   ", jan31.AddDate(0, 1, 0).Format(time.DateOnly), "<- 31/fev virou 3/mar")
	fmt.Println("ultimo dia de fev:", time.Date(2026, 3, 0, 0, 0, 0, 0, time.UTC).Format(time.DateOnly), "(dia 0 de marco)")
	fmt.Println("ano bissexto 2028:", time.Date(2028, 3, 0, 0, 0, 0, 0, time.UTC).Format(time.DateOnly))

	// fusos fixos (sem regras de verao)
	brt := time.FixedZone("BRT", -3*3600)
	fmt.Println("\nFixedZone:", time.Date(2026, 10, 4, 12, 0, 0, 0, brt))

	// duracoes
	d := 90*time.Minute + 30*time.Second
	fmt.Println("\nDuration:        ", d, "| Hours:", d.Hours(), "| Minutes:", d.Minutes())
	fmt.Println("Truncate/Round:  ", d.Truncate(time.Hour), d.Round(time.Hour))
	pd, err := time.ParseDuration("1h15m30.5s")
	fmt.Println("ParseDuration:   ", pd, err)
	_, err = time.ParseDuration("2 dias")
	fmt.Println("ParseDuration:   ", err)
	fmt.Println("int * Duration:  ", time.Duration(3)*time.Second, "(precisa converter o int)")
	fmt.Println("maior Duration:  ", time.Duration(1<<63-1), "= ~292 anos")

	// diferenca em dias de calendario (nao em 24h)
	a := time.Date(2026, 3, 7, 0, 0, 0, 0, ny)
	b := time.Date(2026, 3, 9, 0, 0, 0, 0, ny)
	fmt.Println("\nSub em horas:    ", b.Sub(a), "(47h por causa do horario de verao)")
	fmt.Println("dias de calendario:", int(b.Sub(a).Hours()/24), "<- errado (arredonda p/ baixo)")
	fmt.Println("dias corretos:     ", diasEntre(a, b))
}

// diasEntre conta dias de calendario, ignorando horario de verao.
func diasEntre(a, b time.Time) int {
	a = time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
	b = time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
	return int(b.Sub(a).Hours() / 24)
}
