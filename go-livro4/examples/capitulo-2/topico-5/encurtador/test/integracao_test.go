// Package test reúne testes que atravessam vários pacotes do módulo.
// Pode importar internal/ porque mora dentro do mesmo módulo.
package test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-5/encurtador/internal/app"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-5/encurtador/internal/armazem"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-5/encurtador/internal/migracao"
)

// TestMigrarContraServidorReal sobe o servidor de verdade, importa uma lista
// pelo mesmo código que o comando migrar usa e confere o resultado seguindo
// os redirecionamentos pela rede.
func TestMigrarContraServidorReal(t *testing.T) {
	ctx, cancelar := context.WithCancel(t.Context())
	defer cancelar()
	leitor, escritor := io.Pipe()
	go func() {
		app.Executar(ctx, []string{"-addr", "127.0.0.1:0"}, func(string) string { return "" }, escritor, io.Discard)
		escritor.Close()
	}()
	linhas := bufio.NewScanner(leitor)
	if !linhas.Scan() {
		t.Fatal("servidor não subiu")
	}
	base := "http://" + strings.TrimPrefix(linhas.Text(), "ouvindo em ")

	res, err := migracao.Importar(ctx, migracao.Cliente{Base: base}, strings.NewReader("https://go.dev/doc/\nhttps://pkg.go.dev/std\n"))
	if err != nil || res.Criadas != 2 {
		t.Fatalf("res=%+v err=%v", res, err)
	}

	// Sem seguir o redirecionamento: queremos ver o 302 e o Location.
	cliente := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := cliente.Post(base+"/links", "application/json", strings.NewReader(`{"url":"https://go.dev/blog/"}`))
	if err != nil {
		t.Fatal(err)
	}
	var criada struct{ Codigo string }
	_ = json.NewDecoder(resp.Body).Decode(&criada)
	resp.Body.Close()

	resp, err = cliente.Get(base + "/" + criada.Codigo)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != "https://go.dev/blog/" {
		t.Fatalf("status %d, Location %q", resp.StatusCode, resp.Header.Get("Location"))
	}

	resp, err = http.Get(base + "/links/" + criada.Codigo)
	if err != nil {
		t.Fatal(err)
	}
	var l armazem.Link
	_ = json.NewDecoder(resp.Body).Decode(&l)
	resp.Body.Close()
	if l.Acessos != 1 {
		t.Errorf("acessos = %d, esperado 1", l.Acessos)
	}
}
