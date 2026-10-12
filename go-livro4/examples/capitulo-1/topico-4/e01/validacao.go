package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const tamanhoMaxDestino = 2048

var errDestinoInvalido = errors.New("destino inválido")

// validarDestino confere o endereço recebido e devolve a forma normalizada.
func validarDestino(bruto string) (string, error) {
	bruto = strings.TrimSpace(bruto)
	if bruto == "" {
		return "", fmt.Errorf("%w: vazio", errDestinoInvalido)
	}
	if len(bruto) > tamanhoMaxDestino {
		return "", fmt.Errorf("%w: mais de %d bytes", errDestinoInvalido, tamanhoMaxDestino)
	}
	u, err := url.Parse(bruto)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errDestinoInvalido, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%w: esquema %q não é http nem https", errDestinoInvalido, u.Scheme)
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("%w: sem host", errDestinoInvalido)
	}
	return u.String(), nil
}
