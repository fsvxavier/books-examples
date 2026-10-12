package main

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
)

// Armazenamento e o COMPONENTE: grava e le um bloco de bytes.
type Armazenamento interface {
	Gravar(dados []byte) error
	Ler() ([]byte, error)
}

// Memoria e o componente CONCRETO: simula um arquivo em disco.
type Memoria struct{ guardado []byte }

func (m *Memoria) Gravar(d []byte) error { m.guardado = bytes.Clone(d); return nil }
func (m *Memoria) Ler() ([]byte, error)  { return bytes.Clone(m.guardado), nil }

// Bruto devolve o que esta realmente gravado (so para o experimento).
func (m *Memoria) Bruto() []byte { return m.guardado }

// ---------- Decoradores: mesma interface, guardam o proximo ----------

type Compressao struct{ prox Armazenamento }

func ComCompressao(p Armazenamento) *Compressao { return &Compressao{p} }

func (c *Compressao) Gravar(d []byte) error {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(d); err != nil {
		return err
	}
	if err := zw.Close(); err != nil { // Close descarrega o fim do fluxo
		return err
	}
	return c.prox.Gravar(buf.Bytes())
}

func (c *Compressao) Ler() ([]byte, error) {
	d, err := c.prox.Ler()
	if err != nil {
		return nil, err
	}
	zr, err := gzip.NewReader(bytes.NewReader(d))
	if err != nil {
		return nil, fmt.Errorf("descomprimindo: %w", err)
	}
	return io.ReadAll(zr)
}

type Criptografia struct {
	prox Armazenamento
	aead cipher.AEAD
}

func ComCriptografia(p Armazenamento, chave []byte) (*Criptografia, error) {
	bloco, err := aes.NewCipher(chave)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCMWithRandomNonce(bloco)
	if err != nil {
		return nil, err
	}
	return &Criptografia{p, aead}, nil
}

func (c *Criptografia) Gravar(d []byte) error {
	return c.prox.Gravar(c.aead.Seal(nil, nil, d, nil))
}

func (c *Criptografia) Ler() ([]byte, error) {
	d, err := c.prox.Ler()
	if err != nil {
		return nil, err
	}
	claro, err := c.aead.Open(nil, nil, d, nil)
	if err != nil {
		return nil, fmt.Errorf("decifrando: %w", err)
	}
	return claro, nil
}
