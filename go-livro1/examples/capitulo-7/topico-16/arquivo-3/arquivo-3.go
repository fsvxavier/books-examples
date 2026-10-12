package main

import (
	"errors"
	"fmt"
	"sort"
)

var ErrNaoEncontrado = errors.New("usuário não encontrado")

type Usuario struct {
	ID    int
	Nome  string
	Ativo bool
}

// Repositorio é definido AQUI, junto de quem o consome, e só com o que o serviço precisa.
type Repositorio interface {
	Buscar(id int) (Usuario, error)
	Salvar(u Usuario) error
}

// Servico depende da abstração, e não de um banco de dados.
type Servico struct {
	repo Repositorio
}

func NovoServico(repo Repositorio) *Servico { return &Servico{repo: repo} }

// Desativar busca o usuário, marca como inativo e grava de volta.
func (s *Servico) Desativar(id int) error {
	u, err := s.repo.Buscar(id)
	if err != nil {
		return fmt.Errorf("desativar usuário %d: %w", id, err)
	}
	if !u.Ativo {
		return nil // já estava inativo: operação idempotente
	}
	u.Ativo = false
	return s.repo.Salvar(u)
}

// RepoMemoria é uma implementação real, simples, que serve para testes e protótipos.
type RepoMemoria struct {
	dados map[int]Usuario
}

func NovoRepoMemoria(usuarios ...Usuario) *RepoMemoria {
	r := &RepoMemoria{dados: make(map[int]Usuario)}
	for _, u := range usuarios {
		r.dados[u.ID] = u
	}
	return r
}

func (r *RepoMemoria) Buscar(id int) (Usuario, error) {
	u, ok := r.dados[id]
	if !ok {
		return Usuario{}, ErrNaoEncontrado
	}
	return u, nil
}

func (r *RepoMemoria) Salvar(u Usuario) error {
	r.dados[u.ID] = u
	return nil
}

func (r *RepoMemoria) IDs() []int {
	ids := make([]int, 0, len(r.dados))
	for id := range r.dados {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

func main() {
	repo := NovoRepoMemoria(Usuario{1, "Ana", true}, Usuario{2, "Bia", false})
	s := NovoServico(repo)

	fmt.Println(s.Desativar(1))
	u, _ := repo.Buscar(1)
	fmt.Println(u)
	fmt.Println(s.Desativar(99))
}
