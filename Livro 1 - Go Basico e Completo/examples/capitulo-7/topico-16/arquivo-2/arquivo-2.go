package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

// Filtrar copia de r para w só as linhas que passam no filtro, e devolve quantas foram gravadas.
// Funciona com qualquer origem e destino: arquivo, rede, buffer, entrada padrão...
func Filtrar(r io.Reader, w io.Writer, manter func(linha string) bool) (int, error) {
	escritor := bufio.NewWriter(w)
	leitor := bufio.NewScanner(r)
	gravadas := 0
	for leitor.Scan() {
		linha := leitor.Text()
		if !manter(linha) {
			continue
		}
		if _, err := fmt.Fprintln(escritor, linha); err != nil {
			return gravadas, err
		}
		gravadas++
	}
	if err := leitor.Err(); err != nil {
		return gravadas, err
	}
	return gravadas, escritor.Flush()
}

const logExemplo = `INFO  iniciando servidor
ERRO  falha ao conectar no banco
INFO  conectado
ERRO  timeout na requisição /pedidos
INFO  encerrando`

func main() {
	apenasErros := func(l string) bool { return strings.HasPrefix(l, "ERRO") }

	// 1. string -> stdout
	n, err := Filtrar(strings.NewReader(logExemplo), os.Stdout, apenasErros)
	fmt.Println(n, err)

	// 2. string -> buffer em memória (ideal para testes)
	var buf bytes.Buffer
	Filtrar(strings.NewReader(logExemplo), &buf, apenasErros)
	fmt.Println(buf.Len(), "bytes")

	// 3. string -> gzip -> buffer, e de volta
	var comprimido bytes.Buffer
	gz := gzip.NewWriter(&comprimido)
	Filtrar(strings.NewReader(logExemplo), gz, apenasErros)
	gz.Close()

	leitorGz, _ := gzip.NewReader(&comprimido)
	original, _ := io.ReadAll(leitorGz)
	fmt.Print(string(original))
}
