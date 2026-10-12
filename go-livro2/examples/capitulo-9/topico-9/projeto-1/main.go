package main

import (
	"errors"
	"fmt"
)

func main() {
	// O receptor: um documento (aqui, so um texto).
	texto := "Ola, mundo"

	// UM comando, tres remetentes (o problema do catalogo: botao, menu e atalho).
	var anterior string
	exclamar := Par{Nome: "exclamar",
		Faz:    func() error { anterior = texto; texto += "!"; return nil },
		Desfaz: func() error { texto = anterior; return nil },
	}
	botao := &Remetente{Nome: "botao da barra", Comando: exclamar}
	menu := &Remetente{Nome: "item de menu", Comando: exclamar}
	atalho := &Remetente{Nome: "Ctrl+M", Comando: exclamar}

	botao.Acionar()
	fmt.Printf("depois do botao:  %q\n", texto)
	menu.Acionar()
	fmt.Printf("depois do menu:   %q\n", texto)
	atalho.Acionar()
	fmt.Printf("depois do atalho: %q\n", texto)

	// Trocar o comando de um remetente muda o comportamento dele em tempo de execucao.
	atalho.Comando = Par{Nome: "limpar", Faz: func() error { texto = ""; return nil }, Desfaz: func() error { return nil }}
	atalho.Acionar()
	fmt.Printf("atalho com outro comando: %q\n", texto)

	var vazio Remetente
	vazio.Nome = "botao sem comando"
	fmt.Println("remetente sem comando:", vazio.Acionar())

	fmt.Println("--- Macro: tudo ou nada")
	var log []string
	passo := func(nome string, falha bool) Comando {
		return Par{Nome: nome,
			Faz: func() error {
				if falha {
					return errors.New("disco cheio")
				}
				log = append(log, "fez "+nome)
				return nil
			},
			Desfaz: func() error { log = append(log, "desfez "+nome); return nil },
		}
	}

	ok := &Macro{Nome: "exportar", Comandos: []Comando{passo("A", false), passo("B", false), passo("C", false)}}
	fmt.Println("macro ok:", ok.Executar(), log)

	log = nil
	ruim := &Macro{Nome: "exportar", Comandos: []Comando{passo("A", false), passo("B", false), passo("C", true), passo("D", false)}}
	err := ruim.Executar()
	fmt.Println("macro que falha no passo 3:", err)
	fmt.Println("efeitos:", log)

	log = nil
	ok.Executar()
	log = nil
	ok.Desfazer()
	fmt.Println("desfazer a macro inteira:", log)
}
