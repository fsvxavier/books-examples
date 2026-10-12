package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx := context.Background()

	// O cliente so conhece a interface Repositorio.
	var repo Repositorio = NovoProxyPreguicoso(abrir)

	inicio := time.Now()
	fmt.Println("criar o proxy custa:", time.Since(inicio) < 10*time.Millisecond, "(nada foi aberto: aberturas =", aberturas, ")")

	inicio = time.Now()
	v, _ := repo.Buscar(ctx, 1)
	fmt.Printf("1a chamada: %q em ~%v | aberturas = %d\n", v, time.Since(inicio).Round(100*time.Millisecond), aberturas)

	inicio = time.Now()
	v, _ = repo.Buscar(ctx, 2)
	fmt.Printf("2a chamada: %q em menos de 10ms: %v | aberturas = %d\n", v, time.Since(inicio) < 10*time.Millisecond, aberturas)

	repo.Close()

	// Um proxy que nunca e usado nao abre nada, e fechar nao faz nada.
	nunca := NovoProxyPreguicoso(abrir)
	fmt.Println("Close de proxy nunca usado:", nunca.Close(), "| aberturas =", aberturas)

	// Falha na abertura: nao fica guardada; a proxima chamada tenta de novo.
	tentativas := 0
	instavel := NovoProxyPreguicoso(func(ctx context.Context) (Repositorio, error) {
		tentativas++
		if tentativas < 3 {
			return nil, ErrIndisponivel
		}
		return &repoReal{id: 99}, nil
	})
	for i := 1; i <= 4; i++ {
		v, err := instavel.Buscar(ctx, i)
		fmt.Printf("instavel, chamada %d: %q err=%v\n", i, v, err)
	}
}
