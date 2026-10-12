package logs

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// Gerar produz n linhas de log de acesso com um gerador pseudoaleatório de
// semente fixa: os mesmos dados em toda execução. Cerca de 1% das linhas são
// inválidas (lixo), para exercitar o caminho de erro do analisador.
func Gerar(n int) []string {
	r := rand.New(rand.NewPCG(42, 1024))
	rotas := []string{
		"/api/produtos/%d",
		"/api/pedidos/%d",
		"/api/pedidos/%d/itens/%d",
		"/api/clientes/%d",
		"/api/clientes/%d/enderecos",
		"/saude",
	}
	metodos := []string{"GET", "GET", "GET", "POST", "PUT", "DELETE"}
	usuarios := []string{"ana", "bruno", "carla", "davi", "edu", "fabi", "gil", "hana"}
	status := []int{200, 200, 200, 200, 201, 204, 301, 400, 404, 500}
	linhas := make([]string, 0, n)
	for i := range n {
		if r.IntN(100) == 0 {
			linhas = append(linhas, "lixo")
			continue
		}
		rota := rotas[r.IntN(len(rotas))]
		ids := []any{r.IntN(100000), r.IntN(1000)}
		caminho := fmt.Sprintf(rota, ids[:strings.Count(rota, "%d")]...)
		ms := 5 + r.Float64()*float64(20+r.IntN(180))
		linhas = append(linhas, fmt.Sprintf("2026-10-05T12:%02d:%02dZ %s %s %d %.1fms usuario=%s",
			(i/60)%60, i%60, metodos[r.IntN(len(metodos))], caminho,
			status[r.IntN(len(status))], ms, usuarios[r.IntN(len(usuarios))]))
	}
	return linhas
}
