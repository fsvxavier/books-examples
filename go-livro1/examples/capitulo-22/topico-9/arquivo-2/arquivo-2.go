package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"time"
)

// Seguranca acrescenta cabecalhos defensivos a toda resposta.
func Seguranca(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		h.Set("Permissions-Policy", "geolocation=(), camera=()")
		next.ServeHTTP(w, r)
	})
}

// redirecionar so aceita destinos locais (evita open redirect).
func destinoSeguro(alvo string) string {
	if strings.HasPrefix(alvo, "/") && !strings.HasPrefix(alvo, "//") && !strings.Contains(alvo, `\`) {
		return alvo
	}
	return "/"
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /entrar", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "sessao",
			Value:    "token-opaco",
			Path:     "/",
			Expires:  time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			HttpOnly: true,                 // JavaScript nao le
			Secure:   true,                 // so trafega por HTTPS
			SameSite: http.SameSiteLaxMode, // nao vai em POST de outro site
		})
		http.Redirect(w, r, destinoSeguro(r.URL.Query().Get("volta")), http.StatusSeeOther)
	})
	srv := httptest.NewServer(Seguranca(mux))
	defer srv.Close()

	cli := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	for _, volta := range []string{"/painel", "https://malicioso.example/", "//malicioso.example/"} {
		resp, err := cli.Get(srv.URL + "/entrar?volta=" + volta)
		if err != nil {
			panic(err)
		}
		resp.Body.Close()
		fmt.Printf("volta=%-28s -> %d Location: %s\n", volta, resp.StatusCode, resp.Header.Get("Location"))
		if volta == "/painel" {
			chaves := make([]string, 0)
			for k := range resp.Header {
				chaves = append(chaves, k)
			}
			sort.Strings(chaves)
			for _, k := range chaves {
				if k == "Date" || k == "Content-Length" || k == "Content-Type" || k == "Location" {
					continue
				}
				fmt.Printf("    %s: %s\n", k, resp.Header.Get(k))
			}
		}
	}
}
