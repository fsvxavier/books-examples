package main

import (
	"fmt"
	"iter"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
)

// Singleton: sync.OnceValue (cap. 27)
var config = sync.OnceValue(func() map[string]string {
	fmt.Println("  (carregando configuracao uma unica vez)")
	return map[string]string{"ambiente": "lab"}
})

// Strategy: um tipo funcao (cap. 35)
type Ordenacao func(a, b string) int

// Iterator: iter.Seq, consumido com range (cap. 33)
func Pares(max int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i <= max; i += 2 {
			if !yield(i) {
				return
			}
		}
	}
}

// Command: uma closure que guarda o que fazer e como desfazer (cap. 32)
type Comando struct {
	Fazer    func()
	Desfazer func()
}

// Observer: callbacks registrados (cap. 34)
type Assunto struct{ ouvintes []func(string) }

func (a *Assunto) Assinar(f func(string)) { a.ouvintes = append(a.ouvintes, f) }
func (a *Assunto) Emitir(m string) {
	for _, f := range a.ouvintes {
		f(m)
	}
}

// Decorator: middleware que embrulha um http.Handler (cap. 29)
func ComCabecalho(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Decorado", "sim")
		next.ServeHTTP(w, r)
	})
}

// Factory: uma funcao que devolve uma interface (cap. 25)
type Animal interface{ Som() string }
type cao struct{}
type gato struct{}

func (cao) Som() string  { return "au" }
func (gato) Som() string { return "miau" }

func NovoAnimal(tipo string) (Animal, error) {
	switch tipo {
	case "cao":
		return cao{}, nil
	case "gato":
		return gato{}, nil
	}
	return nil, fmt.Errorf("animal desconhecido: %q", tipo)
}

func main() {
	fmt.Println("Singleton:", config()["ambiente"], config()["ambiente"])

	nomes := []string{"zeca", "ana", "bia"}
	porTamanhoDepoisNome := Ordenacao(func(a, b string) int {
		if c := len(a) - len(b); c != 0 {
			return c
		}
		return strings.Compare(a, b)
	})
	slices.SortFunc(nomes, porTamanhoDepoisNome)
	fmt.Println("Strategy: ", nomes)

	var soma int
	for p := range Pares(10) {
		soma += p
	}
	fmt.Println("Iterator: ", soma)

	pilha := 0
	c := Comando{Fazer: func() { pilha += 10 }, Desfazer: func() { pilha -= 10 }}
	c.Fazer()
	c.Fazer()
	c.Desfazer()
	fmt.Println("Command:  ", pilha)

	var a Assunto
	var recebidos []string
	a.Assinar(func(m string) { recebidos = append(recebidos, "A:"+m) })
	a.Assinar(func(m string) { recebidos = append(recebidos, "B:"+m) })
	a.Emitir("oi")
	fmt.Println("Observer: ", recebidos)

	rec := httptest.NewRecorder()
	ComCabecalho(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {})).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Println("Decorator:", rec.Header().Get("X-Decorado"))

	for _, t := range []string{"cao", "gato", "boi"} {
		an, err := NovoAnimal(t)
		if err != nil {
			fmt.Println("Factory:  ", err)
			continue
		}
		fmt.Println("Factory:  ", t, "->", an.Som())
	}
}
