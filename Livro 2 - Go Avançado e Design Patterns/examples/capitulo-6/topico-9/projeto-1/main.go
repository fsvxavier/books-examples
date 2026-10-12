package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	ctx := context.Background()

	// A funcao "cara": falha na 1a chamada de cada CEP (rede instavel) e depois acerta.
	var chamadas int
	vistos := map[string]bool{}
	buscar := func(ctx context.Context, cep string) (string, error) {
		chamadas++
		if !vistos[cep] {
			vistos[cep] = true
			return "", errors.New("timeout")
		}
		return "Rua do CEP " + cep, nil
	}

	var linhas []string
	reg := func(s string) { linhas = append(linhas, s) }
	zerar := func() { chamadas, linhas, vistos = 0, nil, map[string]bool{} }

	run := func(titulo string, f Func[string, string]) {
		zerar()
		fmt.Println(titulo)
		for _, cep := range []string{"01310", "01310", "20040"} {
			v, err := f(ctx, cep)
			fmt.Printf("  cliente pediu %s -> %q %v\n", cep, v, err)
		}
		fmt.Printf("  chamadas a funcao cara: %d | anotacoes: %d\n", chamadas, len(linhas))
		for _, l := range linhas {
			fmt.Println("    ", l)
		}
	}

	// Anotar por FORA: registra o que o CLIENTE enxerga (uma linha por pedido).
	run("Anotar por fora:", Aplicar(buscar, Anotar[string, string]("cliente", reg), Memoizar[string, string](), Retentar[string, string](3)))
	// Anotar por DENTRO: registra o que chega a funcao cara (cada tentativa; o cache nao aparece).
	run("Anotar por dentro:", Aplicar(buscar, Memoizar[string, string](), Retentar[string, string](3), Anotar[string, string]("funcao", reg)))
}
