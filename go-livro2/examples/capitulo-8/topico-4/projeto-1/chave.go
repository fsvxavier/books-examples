package main

import (
	"errors"
	"fmt"
)

// Chave descreve o estado INTRINSECO: e comparavel, entao serve de chave de mapa.
type Chave struct {
	Nome string
	Cor  string
}

// Tipo e o flyweight: campos nao exportados e sem setters, so metodos de leitura.
// Uma vez criado, ninguem de fora do pacote consegue altera-lo.
type Tipo struct {
	nome    string
	cor     string
	textura []byte
}

func (t *Tipo) Nome() string { return t.nome }
func (t *Tipo) Cor() string  { return t.cor }

// Textura devolve uma COPIA: expor o slice interno deixaria qualquer cliente corromper o
// estado compartilhado por todos os outros.
func (t *Tipo) Textura() []byte { return append([]byte(nil), t.textura...) }

func (t *Tipo) Desenhar(x, y int) string {
	return fmt.Sprintf("%s %s em (%d,%d), textura %d bytes", t.cor, t.nome, x, y, len(t.textura))
}

var ErrChave = errors.New("chave invalida")

func criarTipo(c Chave) (*Tipo, error) {
	if c.Nome == "" {
		return nil, fmt.Errorf("%w: nome vazio", ErrChave)
	}
	return &Tipo{nome: c.Nome, cor: c.Cor, textura: make([]byte, 4096)}, nil
}
