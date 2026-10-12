package main

import (
	"errors"
	"testing"
)

// repoFalho é um dublê de teste: devolve o erro que o teste pedir.
type repoFalho struct {
	erroSalvar error
	salvou     bool
}

func (r *repoFalho) Buscar(id int) (Usuario, error) { return Usuario{ID: id, Ativo: true}, nil }
func (r *repoFalho) Salvar(u Usuario) error {
	r.salvou = true
	return r.erroSalvar
}

// repoContador reaproveita RepoMemoria por embedding e só acrescenta a contagem de gravações.
type repoContador struct {
	*RepoMemoria
	gravacoes int
}

func (r *repoContador) Salvar(u Usuario) error {
	r.gravacoes++
	return r.RepoMemoria.Salvar(u)
}

func TestDesativar(t *testing.T) {
	t.Run("desativa usuário ativo", func(t *testing.T) {
		repo := NovoRepoMemoria(Usuario{1, "Ana", true})
		if err := NovoServico(repo).Desativar(1); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		u, _ := repo.Buscar(1)
		if u.Ativo {
			t.Error("usuário deveria estar inativo")
		}
	})

	t.Run("usuário inexistente", func(t *testing.T) {
		err := NovoServico(NovoRepoMemoria()).Desativar(42)
		if !errors.Is(err, ErrNaoEncontrado) {
			t.Errorf("quer ErrNaoEncontrado, veio %v", err)
		}
	})

	t.Run("propaga falha do repositório", func(t *testing.T) {
		falha := errors.New("disco cheio")
		repo := &repoFalho{erroSalvar: falha}
		err := NovoServico(repo).Desativar(7)
		if !errors.Is(err, falha) || !repo.salvou {
			t.Errorf("erro = %v, salvou = %t", err, repo.salvou)
		}
	})

	t.Run("já inativo não grava", func(t *testing.T) {
		repo := &repoContador{RepoMemoria: NovoRepoMemoria(Usuario{2, "Bia", false})}
		if err := NovoServico(repo).Desativar(2); err != nil {
			t.Fatal(err)
		}
		if repo.gravacoes != 0 {
			t.Errorf("gravou %d vezes, queria 0", repo.gravacoes)
		}
	})
}
