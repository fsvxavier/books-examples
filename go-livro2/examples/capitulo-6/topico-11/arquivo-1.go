package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
)

// contadorW e um decorador de io.Writer: conta o que passa.
type contadorW struct {
	w io.Writer
	n int
}

func (c *contadorW) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += n
	return n, err
}

func main() {
	texto := strings.Repeat("linha de log com bastante repeticao\n", 2000)

	// ESCRITA: bufio -> contador(entrada) -> gzip -> contador(saida) -> buffer
	var destino bytes.Buffer
	saida := &contadorW{w: &destino}
	zw := gzip.NewWriter(saida)
	entrada := &contadorW{w: zw}
	bw := bufio.NewWriterSize(entrada, 4096)

	for _, linha := range strings.SplitAfter(texto, "\n") {
		bw.WriteString(linha)
	}
	bw.Flush() // 1) esvazia o bufio para o gzip
	zw.Close() // 2) fecha o gzip: escreve o fim do fluxo (a ordem importa)

	fmt.Printf("escritas pequenas: %d linhas\n", strings.Count(texto, "\n"))
	fmt.Printf("bytes que entraram no gzip: %d\n", entrada.n)
	fmt.Printf("bytes gravados no destino:  %d (%.1f%% do original)\n", saida.n, 100*float64(saida.n)/float64(entrada.n))

	// LEITURA: buffer -> gzip -> TeeReader(sha256) -> LimitReader -> leitura
	zr, _ := gzip.NewReader(&destino)
	h := sha256.New()
	tee := io.TeeReader(zr, h) // tudo que e lido tambem alimenta o hash
	limitado := io.LimitReader(tee, 1<<20)
	lido, _ := io.ReadAll(limitado)

	fmt.Printf("bytes lidos de volta: %d, iguais ao original: %v\n", len(lido), string(lido) == texto)
	esperado := sha256.Sum256([]byte(texto))
	fmt.Printf("sha256 calculado durante a leitura confere: %v\n", bytes.Equal(h.Sum(nil), esperado[:]))

	// Sem Flush/Close, o que e gravado fica incompleto.
	var d2 bytes.Buffer
	zw2 := gzip.NewWriter(&d2)
	zw2.Write([]byte(texto))
	fmt.Printf("sem Close: %d bytes no destino; ", d2.Len())
	_, err := io.ReadAll(mustReader(&d2))
	fmt.Println("leitura:", err)
}

func mustReader(r io.Reader) io.Reader {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return strings.NewReader("")
	}
	return zr
}
