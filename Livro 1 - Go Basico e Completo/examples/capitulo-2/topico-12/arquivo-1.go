package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Centavos guarda dinheiro como inteiro, evitando os erros de arredondamento do float.
type Centavos int64

// String formata o valor como moeda brasileira, por exemplo "R$ 1234,50".
func (c Centavos) String() string {
	sinal := ""
	if c < 0 {
		sinal = "-"
		c = -c
	}
	return fmt.Sprintf("%sR$ %d,%02d", sinal, c/100, c%100)
}

// ParseReais converte um texto como "19,90" em centavos.
func ParseReais(texto string) (Centavos, error) {
	reais, centavos, temVirgula := strings.Cut(strings.TrimSpace(texto), ",")

	r, err := strconv.ParseInt(reais, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parte inteira inválida em %q: %w", texto, err)
	}

	var c int64
	if temVirgula {
		if len(centavos) != 2 {
			return 0, fmt.Errorf("use exatamente 2 casas decimais em %q", texto)
		}
		c, err = strconv.ParseInt(centavos, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("centavos inválidos em %q: %w", texto, err)
		}
	}
	return Centavos(r*100 + c), nil
}

// ItemCarrinho representa uma linha do pedido.
type ItemCarrinho struct {
	Descricao  string
	PrecoUnit  Centavos
	Quantidade int
}

func (i ItemCarrinho) Subtotal() Centavos {
	return i.PrecoUnit * Centavos(i.Quantidade)
}

// Desconto aplica um percentual em pontos-base (1% = 100) com aritmética inteira.
// O resultado é truncado para baixo, o que favorece quem vende; outra política é possível.
func Desconto(valor Centavos, pontosBase int64) Centavos {
	return valor * Centavos(pontosBase) / 10000
}

func main() {
	var soma float64
	for range 10 {
		soma += 0.1
	}
	fmt.Println("float: 10 x 0,10 =", soma, soma == 1.0)

	var total Centavos
	for range 10 {
		total += 10
	}
	fmt.Println("centavos: 10 x R$ 0,10 =", total, total == 100)

	preco, err := ParseReais("19,90")
	fmt.Println(preco, err)
	_, err = ParseReais("19,9")
	fmt.Println(err)
	_, err = ParseReais("abc")
	fmt.Println(err)

	itens := []ItemCarrinho{
		{"Caneca", 2990, 2},
		{"Camiseta", 5990, 1},
		{"Adesivo", 350, 10},
	}

	var subtotal Centavos
	for _, it := range itens {
		fmt.Printf("%-10s %2d x %-10v = %v\n", it.Descricao, it.Quantidade, it.PrecoUnit, it.Subtotal())
		subtotal += it.Subtotal()
	}

	desc := Desconto(subtotal, 1250) // 12,5%
	fmt.Println("Subtotal:", subtotal)
	fmt.Println("Desconto 12,5%:", desc)
	fmt.Println("Total:", subtotal-desc)
	fmt.Println(Centavos(-1505))
}
