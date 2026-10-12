package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Relogio interface{ Agora() time.Time }

type RelogioReal struct{}

func (RelogioReal) Agora() time.Time { return time.Now() }

type RelogioFixo struct{ T time.Time }

func (r RelogioFixo) Agora() time.Time { return r.T }

type Cupom struct {
	Codigo string
	Expira time.Time
}

// Valido recebe o relogio: nada de time.Now() escondido dentro da regra.
func (c Cupom) Valido(r Relogio) bool { return r.Agora().Before(c.Expira) }

func main() {
	// o horario de parede e o relogio monotonico
	t := time.Now()
	fmt.Println("Now() mostra o monotonico:", strings.Contains(t.String(), "m=+"))
	fmt.Println("Round(0) remove:          ", !strings.Contains(t.Round(0).String(), "m=+"))
	inicio := time.Now()
	time.Sleep(20 * time.Millisecond)
	fmt.Println("time.Since usa o monotonico, ~20ms:", time.Since(inicio).Round(10*time.Millisecond))

	// time.Time como chave de mapa: == compara instante E fuso E monotonico
	utc := time.Date(2026, 10, 4, 15, 0, 0, 0, time.UTC)
	sp := utc.In(time.FixedZone("BRT", -3*3600))
	m := map[time.Time]string{utc: "utc"}
	_, achou := m[sp]
	fmt.Println("\nmesmo instante como chave de mapa, outro fuso:", achou, "<- nao acha")
	m2 := map[time.Time]string{utc.UTC(): "utc"}
	_, achou = m2[sp.UTC()]
	fmt.Println("normalizando com .UTC():                      ", achou)

	// JSON
	c := Cupom{Codigo: "PROMO10", Expira: time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)}
	b, _ := json.Marshal(c)
	fmt.Println("\nJSON:", string(b))
	var lido Cupom
	json.Unmarshal([]byte(`{"Codigo":"X","Expira":"2026-12-31T20:59:59-03:00"}`), &lido)
	fmt.Println("lido com -03:00:", lido.Expira, "| em UTC:", lido.Expira.UTC())
	err := json.Unmarshal([]byte(`{"Expira":"31/12/2026"}`), &lido)
	fmt.Println("data fora de RFC 3339:", err)

	// relogio injetado: o teste controla o tempo, sem time.Sleep
	antes := RelogioFixo{time.Date(2026, 12, 31, 23, 59, 58, 0, time.UTC)}
	depois := RelogioFixo{time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)}
	fmt.Println("\ncupom valido antes do prazo:  ", c.Valido(antes))
	fmt.Println("cupom valido depois do prazo: ", c.Valido(depois))
	_ = RelogioReal{} // em producao, o main injeta o relogio real
}
