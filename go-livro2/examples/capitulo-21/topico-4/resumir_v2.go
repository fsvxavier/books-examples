package logs

import (
	"slices"
	"strconv"
	"strings"
)

// campos separa a linha sem alocar: cada campo e um pedaco da propria linha.
func campos(linha string) (metodo, caminho, status, ms, usuario string, ok bool) {
	var resto string
	if _, resto, ok = strings.Cut(linha, " "); !ok { // pula o carimbo de tempo
		return
	}
	if metodo, resto, ok = strings.Cut(resto, " "); !ok {
		return
	}
	if caminho, resto, ok = strings.Cut(resto, " "); !ok {
		return
	}
	if status, resto, ok = strings.Cut(resto, " "); !ok {
		return
	}
	if ms, resto, ok = strings.Cut(resto, " "); !ok {
		return
	}
	ms, ok = strings.CutSuffix(ms, "ms")
	if !ok {
		return
	}
	usuario, ok = strings.CutPrefix(resto, "usuario=")
	return
}

// normalizar troca os segmentos numericos de um caminho por :id.
func normalizar(caminho string) string {
	var sb strings.Builder
	sb.Grow(len(caminho) + 8)
	primeiro := true
	for seg := range strings.SplitSeq(caminho, "/") {
		if !primeiro {
			sb.WriteByte('/')
		}
		primeiro = false
		if numerico(seg) {
			sb.WriteString(":id")
		} else {
			sb.WriteString(seg)
		}
	}
	return sb.String()
}

func numerico(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// ResumirV2 troca as expressoes regulares por cortes de string.
func ResumirV2(linhas []string) Relatorio {
	r := Relatorio{PorRota: map[string]int{}}
	var latencias []float64
	usuarios := map[string]int{}

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
		r.PorRota[metodo+" "+normalizar(caminho)]++
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
