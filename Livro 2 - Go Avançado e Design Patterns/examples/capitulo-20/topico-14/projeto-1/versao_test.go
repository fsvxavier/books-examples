package zlibc

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func texto(n int) []byte {
	return []byte(strings.Repeat("o rato roeu a roupa do rei de roma. ", n/36+1))[:n]
}

// codigoFonte junta arquivos .go da biblioteca padrao: um texto realista.
func codigoFonte(tb testing.TB, n int) []byte {
	tb.Helper()
	arquivos, _ := filepath.Glob(filepath.Join(os.Getenv("GOROOT_FONTES"), "*.go"))
	var buf bytes.Buffer
	for _, a := range arquivos {
		if strings.HasSuffix(a, "_test.go") {
			continue
		}
		b, err := os.ReadFile(a)
		if err != nil {
			tb.Fatal(err)
		}
		buf.Write(b)
		if buf.Len() >= n {
			break
		}
	}
	if buf.Len() < n {
		tb.Skip("defina GOROOT_FONTES com um diretorio de fontes Go com pelo menos 1 MiB")
	}
	return buf.Bytes()[:n]
}

func aleatorio(n int) []byte {
	r := rand.New(rand.NewPCG(7, 9))
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.IntN(256))
	}
	return b
}

func TestIdaEVolta(t *testing.T) {
	for _, n := range []int{0, 1, 100, 100_000} {
		for nome, dados := range map[string][]byte{"texto": texto(n), "aleatorio": aleatorio(n)} {
			t.Run(nome, func(t *testing.T) {
				c, err := Comprimir(dados, Padrao)
				if err != nil {
					t.Fatal(err)
				}
				d, err := Descomprimir(c, len(dados))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(d, dados) {
					t.Fatalf("n=%d: dados diferentes", n)
				}
			})
		}
	}
}

// Interoperabilidade nos dois sentidos com o compress/zlib da biblioteca padrao.
func TestInteroperaComCompressZlib(t *testing.T) {
	dados := texto(50_000)

	c, err := Comprimir(dados, Melhor)
	if err != nil {
		t.Fatal(err)
	}
	r, err := zlib.NewReader(bytes.NewReader(c))
	if err != nil {
		t.Fatal(err)
	}
	lido, err := io.ReadAll(r)
	if err != nil || !bytes.Equal(lido, dados) {
		t.Fatalf("Go nao leu o que o C escreveu: %v", err)
	}

	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(dados)
	w.Close()
	d, err := Descomprimir(buf.Bytes(), len(dados))
	if err != nil || !bytes.Equal(d, dados) {
		t.Fatalf("C nao leu o que o Go escreveu: %v", err)
	}
}

func TestErros(t *testing.T) {
	c, _ := Comprimir(texto(1000), Padrao)
	if _, err := Descomprimir(c, 10); !errors.Is(err, ErrTamanho) {
		t.Errorf("tamanho pequeno: %v", err)
	}
	if _, err := Descomprimir([]byte("isto nao e zlib"), 100); !errors.Is(err, ErrDadosCorrompidos) {
		t.Errorf("lixo: %v", err)
	}
	c[len(c)/2] ^= 0xff
	if _, err := Descomprimir(c, 1000); err == nil {
		t.Error("dados adulterados deveriam falhar")
	}
}

func TestVersao(t *testing.T) { t.Log("zlib", Versao()) }

func BenchmarkComprimir(b *testing.B) {
	dados := codigoFonte(b, 1<<20)
	b.Run("zlib-C", func(b *testing.B) {
		b.SetBytes(int64(len(dados)))
		for b.Loop() {
			if _, err := Comprimir(dados, Padrao); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("compress-zlib", func(b *testing.B) {
		b.SetBytes(int64(len(dados)))
		for b.Loop() {
			var buf bytes.Buffer
			w := zlib.NewWriter(&buf)
			w.Write(dados)
			w.Close()
		}
	})
}

func BenchmarkDescomprimir(b *testing.B) {
	dados := codigoFonte(b, 1<<20)
	c, _ := Comprimir(dados, Padrao)
	b.Run("zlib-C", func(b *testing.B) {
		b.SetBytes(int64(len(dados)))
		for b.Loop() {
			if _, err := Descomprimir(c, len(dados)); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("compress-zlib", func(b *testing.B) {
		b.SetBytes(int64(len(dados)))
		for b.Loop() {
			r, _ := zlib.NewReader(bytes.NewReader(c))
			io.Copy(io.Discard, r)
		}
	})
}
