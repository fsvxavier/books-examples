package main

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

const layoutLog = "02/Jan/2006:15:04:05 -0700"

type Acesso struct {
	IP      string
	Quando  time.Time
	Metodo  string
	Caminho string
	Status  int
	Bytes   int
}

var reLog = regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "([A-Z]+) (\S+) [^"]*" (\d{3}) (\d+|-)$`)

// ParseRegexp usa uma expressao regular.
func ParseRegexp(linha string) (Acesso, error) {
	m := reLog.FindStringSubmatch(linha)
	if m == nil {
		return Acesso{}, fmt.Errorf("linha fora do formato: %.40q", linha)
	}
	return montar(m[1], m[2], m[3], m[4], m[5], m[6])
}

// ParseManual usa Cut: mais rapido, mais codigo, mais atencao.
func ParseManual(linha string) (Acesso, error) {
	ip, resto, ok := strings.Cut(linha, " ")
	if !ok {
		return Acesso{}, fmt.Errorf("sem ip")
	}
	_, resto, _ = strings.Cut(resto, "[")
	quando, resto, ok := strings.Cut(resto, "] \"")
	if !ok {
		return Acesso{}, fmt.Errorf("sem data")
	}
	req, resto, ok := strings.Cut(resto, "\" ")
	if !ok {
		return Acesso{}, fmt.Errorf("sem requisicao")
	}
	metodo, caminho, _ := strings.Cut(req, " ")
	caminho, _, _ = strings.Cut(caminho, " ")
	status, bytes, ok := strings.Cut(resto, " ")
	if !ok {
		return Acesso{}, fmt.Errorf("sem status")
	}
	return montar(ip, quando, metodo, caminho, status, bytes)
}

func montar(ip, quando, metodo, caminho, status, bytes string) (Acesso, error) {
	t, err := time.Parse(layoutLog, quando)
	if err != nil {
		return Acesso{}, err
	}
	st, err := strconv.Atoi(status)
	if err != nil {
		return Acesso{}, err
	}
	n := 0
	if bytes != "-" {
		if n, err = strconv.Atoi(bytes); err != nil {
			return Acesso{}, err
		}
	}
	return Acesso{ip, t, metodo, caminho, st, n}, nil
}

func gerar(n int) []string {
	r := rand.New(rand.NewPCG(1, 2))
	base := time.Date(2026, 10, 4, 10, 0, 0, 0, time.FixedZone("", -3*3600))
	caminhos := []string{"/", "/api/pedidos", "/api/pedidos/42", "/login", "/static/app.js"}
	status := []int{200, 200, 200, 200, 201, 304, 404, 500}
	out := make([]string, n)
	for i := range out {
		t := base.Add(time.Duration(i*20) * time.Millisecond) // ~50 req/s
		out[i] = fmt.Sprintf(`203.0.113.%d - - [%s] "GET %s HTTP/1.1" %d %d`,
			r.IntN(250)+1, t.Format(layoutLog), caminhos[r.IntN(len(caminhos))], status[r.IntN(len(status))], r.IntN(5000))
	}
	out = append(out, `lixo que nao e uma linha de log`)
	return out
}

func main() {
	linhas := gerar(60_000)

	rodar := func(nome string, f func(string) (Acesso, error)) []Acesso {
		ini := time.Now()
		var acessos []Acesso
		erros := 0
		for _, l := range linhas {
			a, err := f(l)
			if err != nil {
				erros++
				continue
			}
			acessos = append(acessos, a)
		}
		fmt.Printf("%-8s %d linhas ok, %d com erro, em %s\n", nome, len(acessos), erros, time.Since(ini).Round(time.Millisecond))
		return acessos
	}
	a1 := rodar("regexp", ParseRegexp)
	a2 := rodar("manual", ParseManual)
	fmt.Println("resultados iguais:", slices.Equal(a1, a2))

	// agregacao por minuto, no fuso do proprio log
	porMinuto := map[time.Time]struct{ total, erros int }{}
	porStatus := map[int]int{}
	for _, a := range a1 {
		min := a.Quando.Truncate(time.Minute)
		c := porMinuto[min]
		c.total++
		if a.Status >= 500 {
			c.erros++
		}
		porMinuto[min] = c
		porStatus[a.Status]++
	}
	chaves := make([]time.Time, 0, len(porMinuto))
	for k := range porMinuto {
		chaves = append(chaves, k)
	}
	slices.SortFunc(chaves, func(a, b time.Time) int { return a.Compare(b) })
	fmt.Println("\nminuto               req   5xx")
	for _, k := range chaves[:4] { // so os primeiros minutos, para a saida caber
		c := porMinuto[k]
		fmt.Printf("%s  %4d  %4d (%.1f%%)\n", k.Format("2006-01-02 15:04"), c.total, c.erros, 100*float64(c.erros)/float64(c.total))
	}
	fmt.Printf("... (%d minutos no total)\n", len(chaves))
	codigos := make([]int, 0)
	for c := range porStatus {
		codigos = append(codigos, c)
	}
	sort.Ints(codigos)
	fmt.Print("\nstatus:")
	for _, c := range codigos {
		fmt.Printf(" %d=%d", c, porStatus[c])
	}
	fmt.Println()
}
