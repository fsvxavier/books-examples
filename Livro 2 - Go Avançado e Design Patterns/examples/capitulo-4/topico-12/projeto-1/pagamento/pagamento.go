// Package pagamento define o contrato e o registro dos meios de pagamento.
package pagamento

import (
	"fmt"
	"slices"
	"sync"
)

// Meio e o que toda forma de pagamento sabe fazer.
type Meio interface {
	Cobrar(centavos int64) (string, error)
}

// Fabrica cria um Meio novo. Guardamos fabricas, nao instancias,
// para que cada uso possa ter o seu proprio estado.
type Fabrica func() Meio

var (
	mu       sync.RWMutex
	fabricas = map[string]Fabrica{}
)

// Registrar associa um nome a uma fabrica. Panica se o nome for vazio,
// a fabrica for nula ou o nome ja existir: registrar errado e erro de
// programacao e deve aparecer na partida, nao em producao.
func Registrar(nome string, f Fabrica) {
	if nome == "" || f == nil {
		panic("pagamento: nome vazio ou fabrica nula")
	}
	mu.Lock()
	defer mu.Unlock()
	if _, existe := fabricas[nome]; existe {
		panic("pagamento: Registrar chamado duas vezes para " + nome)
	}
	fabricas[nome] = f
}

// Novo cria um Meio pelo nome.
func Novo(nome string) (Meio, error) {
	mu.RLock()
	f, ok := fabricas[nome]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("pagamento: meio %q desconhecido (esqueceu o import _ do pacote?)", nome)
	}
	return f(), nil
}

// Nomes lista os meios registrados, em ordem alfabetica.
func Nomes() []string {
	mu.RLock()
	defer mu.RUnlock()
	nomes := make([]string, 0, len(fabricas))
	for n := range fabricas {
		nomes = append(nomes, n)
	}
	slices.Sort(nomes)
	return nomes
}
