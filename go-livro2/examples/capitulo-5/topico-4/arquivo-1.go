package main

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"sort"
	"strings"
)

func main() {
	// 1) sort.Reverse adapta uma sort.Interface para a ordem inversa.
	nums := []int{3, 1, 2}
	sort.Sort(sort.Reverse(sort.IntSlice(nums)))
	fmt.Println("sort.Reverse(IntSlice):", nums)

	// 2) http.HandlerFunc adapta uma FUNCAO para a interface http.Handler.
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ola de uma funcao")
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Println("http.HandlerFunc:", rec.Body.String())

	// 3) strings.NewReader adapta uma string para io.Reader (e io.NopCloser, para io.ReadCloser).
	rc := io.NopCloser(strings.NewReader("linha 1\nlinha 2\nlinha 3"))
	defer rc.Close()
	// 4) bufio.Scanner adapta um io.Reader para um fluxo de linhas.
	sc := bufio.NewScanner(rc)
	n := 0
	for sc.Scan() {
		n++
	}
	fmt.Println("io.NopCloser + bufio.Scanner: linhas =", n)

	// 5) slices.Values e maps.Keys adaptam colecoes para iteradores (iter.Seq).
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	fmt.Println("slices.Sorted(maps.Keys(m)):", slices.Sorted(maps.Keys(m)))
	soma := 0
	for v := range slices.Values([]int{10, 20, 30}) {
		soma += v
	}
	fmt.Println("range slices.Values:", soma)

	// 6) io.Pipe adapta escrita para leitura: quem escreve nao conhece quem le.
	pr, pw := io.Pipe()
	go func() {
		fmt.Fprint(pw, "dados via pipe")
		pw.Close()
	}()
	b, _ := io.ReadAll(pr)
	fmt.Println("io.Pipe:", string(b))
}
