package main

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-11/topico-3/projeto-2/doc"
)

func main() {
	a := doc.Novo("A")
	b := doc.Novo("B")
	a.Escrever("texto de A")
	b.Escrever("texto de B")

	// Interface selada: o guardião só guarda Snapshots.
	var pilha []doc.Snapshot
	pilha = append(pilha, a.Foto())
	a.Escrever(" + mais")
	fmt.Println("antes:", a, "| foto:", pilha[0].Descricao())
	a.Voltar(pilha[0])
	fmt.Println("depois de voltar:", a)

	// A foto de A aplicada em B: a forma 2 aceita, a checada recusa.
	b.Voltar(pilha[0])
	fmt.Println("B depois de Voltar(foto de A):", b)
	c := doc.Novo("C")
	fmt.Println("VoltarChecado:", c.VoltarChecado(pilha[0]))

	// Estrito: o memento se restaura sozinho no originador certo.
	e := a.Estrito()
	a.Escrever(" !!!")
	fmt.Println("estrito, antes:", a)
	e.Restaurar()
	fmt.Println("estrito, depois:", a)
}
