package main

import (
	"fmt"
	"strings"
	"time"
)

// Handler é qualquer função que transforma um texto em outro.
type Handler func(string) string

// Middleware recebe um Handler e devolve um Handler com comportamento extra.
type Middleware func(Handler) Handler

func ComLog(nome string) Middleware {
	return func(proximo Handler) Handler {
		return func(entrada string) string {
			fmt.Printf("[%s] entrada=%q\n", nome, entrada)
			saida := proximo(entrada)
			fmt.Printf("[%s] saida=%q\n", nome, saida)
			return saida
		}
	}
}

func ComTempo(proximo Handler) Handler {
	return func(entrada string) string {
		inicio := time.Now()
		saida := proximo(entrada)
		decorrido := time.Since(inicio)
		fmt.Println("demorou menos de 1s:", decorrido < time.Second)
		return saida
	}
}

func ComCache(proximo Handler) Handler {
	cache := make(map[string]string)
	return func(entrada string) string {
		if v, ok := cache[entrada]; ok {
			fmt.Println("cache hit")
			return v
		}
		v := proximo(entrada)
		cache[entrada] = v
		return v
	}
}

// Encadear aplica os middlewares de fora para dentro: o primeiro da lista é o mais externo.
func Encadear(h Handler, mws ...Middleware) Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func main() {
	base := func(s string) string { return strings.ToUpper(s) }

	h := Encadear(base, ComLog("externo"), ComCache, ComTempo)
	fmt.Println(h("go"))
	fmt.Println(h("go"))
}
