package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main() {
	chave := bytes.Repeat([]byte{7}, 32)
	texto := []byte(strings.Repeat("pedido 1042;cliente ana;item caneta;", 300))
	fmt.Println("tamanho original:", len(texto), "bytes")

	// 1) so o armazenamento
	m1 := &Memoria{}
	m1.Gravar(texto)
	fmt.Println("sem decoradores:           ", len(m1.Bruto()), "bytes gravados")

	// 2) comprimir ANTES de cifrar. Gravar passa primeiro pelo decorador de FORA:
	//    ComCompressao(ComCriptografia(Memoria)) = comprime, depois cifra, depois grava.
	m2 := &Memoria{}
	cripto2, _ := ComCriptografia(m2, chave)
	var a Armazenamento = ComCompressao(cripto2) // fora: comprime; dentro: cifra
	a.Gravar(texto)
	fmt.Println("comprime, depois cifra:    ", len(m2.Bruto()), "bytes gravados")

	// 3) cifrar ANTES de comprimir: ComCriptografia(ComCompressao(Memoria)).
	m3 := &Memoria{}
	comp3 := ComCompressao(m3)
	var b Armazenamento
	b, _ = ComCriptografia(comp3, chave) // fora: cifra; dentro: comprime
	b.Gravar(texto)
	fmt.Println("cifra, depois comprime:    ", len(m3.Bruto()), "bytes gravados")

	// Ida e volta nas duas ordens.
	for _, c := range []struct {
		nome string
		s    Armazenamento
	}{{"comprime>cifra", a}, {"cifra>comprime", b}} {
		lido, err := c.s.Ler()
		fmt.Printf("ida e volta %-15s igual=%v err=%v\n", c.nome, bytes.Equal(lido, texto), err)
	}

	// Ler com a pilha ERRADA (outra ordem) nao decifra nada.
	m4 := &Memoria{}
	cripto4, _ := ComCriptografia(m4, chave)
	ComCompressao(cripto4).Gravar(texto)
	errada, _ := ComCriptografia(m4, chave) // sem descomprimir
	lido, _ := errada.Ler()
	fmt.Printf("ler sem descomprimir: %d bytes (comecam com gzip: %v)\n", len(lido), len(lido) > 2 && lido[0] == 0x1f && lido[1] == 0x8b)

	// Chave errada.
	outra, _ := ComCriptografia(m2, bytes.Repeat([]byte{9}, 32))
	_, err := outra.Ler()
	fmt.Println("chave errada:", err)
}
