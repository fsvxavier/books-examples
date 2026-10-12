package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// saudacoes mapeia o código do idioma para o texto de cumprimento.
var saudacoes = map[string]string{
	"pt": "Olá",
	"en": "Hello",
	"es": "Hola",
}

func main() {
	nome := flag.String("nome", "", "nome da pessoa a cumprimentar (obrigatório)")
	idioma := flag.String("idioma", "pt", "idioma da saudação: pt, en ou es")
	flag.Parse()

	if strings.TrimSpace(*nome) == "" {
		fmt.Fprintln(os.Stderr, "erro: informe o nome com -nome")
		flag.Usage()
		os.Exit(2)
	}

	texto, ok := saudacoes[*idioma]
	if !ok {
		fmt.Fprintf(os.Stderr, "erro: idioma %q não suportado\n", *idioma)
		os.Exit(2)
	}

	fmt.Printf("%s, %s!\n", texto, *nome)
}
