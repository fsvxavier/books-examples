package main

import (
	"errors"
	"fmt"
	"time"
)

type resposta struct {
	replica string
	valor   string
	err     error
}

// consultar simula uma réplica com latência variável.
func consultar(nome string, latencia time.Duration, falha bool) resposta {
	time.Sleep(latencia)
	if falha {
		return resposta{nome, "", fmt.Errorf("réplica %s indisponível", nome)}
	}
	return resposta{nome, "dados da " + nome, nil}
}

// primeiraResposta dispara todas as consultas e devolve a PRIMEIRA que tiver sucesso.
// O canal tem buffer para as goroutines retardatárias nunca ficarem presas ao enviar.
func primeiraResposta(timeout time.Duration, consultas ...func() resposta) (resposta, error) {
	ch := make(chan resposta, len(consultas))
	for _, q := range consultas {
		go func() { ch <- q() }()
	}

	var erros []error
	limite := time.After(timeout)
	for range consultas {
		select {
		case r := <-ch:
			if r.err == nil {
				return r, nil
			}
			erros = append(erros, r.err)
		case <-limite:
			return resposta{}, fmt.Errorf("timeout após %v: %w", timeout, errors.Join(erros...))
		}
	}
	return resposta{}, fmt.Errorf("todas as réplicas falharam: %w", errors.Join(erros...))
}

func main() {
	r, err := primeiraResposta(time.Second,
		func() resposta { return consultar("A", 80*time.Millisecond, false) },
		func() resposta { return consultar("B", 20*time.Millisecond, true) },
		func() resposta { return consultar("C", 40*time.Millisecond, false) },
	)
	fmt.Println(r.valor, err)

	_, err = primeiraResposta(100*time.Millisecond,
		func() resposta { return consultar("A", 500*time.Millisecond, false) },
		func() resposta { return consultar("B", 10*time.Millisecond, true) },
	)
	fmt.Println(err)

	_, err = primeiraResposta(time.Second,
		func() resposta { return consultar("A", 10*time.Millisecond, true) },
		func() resposta { return consultar("B", 20*time.Millisecond, true) },
	)
	fmt.Println(err)
}
