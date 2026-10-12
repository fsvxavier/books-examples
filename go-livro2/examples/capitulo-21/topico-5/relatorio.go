// Package logs resume um arquivo de log de acesso em estatisticas.
//
// Cada linha tem o formato:
//
//	2026-10-05T12:00:01Z GET /api/produtos/123 200 12.5ms usuario=ana
package logs

// Relatorio e o resultado de uma varredura.
type Relatorio struct {
	Total     int            // linhas validas
	Erros     int            // respostas 5xx
	PorRota   map[string]int // rota normalizada (/api/produtos/:id) -> contagem
	P95Ms     float64        // percentil 95 da latencia, em ms (posto mais proximo)
	MaisAtivo string         // usuario com mais linhas (empate: o menor nome)
}

// maisAtivo escolhe o usuario com mais linhas; em empate, o menor nome.
func maisAtivo(contagem map[string]int) string {
	melhor, n := "", -1
	for u, c := range contagem {
		if c > n || (c == n && u < melhor) {
			melhor, n = u, c
		}
	}
	return melhor
}

// posto devolve o indice do percentil p (0 a 1) pelo metodo do posto mais proximo.
func posto(n int, p float64) int {
	i := int(p*float64(n)+0.999999999) - 1
	if i < 0 {
		i = 0
	}
	if i >= n {
		i = n - 1
	}
	return i
}
