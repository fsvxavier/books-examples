package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ---------- RECEPTOR: o editor de texto. Nao sabe nada de comandos nem de historico. ----------

type Editor struct {
	texto []rune
}

func NovoEditor(s string) *Editor { return &Editor{texto: []rune(s)} }

func (e *Editor) Texto() string { return string(e.texto) }

func (e *Editor) Inserir(pos int, s string) error {
	if pos < 0 || pos > len(e.texto) {
		return fmt.Errorf("posicao %d fora de [0, %d]", pos, len(e.texto))
	}
	novo := make([]rune, 0, len(e.texto)+utf8.RuneCountInString(s))
	novo = append(novo, e.texto[:pos]...)
	novo = append(novo, []rune(s)...)
	novo = append(novo, e.texto[pos:]...)
	e.texto = novo
	return nil
}

func (e *Editor) Apagar(pos, n int) (string, error) {
	if pos < 0 || n < 0 || pos+n > len(e.texto) {
		return "", fmt.Errorf("intervalo [%d, %d) fora do texto de %d caracteres", pos, pos+n, len(e.texto))
	}
	apagado := string(e.texto[pos : pos+n])
	e.texto = append(e.texto[:pos:pos], e.texto[pos+n:]...)
	return apagado, nil
}

func (e *Editor) Definir(s string) { e.texto = []rune(s) }

// ---------- COMANDO: uma operacao como OBJETO, com tudo que e preciso para refaze-la e desfaze-la ----------

type Comando interface {
	Executar() error
	Desfazer() error
	Descricao() string
}

// Inserir desfaz-se pela OPERACAO INVERSA (apagar o que inseriu): nao precisa guardar o texto todo.
type InserirTexto struct {
	Ed    *Editor
	Pos   int
	Texto string
}

func (c *InserirTexto) Executar() error { return c.Ed.Inserir(c.Pos, c.Texto) }
func (c *InserirTexto) Desfazer() error {
	_, err := c.Ed.Apagar(c.Pos, utf8.RuneCountInString(c.Texto))
	return err
}
func (c *InserirTexto) Descricao() string { return fmt.Sprintf("inserir %q em %d", c.Texto, c.Pos) }

// ApagarTrecho guarda o que apagou NA EXECUCAO, para poder desfazer.
type ApagarTrecho struct {
	Ed      *Editor
	Pos, N  int
	apagado string
}

func (c *ApagarTrecho) Executar() error {
	s, err := c.Ed.Apagar(c.Pos, c.N)
	if err != nil {
		return err
	}
	c.apagado = s
	return nil
}
func (c *ApagarTrecho) Desfazer() error { return c.Ed.Inserir(c.Pos, c.apagado) }
func (c *ApagarTrecho) Descricao() string {
	return fmt.Sprintf("apagar %d caracteres em %d", c.N, c.Pos)
}

// Maiusculas nao tem inversa simples (perde-se a informacao de quais letras eram maiusculas):
// usa um BACKUP (snapshot) do estado, a segunda estrategia que o catalogo descreve.
type Maiusculas struct {
	Ed     *Editor
	backup string
}

func (c *Maiusculas) Executar() error {
	c.backup = c.Ed.Texto()
	c.Ed.Definir(strings.ToUpper(c.backup))
	return nil
}
func (c *Maiusculas) Desfazer() error   { c.Ed.Definir(c.backup); return nil }
func (c *Maiusculas) Descricao() string { return "converter para maiusculas" }

// ---------- INVOCADOR: executa comandos e guarda o historico ----------

var (
	ErroNadaParaDesfazer = errors.New("nada para desfazer")
	ErroNadaParaRefazer  = errors.New("nada para refazer")
)

type Historico struct {
	feitos    []Comando
	desfeitos []Comando
	Limite    int // 0 = sem limite
}

// Executar roda o comando; so entra no historico se deu certo. Um comando novo ENCERRA a trilha do "refazer".
func (h *Historico) Executar(c Comando) error {
	if err := c.Executar(); err != nil {
		return fmt.Errorf("%s: %w", c.Descricao(), err)
	}
	h.feitos = append(h.feitos, c)
	h.desfeitos = nil
	if h.Limite > 0 && len(h.feitos) > h.Limite {
		h.feitos = h.feitos[len(h.feitos)-h.Limite:]
	}
	return nil
}

func (h *Historico) Desfazer() error {
	if len(h.feitos) == 0 {
		return ErroNadaParaDesfazer
	}
	c := h.feitos[len(h.feitos)-1]
	if err := c.Desfazer(); err != nil {
		return fmt.Errorf("desfazer %s: %w", c.Descricao(), err)
	}
	h.feitos = h.feitos[:len(h.feitos)-1]
	h.desfeitos = append(h.desfeitos, c)
	return nil
}

func (h *Historico) Refazer() error {
	if len(h.desfeitos) == 0 {
		return ErroNadaParaRefazer
	}
	c := h.desfeitos[len(h.desfeitos)-1]
	if err := c.Executar(); err != nil {
		return fmt.Errorf("refazer %s: %w", c.Descricao(), err)
	}
	h.desfeitos = h.desfeitos[:len(h.desfeitos)-1]
	h.feitos = append(h.feitos, c)
	return nil
}

func (h *Historico) Tamanho() (feitos, desfeitos int) { return len(h.feitos), len(h.desfeitos) }
