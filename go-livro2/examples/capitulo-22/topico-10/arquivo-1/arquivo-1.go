// Comando carga gera tráfego contra o motor e informa a vazão.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://127.0.0.1:8080/preco", "endereço do serviço")
	dur := flag.Duration("d", 10*time.Second, "duração do teste")
	conc := flag.Int("c", 8, "conexões concorrentes")
	flag.Parse()

	var total atomic.Int64
	var wg sync.WaitGroup
	fim := time.Now().Add(*dur)
	for w := 0; w < *conc; w++ {
		wg.Add(1)
		go func(semente uint64) {
			defer wg.Done()
			r := rand.New(rand.NewPCG(semente, 7))
			cli := &http.Client{}
			for time.Now().Before(fim) {
				corpo := fmt.Sprintf(`{"itens":%d,"subtotal":%.2f,"cliente":%d,"cupom":%t}`,
					1+r.IntN(20), r.Float64()*3000, r.IntN(3), r.IntN(4) == 0)
				resp, err := cli.Post(*url, "application/json", bytes.NewReader([]byte(corpo)))
				if err != nil {
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				total.Add(1)
			}
		}(uint64(w + 1))
	}
	wg.Wait()
	n := total.Load()
	fmt.Printf("requisições: %d  vazão: %.0f req/s\n", n, float64(n)/dur.Seconds())
}
