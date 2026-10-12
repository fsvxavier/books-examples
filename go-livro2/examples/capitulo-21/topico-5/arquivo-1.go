package logs

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
)

var (
	reLinha  = regexp.MustCompile(`^(\S+) (\S+) (\S+) (\d{3}) ([\d.]+)ms usuario=(\S+)$`)
	reNumero = regexp.MustCompile(`/\d+`)
)

// ResumirV1 so tira as expressoes regulares de dentro do laco.
func ResumirV1(linhas []string) Relatorio {
	r := Relatorio{PorRota: map[string]int{}}
	var latencias []float64
	usuarios := map[string]int{}

	for _, linha := range linhas {
		m := reLinha.FindStringSubmatch(linha)
		if m == nil {
			continue
		}
		r.Total++

		status, _ := strconv.Atoi(m[4])
		if status >= 500 {
			r.Erros++
		}

		rota := reNumero.ReplaceAllString(m[3], "/:id")
		chave := fmt.Sprintf("%s %s", m[2], rota)
		r.PorRota[chave] = r.PorRota[chave] + 1

		ms, _ := strconv.ParseFloat(m[5], 64)
		latencias = append(latencias, ms)

		usuarios[m[6]] = usuarios[m[6]] + 1
	}

	sort.Slice(latencias, func(i, j int) bool { return latencias[i] < latencias[j] })
	if len(latencias) > 0 {
		r.P95Ms = latencias[posto(len(latencias), 0.95)]
	}
	r.MaisAtivo = maisAtivo(usuarios)
	return r
}
