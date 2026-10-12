package main

import (
	"errors"
	"fmt"
	"time"
)

var ErrTemporario = errors.New("falha temporária")

// consultarServico simula um serviço instável: falha nas duas primeiras chamadas.
func consultarServico(chamada int) (string, error) {
	if chamada < 3 {
		return "", fmt.Errorf("chamada %d: %w", chamada, ErrTemporario)
	}
	return "resposta do serviço", nil
}

// ComRetentativas chama a função até maxTentativas vezes, dobrando a espera a cada falha.
func ComRetentativas(maxTentativas int, espera time.Duration, f func(tentativa int) (string, error)) (string, error) {
	var ultimoErro error

	for tentativa := 1; tentativa <= maxTentativas; tentativa++ {
		resultado, err := f(tentativa)
		if err == nil {
			return resultado, nil
		}

		ultimoErro = err
		if !errors.Is(err, ErrTemporario) {
			return "", err // erro definitivo, não adianta insistir
		}

		if tentativa == maxTentativas {
			break
		}

		fmt.Printf("tentativa %d falhou (%v); esperando %v\n", tentativa, err, espera)
		time.Sleep(espera)
		espera *= 2
	}

	return "", fmt.Errorf("desisti após %d tentativas: %w", maxTentativas, ultimoErro)
}

func main() {
	res, err := ComRetentativas(5, 10*time.Millisecond, consultarServico)
	fmt.Println(res, err)

	_, err = ComRetentativas(2, time.Millisecond, consultarServico)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrTemporario))
}
