package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	ttemplate "text/template"
	"time"

	htemplate "html/template"
)

func xss() {
	entrada := `<script>alert('roubar cookie')</script>`
	const modelo = `<p>Ola, {{.}}</p>`

	var a, b bytes.Buffer
	ttemplate.Must(ttemplate.New("t").Parse(modelo)).Execute(&a, entrada)
	htemplate.Must(htemplate.New("h").Parse(modelo)).Execute(&b, entrada)
	fmt.Println("text/template:", strings.TrimSpace(a.String()))
	fmt.Println("html/template:", strings.TrimSpace(b.String()))
}

func injecaoDeComando() {
	entrada := "arquivo.txt; echo INJETADO"
	// ERRADO: o shell interpreta o ';'
	out, _ := exec.Command("sh", "-c", "echo processando "+entrada).Output()
	fmt.Printf("sh -c  : %q\n", strings.TrimSpace(string(out)))
	// CERTO: sem shell, a entrada e UM argumento
	out, _ = exec.Command("echo", "processando", entrada).Output()
	fmt.Printf("sem sh : %q\n", strings.TrimSpace(string(out)))
}

func travessia() {
	base, _ := os.MkdirTemp("", "base")
	defer os.RemoveAll(base)
	os.WriteFile(filepath.Join(base, "publico.txt"), []byte("ok"), 0o644)
	segredo := filepath.Join(filepath.Dir(base), "segredo-"+filepath.Base(base)+".txt")
	os.WriteFile(segredo, []byte("SEGREDO"), 0o644)
	defer os.Remove(segredo)

	pedido := "../" + filepath.Base(segredo)
	caminho := filepath.Join(base, pedido) // Join normaliza o "..": escapa da pasta!
	dados, err := os.ReadFile(caminho)
	fmt.Printf("os.ReadFile(Join): %q err=%v\n", dados, err)
	fmt.Println("filepath.IsLocal:", filepath.IsLocal(pedido), "/", filepath.IsLocal("publico.txt"))

	raiz, _ := os.OpenRoot(base)
	defer raiz.Close()
	_, err = raiz.ReadFile(pedido)
	fmt.Println("os.Root.ReadFile:", err)
	d, err := raiz.ReadFile("publico.txt")
	fmt.Printf("os.Root legitimo: %q %v\n", d, err)
}

var ErrDestinoProibido = errors.New("destino proibido")

// dialerSeguro recusa conexao a enderecos internos ANTES de conectar,
// inclusive depois da resolucao de DNS (evita DNS rebinding).
func dialerSeguro() *net.Dialer {
	return &net.Dialer{
		Timeout: 3 * time.Second,
		Control: func(network, address string, _ syscall.RawConn) error {
			ap, err := netip.ParseAddrPort(address)
			if err != nil {
				return err
			}
			ip := ap.Addr().Unmap()
			if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast() {
				return fmt.Errorf("%w: %s", ErrDestinoProibido, ip)
			}
			return nil
		},
	}
}

func ssrf() {
	interno := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "metadados internos")
	}))
	defer interno.Close()

	inseguro := &http.Client{}
	resp, err := inseguro.Get(interno.URL)
	if err == nil {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("cliente comum acessou o interno: %q\n", b)
	}
	seguro := &http.Client{Transport: &http.Transport{DialContext: dialerSeguro().DialContext}}
	_, err = seguro.Get(interno.URL)
	fmt.Println("cliente seguro:", errors.Is(err, ErrDestinoProibido), "-", err)
	_, err = seguro.Get("http://169.254.169.254/latest/meta-data/")
	fmt.Println("metadados da nuvem:", errors.Is(err, ErrDestinoProibido))
}

func cabecalho() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Eco", "valor\r\nSet-Cookie: sessao=roubada")
		w.WriteHeader(200)
	}))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		fmt.Println("injecao de cabecalho: erro", err)
		return
	}
	resp.Body.Close()
	fmt.Printf("X-Eco=%q Set-Cookie presente=%v\n", resp.Header.Get("X-Eco"), resp.Header.Get("Set-Cookie") != "")
}

func redos() {
	re := regexp.MustCompile(`^(a+)+$`)
	entrada := strings.Repeat("a", 50) + "!"
	ini := time.Now()
	ok := re.MatchString(entrada)
	fmt.Printf("regexp (a+)+$ em 51 caracteres: %v em %s (RE2: tempo linear)\n", ok, time.Since(ini).Round(time.Microsecond))
}

func main() {
	xss()
	fmt.Println()
	injecaoDeComando()
	fmt.Println()
	travessia()
	fmt.Println()
	ssrf()
	fmt.Println()
	cabecalho()
	fmt.Println()
	redos()
}
