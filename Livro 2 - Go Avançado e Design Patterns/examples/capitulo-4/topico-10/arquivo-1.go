package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"os"
	"time"
)

func main() {
	// 1) log e slog compartilham o padrao global.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(g []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(g) == 0 {
				return slog.Attr{}
			}
			return a
		},
	})))
	log.Print("mensagem do pacote log antigo")
	slog.Info("mensagem do slog", "chave", "valor")

	// 2) http.DefaultClient e um ponteiro global: mudar o Timeout muda para todos.
	fmt.Println("DefaultClient.Timeout antes:", http.DefaultClient.Timeout)
	http.DefaultClient.Timeout = 1 * time.Millisecond

	lento := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		fmt.Fprint(w, "ok")
	}))
	defer lento.Close()

	// "outra biblioteca" que usa http.Get sem saber que alguem mexeu no global.
	_, err := http.Get(lento.URL)
	fmt.Println("http.Get de terceiros:", err != nil)
	if err != nil {
		fmt.Println("erro:", err)
	}
	http.DefaultClient.Timeout = 0

	// 3) importar net/http/pprof registra rotas no DefaultServeMux global.
	srv := httptest.NewServer(http.DefaultServeMux)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/debug/pprof/cmdline")
	if err != nil {
		fmt.Println(err)
		return
	}
	resp.Body.Close()
	fmt.Println("GET /debug/pprof/cmdline no DefaultServeMux:", resp.Status)
}
