// Package proc e um registro de processadores de pagamento no estilo de database/sql:
// cada implementacao se registra pelo nome, e o cliente pede pelo nome.
package proc

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

type Processador interface {
	Cobrar(ctx context.Context, centavos int64) (string, error)
}

// Fabrica constroi um Processador a partir de uma configuracao (mapa simples).
type Fabrica func(cfg map[string]string) (Processador, error)

var (
	mu       sync.RWMutex
	fabricas = map[string]Fabrica{}
)

var ErrDesconhecido = errors.New("processador desconhecido")

// Registrar normalmente e chamado por init() do pacote da implementacao.
// Como sql.Register, entra em panico para nome vazio, fabrica nula ou duplicado:
// sao erros de programacao, descobertos na partida.
func Registrar(nome string, f Fabrica) {
	mu.Lock()
	defer mu.Unlock()
	if nome == "" || f == nil {
		panic("proc: nome vazio ou fabrica nula")
	}
	if _, dup := fabricas[nome]; dup {
		panic("proc: Registrar chamado duas vezes para " + nome)
	}
	fabricas[nome] = f
}

// Novo e o factory method que o resto do programa usa.
func Novo(nome string, cfg map[string]string) (Processador, error) {
	mu.RLock()
	f, ok := fabricas[nome]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %q (registrados: %v)", ErrDesconhecido, nome, Nomes())
	}
	return f(cfg)
}

func Nomes() []string {
	mu.RLock()
	defer mu.RUnlock()
	n := make([]string, 0, len(fabricas))
	for k := range fabricas {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}
