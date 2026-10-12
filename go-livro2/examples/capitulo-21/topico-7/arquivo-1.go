package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"

	logs "github.com/fsvxavier/books-examples/go-livro-2/capitulo-21/topico-4"
)

func main() {
	// Servidor da aplicacao: so as rotas de negocio.
	app := http.NewServeMux()
	app.HandleFunc("GET /resumir", func(w http.ResponseWriter, r *http.Request) {
		rel := logs.ResumirV1(logs.Gerar(3000))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rel)
	})

	// Servidor de administracao: pprof, so em localhost, em outra porta.
	admin := http.NewServeMux()
	admin.HandleFunc("/debug/pprof/", pprof.Index)
	admin.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	admin.HandleFunc("/debug/pprof/profile", pprof.Profile)
	admin.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	admin.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		log.Println("admin em 127.0.0.1:6060")
		log.Fatal(http.ListenAndServe("127.0.0.1:6060", admin))
	}()

	log.Println("app em :8081")
	log.Fatal(http.ListenAndServe(":8081", app))
}
