// Package migracao importa uma lista de endereços para um servidor em execução.
package migracao

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-5/encurtador/internal/contrato"
)

// Cliente fala com o servidor de links curtos pela API HTTP.
type Cliente struct {
	Base string       // por exemplo, http://localhost:8080
	HTTP *http.Client // se nulo, usa http.DefaultClient
}

// Falha registra uma linha que não pôde ser importada.
type Falha struct {
	Linha int
	URL   string
	Erro  string
}

// Resultado resume uma importação.
type Resultado struct {
	Lidas   int
	Criadas int
	Falhas  []Falha
}

// Importar lê um endereço por linha (linhas vazias e as que começam com # são
// ignoradas) e cria um link curto para cada um. Uma linha que falha é
// registrada e a importação continua; só o cancelamento do contexto interrompe.
func Importar(ctx context.Context, c Cliente, r io.Reader) (Resultado, error) {
	cliente := c.HTTP
	if cliente == nil {
		cliente = http.DefaultClient
	}
	var res Resultado
	linhas := bufio.NewScanner(r)
	for n := 1; linhas.Scan(); n++ {
		url := strings.TrimSpace(linhas.Text())
		if url == "" || strings.HasPrefix(url, "#") {
			continue
		}
		res.Lidas++
		if err := ctx.Err(); err != nil {
			return res, err
		}
		if err := criar(ctx, cliente, c.Base, url); err != nil {
			res.Falhas = append(res.Falhas, Falha{Linha: n, URL: url, Erro: err.Error()})
			continue
		}
		res.Criadas++
	}
	return res, linhas.Err()
}

func criar(ctx context.Context, cliente *http.Client, base, url string) error {
	corpo, err := json.Marshal(contrato.PedidoCriar{URL: url})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+"/links", bytes.NewReader(corpo))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := cliente.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		var e contrato.Erro
		if json.NewDecoder(io.LimitReader(resp.Body, 4<<10)).Decode(&e) == nil && e.Mensagem != "" {
			return fmt.Errorf("status %d: %s", resp.StatusCode, e.Mensagem)
		}
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
