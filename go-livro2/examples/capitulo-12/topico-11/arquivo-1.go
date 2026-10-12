package main

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"unicode"
)

type Pessoa struct {
	Nome  string
	Idade int
}

func ordenacao() {
	ps := []Pessoa{{"Bia", 30}, {"Ana", 25}, {"Caio", 30}, {"Davi", 25}}

	// A estratégia de comparação é uma função passada ao algoritmo.
	slices.SortFunc(ps, func(a, b Pessoa) int { return cmp.Compare(a.Nome, b.Nome) })
	fmt.Println("por nome:      ", ps)

	slices.SortFunc(ps, func(a, b Pessoa) int {
		return cmp.Or(cmp.Compare(b.Idade, a.Idade), cmp.Compare(a.Nome, b.Nome))
	})
	fmt.Println("idade desc, nome:", ps)

	// Mesma estrutura, outro algoritmo de ordenação: SortStableFunc.
	slices.SortStableFunc(ps, func(a, b Pessoa) int { return cmp.Compare(a.Idade, b.Idade) })
	fmt.Println("estável por idade:", ps)
}

func divisores() {
	texto := "um,dois;três quatro"

	for _, estrategia := range []struct {
		nome  string
		split bufio.SplitFunc
	}{
		{"palavras", bufio.ScanWords},
		{"linhas", bufio.ScanLines},
		{"virgula-ou-ponto-e-virgula", func(dados []byte, fim bool) (int, []byte, error) {
			if i := bytes.IndexAny(dados, ",;"); i >= 0 {
				return i + 1, dados[:i], nil
			}
			if fim && len(dados) > 0 {
				return len(dados), dados, nil
			}
			return 0, nil, nil
		}},
	} {
		sc := bufio.NewScanner(strings.NewReader(texto))
		sc.Split(estrategia.split) // troca a estratégia do mesmo Scanner
		var itens []string
		for sc.Scan() {
			itens = append(itens, sc.Text())
		}
		fmt.Printf("%-28s %q\n", estrategia.nome, itens)
	}

	fmt.Printf("FieldsFunc com critério próprio: %q\n",
		strings.FieldsFunc(texto, func(r rune) bool { return !unicode.IsLetter(r) }))
}

func redirecionamento() {
	var passos []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/a":
			http.Redirect(w, r, "/b", http.StatusFound)
		case "/b":
			http.Redirect(w, r, "/c", http.StatusFound)
		default:
			fmt.Fprint(w, "fim")
		}
	}))
	defer srv.Close()

	// Estratégia padrão: segue até 10 redirecionamentos.
	r, err := http.Get(srv.URL + "/a")
	fmt.Println("padrão: status", r.StatusCode, err)
	r.Body.Close()

	// Estratégia própria: não segue nenhum.
	c := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	r, err = c.Get(srv.URL + "/a")
	fmt.Println("sem seguir: status", r.StatusCode, "Location:", r.Header.Get("Location"), err)
	r.Body.Close()

	// Estratégia própria: segue no máximo 1.
	c = &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		passos = append(passos, req.URL.Path)
		if len(via) >= 1 {
			return errors.New("redirecionamentos demais")
		}
		return nil
	}}
	_, err = c.Get(srv.URL + "/a")
	fmt.Println("no máximo 1:", err, passos)
}

func main() {
	fmt.Println("== ordenação ==")
	ordenacao()
	fmt.Println("== divisão de texto ==")
	divisores()
	fmt.Println("== redirecionamento ==")
	redirecionamento()
}
