package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-7/topico-7/projeto-1/internal/feature"
)

// exportar grava em CSV as features de comportamento de cada transação dos
// primeiros cartões, calculadas como no serviço (antes de a compra entrar).
// É o arquivo que o Feast lê como fonte offline.
func exportar() {
	dest := "features.csv"
	if len(os.Args) > 2 {
		dest = os.Args[2]
	}
	txs := dados()
	calc := feature.NovaCalculadora()
	f, err := os.Create(dest)
	must(err)
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"cartao", "event_timestamp", "tx_5m", "soma_5m", "tx_1h", "soma_1h", "tx_24h", "soma_24h"})
	n := 0
	for _, t := range txs {
		c := calc.Comportamento(t.Cartao, t.Quando)
		calc.Observar(t.Cartao, t.Quando, t.Centav)
		if t.Cartao >= "c00100" || t.Quando.After(inicio.Add(48*time.Hour)) {
			continue
		}
		reg := []string{t.Cartao, t.Quando.Format("2006-01-02 15:04:05")}
		for _, a := range c {
			reg = append(reg, strconv.FormatInt(a.Contagem, 10), strconv.FormatInt(a.Centav, 10))
		}
		w.Write(reg)
		n++
	}
	w.Flush()
	fmt.Println(n, "linhas em", dest)
}
