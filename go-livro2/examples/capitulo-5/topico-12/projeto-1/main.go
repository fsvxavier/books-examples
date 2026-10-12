package main

import "fmt"

func main() {
	d := Dados{Periodo: "2026-09", Pedidos: 128, Faturamento: 15432.9, Cancelados: 7, MaiorPedido: 980, Cliente: "Tom & Jerry <ltda>"}

	novoTexto := func() Renderizador { return &Texto{} }
	novoJSON := func() Renderizador { return &JSON{} }
	novoHTML := func() Renderizador { return &HTML{} }

	// A MESMA abstracao com tres implementacoes.
	for _, nome := range []struct {
		n string
		f func() Renderizador
	}{{"texto", novoTexto}, {"json", novoJSON}, {"html", novoHTML}} {
		fmt.Printf("--- Resumo em %s\n%s\n", nome.n, Resumo{nome.f}.Gerar(d))
	}

	// A MESMA implementacao com outra abstracao.
	fmt.Printf("--- Detalhado em texto\n%s\n", Detalhado{novoTexto}.Gerar(d))
	fmt.Printf("--- Detalhado em html\n%s\n", Detalhado{novoHTML}.Gerar(d))
}
