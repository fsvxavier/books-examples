package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Produto struct {
	SKU   string
	Nome  string
	Preco int64 // centavos
	Qtd   int
}

// ErroLinha identifica exatamente onde o problema está.
type ErroLinha struct {
	Linha int
	Campo string
	Err   error
}

func (e *ErroLinha) Error() string {
	return fmt.Sprintf("linha %d, campo %q: %v", e.Linha, e.Campo, e.Err)
}

func (e *ErroLinha) Unwrap() error { return e.Err }

var (
	ErrVazio    = errors.New("valor obrigatório")
	ErrNegativo = errors.New("não pode ser negativo")
)

// Importar lê produtos de um CSV. Uma linha ruim NÃO interrompe a importação:
// as linhas válidas são devolvidas junto com todos os erros encontrados.
func Importar(r io.Reader, maxErros int) ([]Produto, error) {
	leitor := csv.NewReader(r)
	leitor.FieldsPerRecord = 4

	var produtos []Produto
	var erros []error
	linha := 0

	for {
		registro, err := leitor.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		linha++
		if linha == 1 {
			continue // cabeçalho
		}
		if err != nil {
			erros = append(erros, &ErroLinha{linha, "(registro)", err})
			continue
		}

		p, err := converter(registro)
		if err != nil {
			var el *ErroLinha
			if errors.As(err, &el) {
				el.Linha = linha
			}
			erros = append(erros, err)
			if len(erros) >= maxErros {
				erros = append(erros, fmt.Errorf("importação interrompida após %d erros", maxErros))
				break
			}
			continue
		}
		produtos = append(produtos, p)
	}
	return produtos, errors.Join(erros...)
}

func converter(reg []string) (Produto, error) {
	var p Produto
	p.SKU = strings.TrimSpace(reg[0])
	if p.SKU == "" {
		return p, &ErroLinha{Campo: "sku", Err: ErrVazio}
	}
	p.Nome = strings.TrimSpace(reg[1])

	preco, err := strconv.ParseInt(strings.TrimSpace(reg[2]), 10, 64)
	if err != nil {
		return p, &ErroLinha{Campo: "preco", Err: err}
	}
	if preco < 0 {
		return p, &ErroLinha{Campo: "preco", Err: ErrNegativo}
	}
	p.Preco = preco

	qtd, err := strconv.Atoi(strings.TrimSpace(reg[3]))
	if err != nil {
		return p, &ErroLinha{Campo: "qtd", Err: err}
	}
	p.Qtd = qtd
	return p, nil
}

const dados = `sku,nome,preco,qtd
A1,Caneca,2990,10
,Sem SKU,1000,1
B2,Camiseta,abc,5
C3,Adesivo,-50,3
D4,Caderno,1890,x
E5,Boné,4500,7
F6,Só três campos
G7,Meia,990,20`

func main() {
	produtos, err := Importar(strings.NewReader(dados), 10)
	fmt.Println("importados:", len(produtos))
	for _, p := range produtos {
		fmt.Printf("  %+v\n", p)
	}
	fmt.Println("erros:")
	fmt.Println(err)

	fmt.Println(errors.Is(err, ErrVazio), errors.Is(err, ErrNegativo), errors.Is(err, strconv.ErrSyntax))

	_, err = Importar(strings.NewReader(dados), 2)
	fmt.Println("---")
	fmt.Println(err)
}
