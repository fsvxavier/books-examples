package main

import (
	"net/http"
	"time"
)

// TransporteFunc adapta uma funcao a http.RoundTripper (o mesmo truque do HandlerFunc).
type TransporteFunc func(*http.Request) (*http.Response, error)

func (f TransporteFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Decorador de cliente: recebe um RoundTripper e devolve outro.
type DecoradorTransporte func(http.RoundTripper) http.RoundTripper

// ComCabecalho acrescenta um cabecalho a toda requisicao. O contrato de RoundTripper
// proibe modificar a requisicao recebida, entao trabalhamos em um CLONE.
func ComCabecalho(chave, valor string) DecoradorTransporte {
	return func(prox http.RoundTripper) http.RoundTripper {
		return TransporteFunc(func(r *http.Request) (*http.Response, error) {
			r2 := r.Clone(r.Context())
			r2.Header.Set(chave, valor)
			return prox.RoundTrip(r2)
		})
	}
}

// ComRetentativa repete requisicoes idempotentes que receberam 503. A espera e injetada.
func ComRetentativa(max int, esperar func(time.Duration)) DecoradorTransporte {
	return func(prox http.RoundTripper) http.RoundTripper {
		return TransporteFunc(func(r *http.Request) (*http.Response, error) {
			var (
				resp *http.Response
				err  error
			)
			for i := 0; i < max; i++ {
				resp, err = prox.RoundTrip(r)
				if err != nil || resp.StatusCode != http.StatusServiceUnavailable || r.Method != http.MethodGet {
					return resp, err
				}
				if i < max-1 {
					resp.Body.Close() // descarta a resposta que vamos repetir
					esperar(time.Duration(i+1) * 10 * time.Millisecond)
				}
			}
			return resp, err
		})
	}
}

// Decorar aplica os decoradores com o PRIMEIRO da lista mais por fora.
func Decorar(base http.RoundTripper, ds ...DecoradorTransporte) http.RoundTripper {
	for i := len(ds) - 1; i >= 0; i-- {
		base = ds[i](base)
	}
	return base
}
