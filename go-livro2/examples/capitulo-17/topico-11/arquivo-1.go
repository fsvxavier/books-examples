// Command e10 compara duas formas de um repositório genérico obter o identificador da entidade.
package main

import (
	"errors"
	"fmt"
)

// RepoRuim extrai o ID com uma asserção de tipo, escondida dentro do método.
type RepoRuim[K comparable, V any] struct{ dados map[K]*V }

// Criar grava se o valor tiver ObterID() K; se não tiver, não faz nada e devolve sucesso.
func (r *RepoRuim[K, V]) Criar(v *V) error {
	if c, ok := any(v).(interface{ ObterID() K }); ok {
		r.dados[c.ObterID()] = v
	}
	return nil
}

// ComIdentificador é a exigência escrita na constraint.
type ComIdentificador[K comparable] interface{ ObterID() K }

// RepoBom exige o identificador na constraint de V.
type RepoBom[K comparable, V ComIdentificador[K]] struct{ dados map[K]V }

// Criar grava o valor pelo seu identificador.
func (r *RepoBom[K, V]) Criar(v V) error {
	r.dados[v.ObterID()] = v
	return nil
}

// Buscar devolve o valor ou um erro.
func (r *RepoBom[K, V]) Buscar(id K) (V, error) {
	v, ok := r.dados[id]
	if !ok {
		var zero V
		return zero, errors.New("não encontrado")
	}
	return v, nil
}

// Usuario implementa ObterID com receptor de valor.
type Usuario struct{ ID, Nome string }

func (u Usuario) ObterID() string { return u.ID }

// Produto esqueceu de implementar ObterID.
type Produto struct{ SKU string }

func main() {
	// Ruim, com o tipo errado: nada acusa, nada é gravado.
	ruim := &RepoRuim[string, Produto]{dados: map[string]*Produto{}}
	err := ruim.Criar(&Produto{SKU: "x1"})
	fmt.Printf("RepoRuim[Produto]: erro = %v, gravados = %d\n", err, len(ruim.dados))

	// Ruim, com o tipo "certo" mas receptor de valor: *Usuario TEM o método (o conjunto de
	// métodos de *T inclui o de T), então aqui funciona.
	ruim2 := &RepoRuim[string, Usuario]{dados: map[string]*Usuario{}}
	_ = ruim2.Criar(&Usuario{ID: "u1"})
	fmt.Printf("RepoRuim[Usuario]: gravados = %d\n", len(ruim2.dados))

	// Bom: o mesmo erro vira erro de compilação (veja o experimento de erros).
	bom := &RepoBom[string, Usuario]{dados: map[string]Usuario{}}
	_ = bom.Criar(Usuario{ID: "u1", Nome: "Ana"})
	u, err := bom.Buscar("u1")
	fmt.Printf("RepoBom[Usuario]: %+v, erro = %v\n", u, err)
}
