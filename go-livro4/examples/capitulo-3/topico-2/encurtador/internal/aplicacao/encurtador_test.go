package aplicacao_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/aplicacao"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/dominio"
)

var instante = time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC)

// repoFalso é um Repositorio escrito à mão: um mapa e, se preciso, um erro
// que ele devolve em Salvar. Sem biblioteca de mocks.
type repoFalso struct {
	links        map[string]dominio.Link
	erroAoSalvar error
	salvos       int
}

func novoRepoFalso() *repoFalso {
	return &repoFalso{links: make(map[string]dominio.Link)}
}

func (r *repoFalso) Salvar(_ context.Context, l dominio.Link) error {
	if r.erroAoSalvar != nil {
		return r.erroAoSalvar
	}
	if _, ok := r.links[l.Codigo]; ok {
		return dominio.ErrCodigoEmUso
	}
	r.links[l.Codigo] = l
	r.salvos++
	return nil
}

func (r *repoFalso) Buscar(_ context.Context, codigo string) (dominio.Link, error) {
	l, ok := r.links[codigo]
	if !ok {
		return dominio.Link{}, dominio.ErrNaoEncontrado
	}
	return l, nil
}

func (r *repoFalso) RegistrarAcesso(ctx context.Context, codigo string) (dominio.Link, error) {
	l, err := r.Buscar(ctx, codigo)
	if err != nil {
		return dominio.Link{}, err
	}
	l.Acessos++
	r.links[codigo] = l
	return l, nil
}

// sequencia devolve um Gerador que entrega os códigos em ordem e repete o último.
func sequencia(codigos ...string) aplicacao.Gerador {
	i := 0
	return func() string {
		c := codigos[min(i, len(codigos)-1)]
		i++
		return c
	}
}

func novoServico(repo aplicacao.Repositorio, codigos ...string) *aplicacao.Servico {
	return aplicacao.NovoServico(repo, sequencia(codigos...), func() time.Time { return instante })
}

func TestCriarGuardaOLink(t *testing.T) {
	repo := novoRepoFalso()
	s := novoServico(repo, "abc234")

	l, err := s.Criar(t.Context(), "https://go.dev/doc/")
	if err != nil {
		t.Fatal(err)
	}
	if l.Codigo != "abc234" || !l.Criado.Equal(instante) {
		t.Errorf("link = %+v", l)
	}
	if repo.salvos != 1 {
		t.Errorf("salvos = %d", repo.salvos)
	}
}

func TestCriarRejeitaDestinoInvalidoSemTocarNoRepositorio(t *testing.T) {
	repo := novoRepoFalso()
	s := novoServico(repo, "abc234")

	_, err := s.Criar(t.Context(), "ftp://x.org")
	if !errors.Is(err, dominio.ErrDestinoInvalido) {
		t.Fatalf("erro = %v", err)
	}
	if repo.salvos != 0 {
		t.Errorf("o repositório foi chamado %d vez(es) para um destino inválido", repo.salvos)
	}
}

func TestCriarTentaDeNovoNaColisao(t *testing.T) {
	repo := novoRepoFalso()
	s := novoServico(repo, "aaaaaa", "aaaaaa", "bbbbbb")
	_, _ = s.Criar(t.Context(), "https://a.test")

	l, err := s.Criar(t.Context(), "https://b.test")
	if err != nil || l.Codigo != "bbbbbb" {
		t.Fatalf("link = %+v, erro = %v", l, err)
	}
}

func TestCriarDesisteDepoisDeCincoColisoes(t *testing.T) {
	repo := novoRepoFalso()
	s := novoServico(repo, "aaaaaa")
	_, _ = s.Criar(t.Context(), "https://a.test")

	_, err := s.Criar(t.Context(), "https://b.test")
	if !errors.Is(err, aplicacao.ErrSemCodigoLivre) {
		t.Fatalf("erro = %v", err)
	}
}

func TestCriarEmbrulhaErroDoRepositorio(t *testing.T) {
	falha := errors.New("disco cheio")
	repo := novoRepoFalso()
	repo.erroAoSalvar = falha
	s := novoServico(repo, "abc234")

	_, err := s.Criar(t.Context(), "https://a.test")
	if !errors.Is(err, falha) {
		t.Fatalf("erro = %v, esperado embrulhar %v", err, falha)
	}
}

func TestResolverContaAcessoEConsultarNao(t *testing.T) {
	s := novoServico(novoRepoFalso(), "abc234")
	_, _ = s.Criar(t.Context(), "https://go.dev")

	if l, _ := s.Consultar(t.Context(), "abc234"); l.Acessos != 0 {
		t.Errorf("Consultar contou acesso: %d", l.Acessos)
	}
	if l, _ := s.Resolver(t.Context(), "abc234"); l.Acessos != 1 {
		t.Errorf("Resolver: acessos = %d, esperado 1", l.Acessos)
	}
	if _, err := s.Resolver(t.Context(), "zzzzzz"); !errors.Is(err, dominio.ErrNaoEncontrado) {
		t.Errorf("erro = %v", err)
	}
}

func TestCodigoAleatorioRespeitaODominio(t *testing.T) {
	vistos := make(map[string]bool)
	for range 10000 {
		c := aplicacao.CodigoAleatorio()
		if err := dominio.ValidarCodigo(c); err != nil {
			t.Fatal(err)
		}
		vistos[c] = true
	}
	if len(vistos) < 9990 {
		t.Errorf("só %d códigos distintos em 10000", len(vistos))
	}
}
