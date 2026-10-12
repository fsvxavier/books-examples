package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// semUnwrap embute a interface http.ResponseWriter: só os 3 métodos dela são promovidos.
type semUnwrap struct {
	http.ResponseWriter
}

// comUnwrap acrescenta Unwrap, que o http.ResponseController usa para achar o original.
type comUnwrap struct {
	http.ResponseWriter
}

func (c comUnwrap) Unwrap() http.ResponseWriter { return c.ResponseWriter }

func eventos(w http.ResponseWriter, r *http.Request) {
	// type assertion antiga: w.(http.Flusher)
	_, ok := w.(http.Flusher)
	fmt.Printf("  assertion w.(http.Flusher): %v\n", ok)

	rc := http.NewResponseController(w)
	err := rc.Flush()
	switch {
	case err == nil:
		fmt.Println("  ResponseController.Flush: ok")
	case errors.Is(err, http.ErrNotSupported):
		fmt.Println("  ResponseController.Flush: ErrNotSupported")
	default:
		fmt.Println("  ResponseController.Flush: erro", err)
	}
}

func main() {
	h := http.HandlerFunc(eventos)
	req := httptest.NewRequest("GET", "/eventos", nil)

	fmt.Println("sem middleware:")
	h.ServeHTTP(httptest.NewRecorder(), req)

	fmt.Println("wrapper sem Unwrap:")
	h.ServeHTTP(semUnwrap{httptest.NewRecorder()}, req)

	fmt.Println("wrapper com Unwrap:")
	h.ServeHTTP(comUnwrap{httptest.NewRecorder()}, req)
}
