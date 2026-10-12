package main

import (
	"fmt"
	"os"
	"runtime/trace"
	"sync"
	"time"
)

func trabalhar() {
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			s := 0
			for i := range 20_000_000 {
				s += i
			}
			_ = s
		})
	}
	wg.Wait()
}

func main() {
	// Janela móvel com os eventos mais recentes: custo baixo, sempre ligada.
	fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{MinAge: 2 * time.Second, MaxBytes: 4 << 20})
	if err := fr.Start(); err != nil {
		fmt.Println("erro:", err)
		return
	}
	defer fr.Stop()

	inicio := time.Now()
	trabalhar()
	if d := time.Since(inicio); d > 5*time.Millisecond { // "requisição lenta": salva a janela
		f, err := os.Create("lento.trace")
		if err != nil {
			fmt.Println("erro:", err)
			return
		}
		defer f.Close()
		n, err := fr.WriteTo(f)
		fmt.Printf("operação levou %v; gravei %d bytes de trace em lento.trace (err=%v)\n", d.Round(time.Millisecond), n, err)
	}
}
