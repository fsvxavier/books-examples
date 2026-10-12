package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/estoura", func(w http.ResponseWriter, r *http.Request) {
		var m map[string]int
		m["x"] = 1 // pânico: gravação em map nil
	})
	mux.HandleFunc("/vivo", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "continuo de pé\n")
	})
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.ErrorLog = log.New(os.Stdout, "[servidor] ", 0)
	srv.Start()
	defer srv.Close()

	_, err := http.Get(srv.URL + "/estoura")
	fmt.Println("cliente, /estoura:", err)
	resp, err := http.Get(srv.URL + "/vivo")
	if err != nil {
		fmt.Println("cliente, /vivo:", err)
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Print("cliente, /vivo: ", string(b))
}
