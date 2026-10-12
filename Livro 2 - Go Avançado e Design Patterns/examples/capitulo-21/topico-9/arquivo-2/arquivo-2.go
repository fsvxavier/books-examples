package main

import (
	"fmt"
	"os"
	"runtime/trace"
	"sync"
	"time"
)

var (
	mu   sync.Mutex
	fila []int
)

// processar demora de proposito a cada 50 chamadas: a "requisicao lenta".
func processar(i int) time.Duration {
	inicio := time.Now()
	mu.Lock()
	defer mu.Unlock()
	espera := 200 * time.Microsecond
	if i%50 == 49 {
		espera = 40 * time.Millisecond
	}
	time.Sleep(espera)
	fila = append(fila, i)
	return time.Since(inicio)
}

func main() {
	fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
		MinAge:   2 * time.Second,
		MaxBytes: 4 << 20,
	})
	if err := fr.Start(); err != nil {
		panic(err)
	}
	defer fr.Stop()

	var wg sync.WaitGroup
	var once sync.Once
	for g := 0; g < 4; g++ {
		wg.Go(func() {
			for i := 0; i < 200; i++ {
				if d := processar(i); d > 30*time.Millisecond {
					// o gatilho: so quando algo anormal acontece, grava a janela
					once.Do(func() {
						f, _ := os.Create("lenta.trace")
						n, err := fr.WriteTo(f)
						f.Close()
						fmt.Printf("requisicao lenta (%v): %d bytes de trace gravados (erro: %v)\n",
							d.Round(time.Millisecond), n, err)
					})
				}
			}
		})
	}
	wg.Wait()
	fmt.Println("processadas:", len(fila))
}
