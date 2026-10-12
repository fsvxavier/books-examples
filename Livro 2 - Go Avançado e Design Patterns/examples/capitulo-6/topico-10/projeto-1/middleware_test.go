package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOrdemEntradaESaida(t *testing.T) {
	d := &Diario{}
	h := Encadear(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { d.Add("handler") }),
		Rastro("A", d), Rastro("B", d))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if got := strings.Join(d.Linhas, ","); got != "-> A,-> B,handler,<- B,<- A" {
		t.Fatalf("ordem = %s", got)
	}
}

func TestAutenticarCurtoCircuito(t *testing.T) {
	chamou := false
	h := Encadear(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { chamou = true }), Autenticar("x"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 401 || chamou {
		t.Fatalf("code=%d chamou=%v", rec.Code, chamou)
	}
}

func TestRecuperar(t *testing.T) {
	d := &Diario{}
	h := Encadear(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("x") }), Recuperar(d))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 500 || len(d.Linhas) != 1 {
		t.Fatalf("code=%d diario=%v", rec.Code, d.Linhas)
	}
}

func TestCabecalhoNaoMutaOriginal(t *testing.T) {
	var recebido *http.Request
	base := TransporteFunc(func(r *http.Request) (*http.Response, error) {
		recebido = r
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	})
	original, _ := http.NewRequest("GET", "http://x", nil)
	Decorar(base, ComCabecalho("X-A", "1")).RoundTrip(original)
	if original.Header.Get("X-A") != "" || recebido.Header.Get("X-A") != "1" {
		t.Fatal("o decorador alterou a requisicao original ou nao enviou o cabecalho")
	}
}

func TestRetentativaSoEmGET(t *testing.T) {
	for metodo, quero := range map[string]int{"GET": 3, "POST": 1} {
		n := 0
		base := TransporteFunc(func(r *http.Request) (*http.Response, error) {
			n++
			return &http.Response{StatusCode: 503, Body: http.NoBody}, nil
		})
		rt := Decorar(base, ComRetentativa(3, func(time.Duration) {}))
		req, _ := http.NewRequest(metodo, "http://x", nil)
		resp, _ := rt.RoundTrip(req)
		if n != quero || resp.StatusCode != 503 {
			t.Errorf("%s: tentativas = %d, quero %d", metodo, n, quero)
		}
	}
}

func TestConcorrente(t *testing.T) {
	d := &Diario{}
	h := Encadear(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), Registrar(d), Rastro("A", d))
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() { h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)) })
	}
	wg.Wait()
	if len(d.Linhas) != 150 {
		t.Fatalf("linhas = %d", len(d.Linhas))
	}
}
