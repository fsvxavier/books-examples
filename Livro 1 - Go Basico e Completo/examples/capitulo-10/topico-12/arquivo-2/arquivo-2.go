package main

import (
	"fmt"
	"slices"
	"time"
)

// site simula a web: cada página aponta para outras (com ciclos, de propósito).
var site = map[string][]string{
	"/":           {"/produtos", "/sobre", "/contato"},
	"/produtos":   {"/produtos/1", "/produtos/2", "/"},
	"/produtos/1": {"/produtos", "/produtos/2"},
	"/produtos/2": {"/produtos/1", "/carrinho"},
	"/sobre":      {"/", "/equipe"},
	"/equipe":     {"/sobre"},
	"/contato":    {"/"},
	"/carrinho":   {"/checkout"},
	"/checkout":   {},
}

// buscar simula uma requisição HTTP: demora um pouco e devolve os links da página.
func buscar(url string) []string {
	time.Sleep(10 * time.Millisecond)
	return site[url]
}

type achado struct {
	de    string
	links []string
}

// Rastrear visita todas as páginas alcançáveis a partir de inicio, usando `workers` goroutines.
// O mapa "visitadas" pertence SÓ ao coordenador (este laço): sem mutex, sem corrida.
func Rastrear(inicio string, workers int) []string {
	urls := make(chan string)
	achados := make(chan achado)

	for range workers {
		go func() {
			for u := range urls {
				achados <- achado{u, buscar(u)}
			}
		}()
	}

	visitadas := map[string]bool{inicio: true}
	fila := []string{inicio} // páginas descobertas, ainda não enviadas aos workers
	emAndamento := 0         // páginas enviadas e ainda sem resposta

	for len(fila) > 0 || emAndamento > 0 {
		// só tentamos enviar se houver algo na fila; senão, o canal "envio" fica nil e o case é ignorado
		var envio chan string
		var proxima string
		if len(fila) > 0 {
			envio = urls
			proxima = fila[0]
		}

		select {
		case envio <- proxima:
			fila = fila[1:]
			emAndamento++
		case a := <-achados:
			emAndamento--
			for _, l := range a.links {
				if !visitadas[l] {
					visitadas[l] = true
					fila = append(fila, l)
				}
			}
		}
	}
	close(urls) // não há mais trabalho: os workers terminam

	resultado := make([]string, 0, len(visitadas))
	for u := range visitadas {
		resultado = append(resultado, u)
	}
	slices.Sort(resultado)
	return resultado
}

func main() {
	inicio := time.Now()
	paginas := Rastrear("/", 1)
	um := time.Since(inicio)

	inicio = time.Now()
	paginas = Rastrear("/", 4)
	quatro := time.Since(inicio)

	fmt.Println(len(paginas), paginas)
	fmt.Println("4 workers são mais rápidos que 1:", quatro < um)
}
