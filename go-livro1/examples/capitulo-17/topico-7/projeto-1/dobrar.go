package fx

import (
	"sort"
	"strings"
	"sync"
)

func Dobrar(v interface{}) interface{} { return v }

func Somar(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		total += i
	}
	return total
}

func Tem(xs []string, alvo string) bool {
	for _, x := range xs {
		if x == alvo {
			return true
		}
	}
	return false
}

func Ordenar(xs []int) {
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
}

func Prefixo(s string) (string, bool) {
	if strings.HasPrefix(s, "id:") {
		return strings.TrimPrefix(s, "id:"), true
	}
	return s, false
}

func Linhas(doc string) int {
	n := 0
	for _, l := range strings.Split(doc, "\n") {
		if l != "" {
			n++
		}
	}
	return n
}

func Paralelo(fs []func()) {
	var wg sync.WaitGroup
	for _, f := range fs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f()
		}()
	}
	wg.Wait()
}

func Copiar(m map[string]int) map[string]int {
	out := make(map[string]int)
	for k, v := range m {
		out[k] = v
	}
	return out
}
