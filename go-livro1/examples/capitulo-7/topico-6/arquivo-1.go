package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// maiusculas é um io.Reader que transforma outro Reader, convertendo para maiúsculas
// apenas as letras ASCII. Bytes de caracteres multibyte (acentos) passam intactos,
// porque tratar um byte de UTF-8 como se fosse um caractere corrompe o texto.
type maiusculas struct{ origem io.Reader }

func (m maiusculas) Read(p []byte) (int, error) {
	n, err := m.origem.Read(p)
	for i := range n {
		if 'a' <= p[i] && p[i] <= 'z' {
			p[i] = p[i] - 'a' + 'A'
		}
	}
	return n, err
}

// contador é um io.Writer que conta bytes e linhas e repassa tudo ao destino.
type contador struct {
	destino io.Writer
	bytes   int
	linhas  int
}

func (c *contador) Write(p []byte) (int, error) {
	n, err := c.destino.Write(p)
	c.bytes += n
	c.linhas += bytes.Count(p[:n], []byte("\n"))
	return n, err
}

func main() {
	r := strings.NewReader("olá, mundo\nsegunda linha\n")
	c := &contador{destino: os.Stdout}
	n, err := io.Copy(c, maiusculas{r})
	fmt.Println(n, err, c.bytes, c.linhas)

	var buf bytes.Buffer
	multi := io.MultiWriter(&buf, os.Stdout)
	fmt.Fprintln(multi, "para dois destinos")
	fmt.Println(buf.Len())

	var espelho strings.Builder
	tee := io.TeeReader(strings.NewReader("lido e copiado"), &espelho)
	dados, _ := io.ReadAll(tee)
	fmt.Println(string(dados), "|", espelho.String())

	limitado := io.LimitReader(strings.NewReader("0123456789"), 4)
	d2, _ := io.ReadAll(limitado)
	fmt.Println(string(d2))

	sc := bufio.NewScanner(strings.NewReader("a\nb\nc"))
	for sc.Scan() {
		fmt.Print(sc.Text(), ";")
	}
	fmt.Println()
}
