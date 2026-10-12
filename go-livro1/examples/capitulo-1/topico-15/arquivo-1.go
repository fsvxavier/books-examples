package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type statusResposta struct {
	Status string    `json:"status"`
	Hora   time.Time `json:"hora"`
}

func saude(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := statusResposta{Status: "ok", Hora: time.Now().UTC()}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("erro ao escrever resposta: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /saude", saude)

	srv := &http.Server{
		Addr:              ":8089",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("ouvindo em http://localhost:8089")
	log.Fatal(srv.ListenAndServe())
}
