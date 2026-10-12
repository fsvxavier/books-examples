package main

import (
	"fmt"
	"math"
)

func main() {
	// Constantes literais são avaliadas com precisão exata em tempo de compilação.
	a := 0.1 + 0.2
	fmt.Println("constante:", a == 0.3)

	// Com variáveis, a soma acontece em float64 em tempo de execução.
	x, y := 0.1, 0.2
	soma := x + y
	fmt.Println("variáveis:", soma == 0.3)
	fmt.Printf("soma = %.20f\n", soma)
	fmt.Println(soma)

	fmt.Println("quase igual:", math.Abs(soma-0.3) < 1e-9)
}
