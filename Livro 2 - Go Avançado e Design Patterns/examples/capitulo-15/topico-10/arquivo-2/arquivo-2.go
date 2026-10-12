package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"time"
)

func goroutines() int {
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	return runtime.NumGoroutine()
}

func rodar(nome string, tratar func(*http.Response)) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "resposta com algum corpo")
	}))
	cli := &http.Client{Transport: &http.Transport{}}
	antes := goroutines()
	for range 100 {
		resp, err := cli.Get(srv.URL)
		if err != nil {
			fmt.Println(nome, "erro:", err)
			return
		}
		tratar(resp)
	}
	depois := goroutines()
	fmt.Printf("%-34s goroutines: %3d -> %3d\n", nome, antes, depois)
	cli.CloseIdleConnections()
	srv.Close()
}

func main() {
	switch os.Args[1] {
	case "ignorado":
		rodar("corpo ignorado (sem Close)", func(r *http.Response) {})
	case "close":
		rodar("Close sem ler o corpo", func(r *http.Response) { r.Body.Close() })
	case "readall":
		rodar("ReadAll + Close", func(r *http.Response) {
			io.ReadAll(r.Body)
			r.Body.Close()
		})
	}
}
