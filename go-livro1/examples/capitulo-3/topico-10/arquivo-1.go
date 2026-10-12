package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	leitor := bufio.NewScanner(os.Stdin)
	var acumulado float64

	fmt.Println("Comandos: somar N, subtrair N, mostrar, zerar, sair")

laco:
	for leitor.Scan() {
		linha := strings.TrimSpace(leitor.Text())
		if linha == "" {
			continue
		}

		partes := strings.Fields(linha)
		comando := strings.ToLower(partes[0])

		switch comando {
		case "somar", "subtrair":
			if len(partes) != 2 {
				fmt.Printf("uso: %s N\n", comando)
				continue
			}
			n, err := strconv.ParseFloat(partes[1], 64)
			if err != nil {
				fmt.Printf("número inválido: %q\n", partes[1])
				continue
			}
			if comando == "somar" {
				acumulado += n
			} else {
				acumulado -= n
			}
		case "mostrar":
			fmt.Printf("acumulado: %.2f\n", acumulado)
		case "zerar":
			acumulado = 0
			fmt.Println("zerado")
		case "sair":
			break laco
		default:
			fmt.Printf("comando desconhecido: %q\n", comando)
		}
	}

	fmt.Printf("resultado final: %.2f\n", acumulado)
}
