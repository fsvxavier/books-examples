package main

import (
	"fmt"
	"os"
)

func main() {
	pedido := []Item{
		Livro{"Go: Do Zero à Produção", 12000},
		Eletronico{"Fone <Pro>", 45000, 12},
		Alimento{"Café", 3500, false},
		Alimento{"Queijo fresco", 2800, true},
		Servico{"Instalação", 3, 9000},
	}
	for _, f := range []struct {
		nome string
		fmt  Formato
	}{{"texto", Texto{}}, {"csv", CSV{}}, {"html", HTML{}}} {
		fmt.Printf("--- %s ---\n", f.nome)
		t, err := Emitir(os.Stdout, f.fmt, pedido)
		if err != nil {
			fmt.Println("erro:", err)
		}
		_ = t
	}
}
