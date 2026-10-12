// Package cli é um adaptador primário (de entrada) de linha de comando: a
// mesma aplicação, dirigida por uma pessoa no terminal ou por um script,
// sem HTTP no caminho.
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/nucleo"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

const uso = `uso: enc <comando> [argumentos]

comandos:
  criar <url>           encurta o endereço e imprime o código
  consultar <codigo>    mostra o destino e os acessos, sem contar acesso
  seguir <codigo>       imprime o destino e conta um acesso
  importar [arquivo]    cria um link por linha ("-" ou omitido: entrada padrão)`

// Executar interpreta o comando e o executa pela porta de entrada. Recebe a
// porta pronta: a CLI não sabe nem quer saber onde os links ficam guardados.
func Executar(ctx context.Context, args []string, enc nucleo.Encurtador, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(stderr, uso)
		return errors.New("faltou o comando")
	}
	comando, resto := args[0], args[1:]
	switch comando {
	case "criar":
		return criar(ctx, resto, enc, stdout)
	case "consultar":
		return consultar(ctx, resto, enc, stdout, false)
	case "seguir":
		return consultar(ctx, resto, enc, stdout, true)
	case "importar":
		return importar(ctx, resto, enc, stdin, stdout)
	default:
		fmt.Fprintln(stderr, uso)
		return fmt.Errorf("comando desconhecido: %q", comando)
	}
}

func criar(ctx context.Context, args []string, enc nucleo.Encurtador, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("uso: enc criar <url>")
	}
	l, err := enc.Criar(ctx, args[0])
	if err != nil {
		return traduzir(err)
	}
	fmt.Fprintln(stdout, l.Codigo)
	return nil
}

func consultar(ctx context.Context, args []string, enc nucleo.Encurtador, stdout io.Writer, contar bool) error {
	if len(args) != 1 {
		return errors.New("uso: enc consultar|seguir <codigo>")
	}
	busca := enc.Consultar
	if contar {
		busca = enc.Resolver
	}
	l, err := busca(ctx, args[0])
	if err != nil {
		return traduzir(err)
	}
	if contar {
		fmt.Fprintln(stdout, l.Destino)
		return nil
	}
	fmt.Fprintf(stdout, "%s -> %s (%d acessos)\n", l.Codigo, l.Destino, l.Acessos)
	return nil
}

// importar é o ator "script em lote": lê endereços e os entrega à aplicação,
// um por um, sem passar por servidor nenhum. Linhas inválidas são relatadas e
// não interrompem o lote.
func importar(ctx context.Context, args []string, enc nucleo.Encurtador, stdin io.Reader, stdout io.Writer) error {
	entrada := stdin
	if len(args) > 1 {
		return errors.New("uso: enc importar [arquivo]")
	}
	if len(args) == 1 && args[0] != "-" {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()
		entrada = f
	}

	var lidas, criadas, falhas int
	linhas := bufio.NewScanner(entrada)
	for n := 1; linhas.Scan(); n++ {
		url := strings.TrimSpace(linhas.Text())
		if url == "" || strings.HasPrefix(url, "#") {
			continue
		}
		lidas++
		if err := ctx.Err(); err != nil {
			return err
		}
		l, err := enc.Criar(ctx, url)
		if err != nil {
			falhas++
			fmt.Fprintf(stdout, "linha %d (%s): %v\n", n, url, traduzir(err))
			continue
		}
		criadas++
		fmt.Fprintf(stdout, "%s %s\n", l.Codigo, l.Destino)
	}
	fmt.Fprintf(stdout, "lidas=%d criadas=%d falhas=%d\n", lidas, criadas, falhas)
	if err := linhas.Err(); err != nil {
		return err
	}
	if falhas > 0 {
		return errors.New("algumas linhas falharam")
	}
	return nil
}

// traduzir é o equivalente, na CLI, do que o adaptador HTTP faz com os status:
// erros do negócio viram mensagens para a pessoa; o resto é devolvido como veio.
func traduzir(err error) error {
	switch {
	case errors.Is(err, dominio.ErrNaoEncontrado):
		return errors.New("link não encontrado")
	case errors.Is(err, nucleo.ErrSemCodigoLivre):
		return errors.New("não consegui gerar um código livre, tente de novo")
	default:
		return err
	}
}
