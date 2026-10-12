package main

import (
	"fmt"
	"time"
)

func main() {
	// um instante fixo para a saida ser reproduzivel
	t := time.Date(2026, time.October, 4, 14, 30, 45, 123456789, time.UTC)
	fmt.Println("String():   ", t)
	fmt.Println("Unix:       ", t.Unix(), " UnixMilli:", t.UnixMilli())
	fmt.Println("componentes:", t.Year(), t.Month(), t.Day(), t.Weekday(), "dia do ano", t.YearDay())
	y, w := t.ISOWeek()
	fmt.Println("semana ISO: ", y, w)

	// layouts: o "2006-01-02 15:04:05" e a data de referencia, nao um modelo qualquer
	fmt.Println("\nRFC3339:    ", t.Format(time.RFC3339))
	fmt.Println("RFC3339Nano:", t.Format(time.RFC3339Nano))
	fmt.Println("DateTime:   ", t.Format(time.DateTime))
	fmt.Println("DateOnly:   ", t.Format(time.DateOnly))
	fmt.Println("TimeOnly:   ", t.Format(time.TimeOnly))
	fmt.Println("brasileiro: ", t.Format("02/01/2006 15:04"))
	fmt.Println("extenso:    ", t.Format("Monday, 2 January 2006"))
	fmt.Println("12h:        ", t.Format("3:04PM"))

	// erro classico: usar um layout inventado
	fmt.Println("\nlayout errado 'yyyy-mm-dd':", t.Format("yyyy-mm-dd"))
	fmt.Println("layout errado '2006-02-01': ", t.Format("2006-02-01"), "<- dia e mes trocados")

	// Parse
	p, err := time.Parse("02/01/2006", "04/10/2026")
	fmt.Println("\nParse dd/mm/aaaa:", p, err)
	_, err = time.Parse("02/01/2006", "31/02/2026")
	fmt.Println("31/02:           ", err)
	_, err = time.Parse(time.DateOnly, "2026-13-01")
	fmt.Println("mes 13:          ", err)
	_, err = time.Parse(time.RFC3339, "2026-10-04 14:30:45")
	fmt.Println("formato errado:  ", err)

	// o fuso na leitura: Parse assume UTC se nao houver fuso na string
	sp, _ := time.LoadLocation("America/Sao_Paulo")
	a, _ := time.Parse(time.DateTime, "2026-10-04 12:00:00")
	b, _ := time.ParseInLocation(time.DateTime, "2026-10-04 12:00:00", sp)
	fmt.Println("\nParse:           ", a)
	fmt.Println("ParseInLocation: ", b)
	fmt.Println("sao o mesmo instante?", a.Equal(b), " diferenca:", a.Sub(b))

	// comparar: Equal, nao ==
	utc := time.Date(2026, 10, 4, 15, 0, 0, 0, time.UTC)
	local := utc.In(sp)
	fmt.Println("\nmesmo instante em fusos diferentes:")
	fmt.Println("  utc:  ", utc)
	fmt.Println("  sp:   ", local)
	fmt.Println("  ==    ", utc == local, "| Equal:", utc.Equal(local), "| Compare:", utc.Compare(local))

	// truncar e arredondar
	fmt.Println("\nTruncate(1h):", t.Truncate(time.Hour).Format(time.TimeOnly))
	fmt.Println("Round(1h):   ", t.Round(time.Hour).Format(time.TimeOnly))
	inicioDia := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	fmt.Println("inicio do dia:", inicioDia)

	// valor zero
	var z time.Time
	fmt.Println("\nzero value:  ", z, "IsZero:", z.IsZero(), "Unix:", z.Unix())
}
