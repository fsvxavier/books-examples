package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
)

// Erros do domínio: a camada de serviço não sabe nada de HTTP.
var (
	ErrNaoEncontrado     = errors.New("recurso não encontrado")
	ErrSaldoInsuficiente = errors.New("saldo insuficiente")
	ErrNaoAutorizado     = errors.New("não autorizado")
)

// ErroValidacao carrega o campo que falhou.
type ErroValidacao struct {
	Campo  string
	Motivo string
}

func (e *ErroValidacao) Error() string { return e.Campo + ": " + e.Motivo }

// statusDe traduz um erro de domínio em código HTTP. É a ÚNICA função que conhece essa ligação.
func statusDe(err error) int {
	var ev *ErroValidacao
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrNaoEncontrado):
		return http.StatusNotFound
	case errors.Is(err, ErrNaoAutorizado):
		return http.StatusForbidden
	case errors.As(err, &ev):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ErrSaldoInsuficiente):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// mensagemPublica evita vazar detalhes internos: erros desconhecidos viram uma mensagem genérica.
func mensagemPublica(err error) string {
	if statusDe(err) == http.StatusInternalServerError {
		return "erro interno"
	}
	return err.Error()
}

type servico struct{}

func (servico) sacar(conta string, valor int) error {
	switch {
	case conta == "":
		return &ErroValidacao{"conta", "é obrigatória"}
	case conta == "fantasma":
		return fmt.Errorf("sacar da conta %q: %w", conta, ErrNaoEncontrado)
	case conta == "alheia":
		return ErrNaoAutorizado
	case valor > 100:
		return fmt.Errorf("sacar %d: %w", valor, ErrSaldoInsuficiente)
	case conta == "quebrada":
		return errors.New("conexão com o banco falhou: senha=segredo123")
	}
	return nil
}

func handlerSaque(s servico) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valor, err := strconv.Atoi(r.URL.Query().Get("valor"))
		if err != nil {
			err = &ErroValidacao{"valor", "precisa ser um número inteiro"}
		} else {
			err = s.sacar(r.URL.Query().Get("conta"), valor)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusDe(err))
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"erro": mensagemPublica(err)})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func main() {
	h := handlerSaque(servico{})
	for _, url := range []string{"/s?conta=ana&valor=150", "/s?conta=&valor=10", "/s?conta=fantasma&valor=10", "/s?conta=alheia&valor=10", "/s?conta=quebrada&valor=10", "/s?conta=ana&valor=abc", "/s?conta=ana&valor=10"} {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest(http.MethodGet, url, nil))
		fmt.Printf("%-30s %d %s", url, rec.Code, rec.Body.String())
	}
}
