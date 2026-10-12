package cli_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/adaptadores/entrada/cli"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/nucleo"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

// aplicacaoFalsa é a porta de entrada nas mãos do teste.
type aplicacaoFalsa struct {
	criados []string
	erro    error
}

func (a *aplicacaoFalsa) Criar(_ context.Context, destino string) (dominio.Link, error) {
	if a.erro != nil {
		return dominio.Link{}, a.erro
	}
	a.criados = append(a.criados, destino)
	return dominio.Link{Codigo: fmt.Sprintf("cod%03d", len(a.criados)), Destino: destino}, nil
}

func (a *aplicacaoFalsa) Resolver(_ context.Context, codigo string) (dominio.Link, error) {
	if a.erro != nil {
		return dominio.Link{}, a.erro
	}
	return dominio.Link{Codigo: codigo, Destino: "https://go.dev", Acessos: 1}, nil
}

func (a *aplicacaoFalsa) Consultar(_ context.Context, codigo string) (dominio.Link, error) {
	if a.erro != nil {
		return dominio.Link{}, a.erro
	}
	return dominio.Link{Codigo: codigo, Destino: "https://go.dev", Acessos: 7}, nil
}

var _ nucleo.Encurtador = (*aplicacaoFalsa)(nil)

func executar(enc nucleo.Encurtador, stdin string, args ...string) (saida, erros string, err error) {
	var out, errOut strings.Builder
	err = cli.Executar(context.Background(), args, enc, strings.NewReader(stdin), &out, &errOut)
	return out.String(), errOut.String(), err
}

func TestComandos(t *testing.T) {
	casos := []struct {
		nome     string
		args     []string
		esperado string
	}{
		{"criar", []string{"criar", "https://go.dev/doc/"}, "cod001\n"},
		{"consultar", []string{"consultar", "abc234"}, "abc234 -> https://go.dev (7 acessos)\n"},
		{"seguir", []string{"seguir", "abc234"}, "https://go.dev\n"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			saida, _, err := executar(&aplicacaoFalsa{}, "", c.args...)
			if err != nil || saida != c.esperado {
				t.Errorf("saída %q, erro %v; esperado %q", saida, err, c.esperado)
			}
		})
	}
}

func TestImportarRelataERecomeca(t *testing.T) {
	app := &aplicacaoFalsa{}
	entrada := "# comentário\nhttps://go.dev/a\n\nhttps://go.dev/b\n"
	saida, _, err := executar(app, entrada, "importar")
	if err != nil {
		t.Fatal(err)
	}
	esperado := "cod001 https://go.dev/a\ncod002 https://go.dev/b\nlidas=2 criadas=2 falhas=0\n"
	if saida != esperado {
		t.Errorf("saída:\n%s\nesperada:\n%s", saida, esperado)
	}
}

func TestImportarComFalha(t *testing.T) {
	app := &aplicacaoFalsa{erro: fmt.Errorf("%w: sem host", dominio.ErrDestinoInvalido)}
	saida, _, err := executar(app, "http:///x\n", "importar", "-")
	if err == nil || !strings.Contains(saida, "linha 1 (http:///x)") || !strings.Contains(saida, "falhas=1") {
		t.Fatalf("saída %q, erro %v", saida, err)
	}
}

func TestErrosDeNegocioViramMensagens(t *testing.T) {
	_, _, err := executar(&aplicacaoFalsa{erro: dominio.ErrNaoEncontrado}, "", "seguir", "zzzzzz")
	if err == nil || err.Error() != "link não encontrado" {
		t.Errorf("erro = %v", err)
	}
	_, _, err = executar(&aplicacaoFalsa{erro: nucleo.ErrSemCodigoLivre}, "", "criar", "https://a.test")
	if err == nil || !strings.Contains(err.Error(), "tente de novo") {
		t.Errorf("erro = %v", err)
	}
	outro := errors.New("o banco caiu")
	_, _, err = executar(&aplicacaoFalsa{erro: outro}, "", "criar", "https://a.test")
	if !errors.Is(err, outro) {
		t.Errorf("erro desconhecido deveria passar adiante: %v", err)
	}
}

func TestUsoIncorreto(t *testing.T) {
	_, uso, err := executar(&aplicacaoFalsa{}, "")
	if err == nil || !strings.Contains(uso, "uso: enc") {
		t.Errorf("sem comando: erro %v, uso %q", err, uso)
	}
	if _, _, err := executar(&aplicacaoFalsa{}, "", "apagar"); err == nil {
		t.Error("comando desconhecido deveria falhar")
	}
	if _, _, err := executar(&aplicacaoFalsa{}, "", "criar"); err == nil {
		t.Error("criar sem argumento deveria falhar")
	}
}
