package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"
)

// ADAPTER: muda a interface. Entra um io.Reader, sai um iter.Seq[string] (outro "formato" de uso).
func Linhas(r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			if !yield(sc.Text()) {
				return
			}
		}
	}
}

// DECORATOR: MANTEM a interface io.Reader e acrescenta comportamento (contar bytes).
type Contador struct {
	io.Reader
	Bytes int
}

func (c *Contador) Read(p []byte) (int, error) {
	n, err := c.Reader.Read(p)
	c.Bytes += n
	return n, err
}

// PROXY: MANTEM a interface io.Reader e CONTROLA o acesso ao objeto real.
var ErrNegado = errors.New("acesso negado")

type Protegido struct {
	abrir     func() io.Reader // so abre o recurso real se for preciso
	pode      bool
	real      io.Reader
	Aberturas int
}

func (p *Protegido) Read(b []byte) (int, error) {
	if !p.pode {
		return 0, ErrNegado
	}
	if p.real == nil {
		p.real = p.abrir()
		p.Aberturas++
	}
	return p.real.Read(b)
}

func main() {
	const texto = "um\ndois\ntres\n"

	fmt.Println("Adapter: io.Reader vira iter.Seq[string]")
	for l := range Linhas(strings.NewReader(texto)) {
		fmt.Printf("  linha %q\n", l)
	}

	fmt.Println("Decorator: continua sendo io.Reader, mas conta")
	c := &Contador{Reader: strings.NewReader(texto)}
	io.Copy(io.Discard, c)
	fmt.Println("  bytes lidos:", c.Bytes)

	fmt.Println("Proxy: continua sendo io.Reader, mas controla o acesso")
	nega := &Protegido{abrir: func() io.Reader { return strings.NewReader(texto) }, pode: false}
	_, err := io.ReadAll(nega)
	fmt.Println("  sem permissao:", err, "| aberturas:", nega.Aberturas)
	libera := &Protegido{abrir: func() io.Reader { return strings.NewReader(texto) }, pode: true}
	b, _ := io.ReadAll(libera)
	fmt.Printf("  com permissao: %d bytes | aberturas: %d\n", len(b), libera.Aberturas)

	fmt.Println("Combinados: proxy -> decorator -> adapter")
	outro := &Protegido{abrir: func() io.Reader { return strings.NewReader(texto) }, pode: true}
	pc := &Contador{Reader: outro}
	n := 0
	for range Linhas(pc) {
		n++
	}
	fmt.Println("  linhas:", n, "| bytes:", pc.Bytes)
}
