package logs

import (
	"slices"
	"strconv"
	"strings"
)

// appendChave monta "METODO /rota/:id" em um buffer reaproveitavel, sem
// criar nenhuma string intermediaria.
func appendChave(b []byte, metodo, caminho string) []byte {
	b = append(b, metodo...)
	b = append(b, ' ')
	for {
		seg, resto, achou := strings.Cut(caminho, "/")
		if numerico(seg) {
			b = append(b, ":id"...)
		} else {
			b = append(b, seg...)
		}
		if !achou {
			return b
		}
		b = append(b, '/')
		caminho = resto
	}
}

// ResumirV3 prealoca o que tem tamanho conhecido e nao aloca para contar.
func ResumirV3(linhas []string) Relatorio {
	r := Relatorio{PorRota: make(map[string]int, 64)}
	latencias := make([]float64, 0, len(linhas))
	usuarios := make(map[string]int, 16)
	var chave []byte

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
		r.PorRota[string(chave)]++ // o compilador nao aloca quando a chave ja existe
		latencias = append(latencias, v)
		usuarios[usuario]++
	}

	slices.Sort(latencias)
	if len(latencias) > 0 {
		r.P95Ms = latencias[posto(len(latencias), 0.95)]
	}
	r.MaisAtivo = maisAtivo(usuarios)
	return r
}
