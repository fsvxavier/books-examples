package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestRunDeFimAFim sobe o programa inteiro numa porta livre, usa-o por HTTP
// de verdade e o desliga cancelando o contexto.
func TestRunDeFimAFim(t *testing.T) {
	ctx, cancelar := context.WithCancel(t.Context())
	leitor, escritor := io.Pipe()
	terminou := make(chan error, 1)
	go func() {
		env := func(string) string { return "" }
		args := []string{"-addr", "127.0.0.1:0", "-base", "http://curto.test"}
		terminou <- run(ctx, args, env, escritor, io.Discard)
		escritor.Close()
	}()

	linhas := bufio.NewScanner(leitor)
	if !linhas.Scan() {
		t.Fatal("o programa não disse onde está ouvindo")
	}
	endereco := strings.TrimPrefix(linhas.Text(), "ouvindo em ")

	resp, err := http.Post("http://"+endereco+"/links", "application/json", strings.NewReader(`{"url":"https://go.dev"}`))
	if err != nil {
		t.Fatal(err)
	}
	var r respostaCriar
	json.NewDecoder(resp.Body).Decode(&r)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated || !strings.HasPrefix(r.Curto, "http://curto.test/") {
		t.Fatalf("status %d, resposta %+v", resp.StatusCode, r)
	}

	cancelar()
	// io.Pipe não tem buffer: o run só termina depois que alguém lê a última
	// linha. Por isso a leitura vem antes de esperar o término.
	ultima := ""
	if linhas.Scan() {
		ultima = linhas.Text()
	}
	select {
	case err := <-terminou:
		if err != nil {
			t.Fatalf("run devolveu %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("o desligamento demorou demais")
	}
	if ultima != "desligado" {
		t.Errorf("faltou a mensagem final: %q", ultima)
	}
}

func TestRunConfiguracaoInvalida(t *testing.T) {
	env := func(string) string { return "" }
	err := run(t.Context(), []string{"-espera", "0s"}, env, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "espera deve ser positiva") {
		t.Fatalf("erro = %v", err)
	}
}

func TestConfigAmbienteEFlag(t *testing.T) {
	env := func(k string) string {
		return map[string]string{"ENCURTADOR_ADDR": ":9000", "ENCURTADOR_BASE": "https://enc.test"}[k]
	}
	cfg, err := lerConfig([]string{"-addr", ":9100"}, env, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.endereco != ":9100" || cfg.baseURL != "https://enc.test" {
		t.Errorf("cfg = %+v (a flag vence o ambiente; o ambiente vence o padrão)", cfg)
	}
}
