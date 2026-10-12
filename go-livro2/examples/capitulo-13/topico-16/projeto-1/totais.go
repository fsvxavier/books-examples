package main

import (
	"encoding/csv"
	"fmt"
	"html"
	"io"
)

// Totais fecha o relatório.
type Totais struct{ Valor, Imposto int }

// Formato são os passos que variam (o Template Method). Os passos do
// meio recebem dados prontos: o formato não decide NADA fiscal.
type Formato interface {
	Cabecalho(w io.Writer) error
	Linha(w io.Writer, desc string, valor, imposto int) error
	Rodape(w io.Writer, t Totais) error
}

// Emitir é o molde: fixa a ordem e o cálculo; o Visitor calcula por item.
func Emitir(w io.Writer, f Formato, itens []Item) (Totais, error) {
	var t Totais
	if err := f.Cabecalho(w); err != nil {
		return t, err
	}
	for _, it := range itens {
		desc := Aceitar[string](it, Descricao{})
		v := Aceitar[int](it, Valor{})
		imp := Aceitar[int](it, Imposto{})
		t.Valor += v
		t.Imposto += imp
		if err := f.Linha(w, desc, v, imp); err != nil {
			return t, err
		}
	}
	return t, f.Rodape(w, t)
}

func reais(c int) string { return fmt.Sprintf("%d,%02d", c/100, c%100) }

type Texto struct{}

func (Texto) Cabecalho(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%-34s %10s %9s\n", "ITEM", "VALOR", "IMPOSTO")
	return err
}
func (Texto) Linha(w io.Writer, d string, v, i int) error {
	_, err := fmt.Fprintf(w, "%-34s %10s %9s\n", d, reais(v), reais(i))
	return err
}
func (Texto) Rodape(w io.Writer, t Totais) error {
	_, err := fmt.Fprintf(w, "%-34s %10s %9s\n", "TOTAL", reais(t.Valor), reais(t.Imposto))
	return err
}

type CSV struct{}

// campos escreve uma linha com encoding/csv (separador ';'), que cuida
// das aspas e dos separadores dentro dos campos.
func campos(w io.Writer, c ...string) error {
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	if err := cw.Write(c); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

func (CSV) Cabecalho(w io.Writer) error { return campos(w, "item", "valor", "imposto") }
func (CSV) Linha(w io.Writer, d string, v, i int) error {
	return campos(w, d, reais(v), reais(i))
}
func (CSV) Rodape(w io.Writer, t Totais) error {
	return campos(w, "total", reais(t.Valor), reais(t.Imposto))
}

type HTML struct{}

func (HTML) Cabecalho(w io.Writer) error {
	_, err := io.WriteString(w, "<table>\n")
	return err
}
func (HTML) Linha(w io.Writer, d string, v, i int) error {
	_, err := fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n", html.EscapeString(d), reais(v), reais(i))
	return err
}
func (HTML) Rodape(w io.Writer, t Totais) error {
	_, err := fmt.Fprintf(w, "<tr><th>total</th><th>%s</th><th>%s</th></tr>\n</table>\n", reais(t.Valor), reais(t.Imposto))
	return err
}
