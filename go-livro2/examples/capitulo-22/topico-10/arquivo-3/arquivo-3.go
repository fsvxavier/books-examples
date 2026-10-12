// Comando motor serve a regra de preço por HTTP e expõe o pprof em outra porta.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-22/topico-10/arquivo-2"
)

type entrada struct {
	Itens    int     `json:"itens"`
	Subtotal float64 `json:"subtotal"`
	Cliente  int     `json:"cliente"`
	Cupom    bool    `json:"cupom"`
}

type saida struct {
	Preco float64 `json:"preco"`
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "endereço do serviço")
	debug := flag.String("debug", "127.0.0.1:6060", "endereço do pprof")
	flag.Parse()

	regra := regras.Preco()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		log.Println(http.ListenAndServe(*debug, mux))
	}()

	http.HandleFunc("/preco", func(w http.ResponseWriter, r *http.Request) {
		var e entrada
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		p := regras.Pedido{Itens: e.Itens, Subtotal: e.Subtotal, Cliente: regras.Nivel(e.Cliente), Cupom: e.Cupom}
		v := regras.Avaliar(regra, &p)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(saida{Preco: v})
	})
	http.HandleFunc("/lote", func(w http.ResponseWriter, r *http.Request) {
		n, err := strconv.Atoi(r.URL.Query().Get("n"))
		if err != nil || n < 1 || n > 100000 {
			http.Error(w, "n deve ficar entre 1 e 100000", http.StatusBadRequest)
			return
		}
		simula := regras.Simular
		if r.URL.Query().Get("reuso") == "1" {
			simula = regras.SimularReuso
		}
		v := simula(regra, n, uint64(n))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(saida{Preco: v})
	})
	log.Fatal(http.ListenAndServe(*addr, nil))
}
