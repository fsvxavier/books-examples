// Package registro oferece um catalogo generico de itens nomeados.
package registro

import (
	"fmt"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
)

// Registro guarda itens de tipo T por nome. Ele tem duas fases:
// registro (escrita permitida) e uso (somente leitura, depois de Congelar).
type Registro[T any] struct {
	nome      string
	mu        sync.Mutex
	itens     map[string]T
	congelado atomic.Pointer[map[string]T]
}

// Novo cria um registro vazio. O nome so aparece nas mensagens de erro.
func Novo[T any](nome string) *Registro[T] {
	return &Registro[T]{nome: nome, itens: map[string]T{}}
}

// Registrar adiciona um item. Panica em nome vazio, nome repetido ou
// registro ja congelado: tudo isso e erro de programacao.
func (r *Registro[T]) Registrar(nome string, item T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.congelado.Load() != nil {
		panic(fmt.Sprintf("%s: Registrar(%q) depois de Congelar", r.nome, nome))
	}
	if nome == "" {
		panic(r.nome + ": nome vazio")
	}
	if _, existe := r.itens[nome]; existe {
		panic(fmt.Sprintf("%s: %q ja registrado", r.nome, nome))
	}
	r.itens[nome] = item
}

// Congelar encerra a fase de registro. Daqui em diante Obter nao usa trava.
func (r *Registro[T]) Congelar() {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := maps.Clone(r.itens)
	r.congelado.Store(&c)
}

// Obter devolve o item e se ele existe.
func (r *Registro[T]) Obter(nome string) (T, bool) {
	if c := r.congelado.Load(); c != nil {
		v, ok := (*c)[nome]
		return v, ok
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.itens[nome]
	return v, ok
}

// Nomes devolve os nomes em ordem alfabetica.
func (r *Registro[T]) Nomes() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Sorted(maps.Keys(r.itens))
}
