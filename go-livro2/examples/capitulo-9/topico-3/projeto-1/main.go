package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	auth := &Autenticacao{Usuarios: map[string]string{"ana": "123", "adm": "root"}}
	limite := NovoLimite(2)
	san := &Sanitizacao{}
	cache := NovoCache()
	pedidos := &Pedidos{}

	// Montagem da cadeia: o cliente liga os elos.
	auth.DefinirProximo(limite).DefinirProximo(san).DefinirProximo(cache).DefinirProximo(pedidos)
	var cadeia Manipulador = auth

	ctx := context.Background()
	casos := []struct {
		nome string
		r    Requisicao
	}{
		{"ok", Requisicao{Usuario: "ana", Senha: "123", IP: "1.1.1.1", Corpo: " 2 canetas "}},
		{"repetida (cache)", Requisicao{Usuario: "ana", Senha: "123", IP: "1.1.1.1", Corpo: "2 canetas"}},
		{"senha errada", Requisicao{Usuario: "ana", Senha: "x", IP: "9.9.9.9", Corpo: "1 caderno"}},
		{"marcacao no corpo", Requisicao{Usuario: "ana", Senha: "123", IP: "1.1.1.1", Corpo: "<script>"}},
	}
	for _, c := range casos {
		r := c.r
		err := cadeia.Tratar(ctx, &r)
		fmt.Printf("%-18s -> err=%v\n", c.nome, err)
	}
	fmt.Println("pedidos criados:", pedidos.Criados, "| acertos de cache:", cache.Acertos)

	// A ordem faz diferenca: se o limite de falhas ficasse DEPOIS da autenticacao, nunca veria as falhas.
	fmt.Println("--- mesma requisicao ruim, 3 vezes, com o limite por DENTRO da autenticacao (limite=2):")
	for i := 1; i <= 3; i++ {
		r := Requisicao{Usuario: "ana", Senha: "errada", IP: "9.9.9.9", Corpo: "x"}
		err := cadeia.Tratar(ctx, &r)
		fmt.Printf("  tentativa %d: bloqueado=%v err=%v\n", i, errors.Is(err, ErrBloqueado), err)
	}

	fmt.Println("--- cadeia reordenada (limite POR FORA da autenticacao):")
	auth2 := &Autenticacao{Usuarios: map[string]string{"ana": "123"}}
	limite2 := NovoLimite(2)
	limite2.DefinirProximo(auth2).DefinirProximo(&Pedidos{})
	for i := 1; i <= 3; i++ {
		r := Requisicao{Usuario: "ana", Senha: "errada", IP: "9.9.9.9", Corpo: "x"}
		err := limite2.Tratar(ctx, &r)
		fmt.Printf("  tentativa %d: bloqueado=%v err=%v\n", i, errors.Is(err, ErrBloqueado), err)
	}
}
