package main

import (
	"errors"
	"fmt"

	"exemplos/livro2/capitulo-11/topico-3/projeto-1/editor"
	"exemplos/livro2/capitulo-11/topico-3/projeto-1/historico"
)

func main() {
	ed := editor.Novo("rascunho")
	h := historico.Novo[editor.Memento](10)

	h.Guardar(ed.Salvar())
	ed.Escrever("Olá, mundo.")
	h.Guardar(ed.Salvar())
	ed.NovaLinha()
	ed.Escrever("Segunda linha.")
	ed.Renomear("final")
	fmt.Println("agora:   ", ed)
	fmt.Printf("texto:\n%s\n", ed.Texto())

	for {
		m, err := h.Desfazer()
		if errors.Is(err, historico.ErrVazio) {
			fmt.Println("histórico vazio")
			break
		}
		ed.Restaurar(m)
		fmt.Printf("restaurado: %s | %s | texto=%q\n", m.Descricao(), ed, ed.Texto())
	}
}
