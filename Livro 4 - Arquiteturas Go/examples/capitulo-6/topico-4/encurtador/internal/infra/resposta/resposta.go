// Package resposta escreve as respostas JSON que todas as fatias compartilham.
package resposta

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/contrato"
)

// JSON escreve v como corpo JSON com o status dado.
func JSON(w http.ResponseWriter, log *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error("escrevendo resposta", "erro", err)
	}
}

// Erro escreve a resposta de erro no formato do contrato.
func Erro(w http.ResponseWriter, log *slog.Logger, status int, msg string) {
	JSON(w, log, status, contrato.Erro{Mensagem: msg})
}
