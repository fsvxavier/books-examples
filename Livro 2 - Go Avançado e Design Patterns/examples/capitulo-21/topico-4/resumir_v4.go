package logs

import (
	"slices"
	"strconv"
)

// ResumirV4 conta com map[string]*int: a busca com string(chave) nao aloca,
// e so a primeira vez que uma rota aparece a chave vira uma string de verdade.
func ResumirV4(linhas []string) Relatorio {
	rotas := make(map[string]*int, 64)
	latencias := make([]float64, 0, len(linhas))
	usuarios := make(map[string]int, 16)
	var chave []byte
	r := Relatorio{}

	for _, linha := range linhas {
		metodo, caminho, status, ms, usuario, ok := campos(linha)
		if !ok {
			continue
		}
		s, err := strconv.Atoi(status)
		if err != nil || len(status) != 3 {
			continue
		}
		v, err := strconv.ParseFloat(ms, 64)
		if err != nil {
			continue
		}
		r.Total++
		if s >= 500 {
			r.Erros++
		}
		chave = appendChave(chave[:0], metodo, caminho)
		if n, ok := rotas[string(chave)]; ok { // leitura: sem alocacao
			*n++
		} else {
			n := 1
			rotas[string(chave)] = &n // insercao: aloca, mas so uma vez por rota
		}
		latencias = append(latencias, v)
		usuarios[usuario]++
	}

	r.PorRota = make(map[string]int, len(rotas))
	for k, n := range rotas {
		r.PorRota[k] = *n
	}
	slices.Sort(latencias)
	if len(latencias) > 0 {
		r.P95Ms = latencias[posto(len(latencias), 0.95)]
	}
	r.MaisAtivo = maisAtivo(usuarios)
	return r
}
