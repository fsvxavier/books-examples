package main

import (
	"errors"
	"fmt"
	"strings"
)

// ---------- O SUBSISTEMA COMPLEXO (vem de uma biblioteca; nao sabe da fachada) ----------

type ArquivoVideo struct{ nome, codec string }

type FabricaCodec struct{}

func (FabricaCodec) Extrair(a ArquivoVideo) string { return a.codec }

type LeitorBitrate struct{}

func (LeitorBitrate) Ler(nome string, codec string) []byte {
	return []byte(strings.Repeat("v", len(nome)*10))
}

func (LeitorBitrate) Converter(buf []byte, destino string) []byte {
	return append([]byte(destino+":"), buf...)
}

type MixadorAudio struct{}

func (MixadorAudio) Normalizar(buf []byte) []byte { return buf }

// ---------- A FACHADA: UM metodo, o que o cliente de fato quer ----------

type Conversor struct {
	fabrica FabricaCodec
	leitor  LeitorBitrate
	audio   MixadorAudio
}

var ErrFormato = errors.New("formato nao suportado")

func (c Conversor) Codificar(nome, formato string) ([]byte, error) {
	if formato != "mp4" && formato != "ogg" {
		return nil, fmt.Errorf("%w: %q", ErrFormato, formato)
	}
	arq := ArquivoVideo{nome: nome, codec: "h264"}
	origem := c.fabrica.Extrair(arq)
	bruto := c.leitor.Ler(nome, origem)
	convertido := c.leitor.Converter(bruto, formato)
	return c.audio.Normalizar(convertido), nil
}

func main() {
	// Sem fachada, o cliente faria TODOS os passos, na ordem certa:
	var f FabricaCodec
	l := LeitorBitrate{}
	m := MixadorAudio{}
	arq := ArquivoVideo{nome: "gatinho.avi", codec: "h264"}
	manual := m.Normalizar(l.Converter(l.Ler(arq.nome, f.Extrair(arq)), "mp4"))
	fmt.Println("manual:  ", len(manual), "bytes")

	// Com a fachada, uma linha:
	saida, err := Conversor{}.Codificar("gatinho.avi", "mp4")
	fmt.Println("fachada: ", len(saida), "bytes, erro:", err)

	_, err = Conversor{}.Codificar("gatinho.avi", "wmv")
	fmt.Println("formato invalido:", err, "| errors.Is:", errors.Is(err, ErrFormato))
}
