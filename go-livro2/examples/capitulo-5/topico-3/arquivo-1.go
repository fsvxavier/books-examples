package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

// --- O sistema existente: recebe cotacoes em XML ---

type Cotacao struct {
	XMLName xml.Name `xml:"cotacao"`
	Ativo   string   `xml:"ativo,attr"`
	Preco   float64  `xml:"preco"`
	Volume  int64    `xml:"volume"`
}

type Fonte interface {
	Cotacoes() ([]Cotacao, error)
}

type fonteXML struct{ doc string }

func (f fonteXML) Cotacoes() ([]Cotacao, error) {
	var envelope struct {
		Itens []Cotacao `xml:"cotacao"`
	}
	if err := xml.Unmarshal([]byte(f.doc), &envelope); err != nil {
		return nil, fmt.Errorf("lendo XML: %w", err)
	}
	return envelope.Itens, nil
}

// --- A biblioteca de terceiros (nao podemos mudar): so aceita JSON ---

type AnalisadorJSON struct{}

func (AnalisadorJSON) Analisar(dados []byte) (string, error) {
	var linhas []struct {
		Ativo  string  `json:"ativo"`
		Preco  float64 `json:"preco"`
		Volume int64   `json:"volume"`
	}
	if err := json.Unmarshal(dados, &linhas); err != nil {
		return "", fmt.Errorf("analisador: %w", err)
	}
	var total float64
	var b strings.Builder
	for _, l := range linhas {
		total += l.Preco * float64(l.Volume)
		fmt.Fprintf(&b, "%s: %.2f x %d\n", l.Ativo, l.Preco, l.Volume)
	}
	fmt.Fprintf(&b, "volume financeiro: %.2f", total)
	return b.String(), nil
}

// --- A interface que o NOSSO codigo espera (a "interface do cliente") ---

type Analista interface {
	Relatorio(fonte Fonte) (string, error)
}

// --- O Adapter: implementa Analista, embrulha o analisador de terceiros ---

type AdaptadorAnalise struct {
	servico AnalisadorJSON
}

func (a AdaptadorAnalise) Relatorio(fonte Fonte) (string, error) {
	cotacoes, err := fonte.Cotacoes()
	if err != nil {
		return "", err
	}
	// Traduz o formato: structs lidas de XML viram JSON.
	type item struct {
		Ativo  string  `json:"ativo"`
		Preco  float64 `json:"preco"`
		Volume int64   `json:"volume"`
	}
	itens := make([]item, 0, len(cotacoes))
	for _, c := range cotacoes {
		itens = append(itens, item{c.Ativo, c.Preco, c.Volume})
	}
	dados, err := json.Marshal(itens)
	if err != nil {
		return "", err
	}
	return a.servico.Analisar(dados)
}

func main() {
	const doc = `<carteira>
  <cotacao ativo="PETR4"><preco>38.50</preco><volume>1200</volume></cotacao>
  <cotacao ativo="VALE3"><preco>61.10</preco><volume>800</volume></cotacao>
</carteira>`

	var analista Analista = AdaptadorAnalise{}
	rel, err := analista.Relatorio(fonteXML{doc})
	if err != nil {
		fmt.Println("erro:", err)
		return
	}
	fmt.Println(rel)

	_, err = analista.Relatorio(fonteXML{"<carteira><cotacao"})
	fmt.Println("XML quebrado:", err)
}
