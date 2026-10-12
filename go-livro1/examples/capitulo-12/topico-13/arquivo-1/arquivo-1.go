package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Servico simula uma chamada remota que respeita o contexto.
type Servico struct {
	Nome     string
	Latencia time.Duration
	Falha    error
}

func (s Servico) Consultar(ctx context.Context) (string, error) {
	select {
	case <-time.After(s.Latencia):
		if s.Falha != nil {
			return "", fmt.Errorf("%s: %w", s.Nome, s.Falha)
		}
		return "dados de " + s.Nome, nil
	case <-ctx.Done():
		return "", fmt.Errorf("%s: %w", s.Nome, context.Cause(ctx))
	}
}

var ErrIndisponivel = errors.New("serviço indisponível")

// Agregar consulta todos os serviços em paralelo. Regras:
//   - o prazo TOTAL é `prazo`;
//   - se um serviço obrigatório falhar, cancela os demais imediatamente (fail-fast).
func Agregar(ctx context.Context, prazo time.Duration, servicos []Servico) (map[string]string, error) {
	ctx, cancelarPrazo := context.WithTimeoutCause(ctx, prazo, fmt.Errorf("prazo de %v esgotado", prazo))
	defer cancelarPrazo()
	ctx, cancelar := context.WithCancelCause(ctx)
	defer cancelar(nil)

	var (
		mu         sync.Mutex
		resultados = make(map[string]string)
		primeiro   error
		wg         sync.WaitGroup
	)
	for _, s := range servicos {
		wg.Go(func() {
			dados, err := s.Consultar(ctx)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if primeiro == nil {
					primeiro = err
					cancelar(err) // avisa os irmãos: não adianta continuar
				}
				return
			}
			resultados[s.Nome] = dados
		})
	}
	wg.Wait()
	return resultados, primeiro
}

func nomes(m map[string]string) []string {
	var ns []string
	for k := range m {
		ns = append(ns, k)
	}
	sort.Strings(ns)
	return ns
}

func main() {
	ok := []Servico{
		{"perfil", 30 * time.Millisecond, nil},
		{"pedidos", 50 * time.Millisecond, nil},
		{"estoque", 20 * time.Millisecond, nil},
	}
	r, err := Agregar(context.Background(), time.Second, ok)
	fmt.Println(nomes(r), err)

	comFalha := []Servico{
		{"perfil", 30 * time.Millisecond, nil},
		{"pedidos", 400 * time.Millisecond, nil}, // lento: será cancelado pela falha do estoque
		{"estoque", 20 * time.Millisecond, ErrIndisponivel},
	}
	inicio := time.Now()
	r, err = Agregar(context.Background(), time.Second, comFalha)
	fmt.Println(nomes(r), err, "| errors.Is:", errors.Is(err, ErrIndisponivel), "| rápido:", time.Since(inicio) < 200*time.Millisecond)

	lentos := []Servico{{"perfil", 300 * time.Millisecond, nil}}
	r, err = Agregar(context.Background(), 50*time.Millisecond, lentos)
	fmt.Println(nomes(r), err)
}
