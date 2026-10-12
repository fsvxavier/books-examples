// Package editor define o originador do Memento: um editor de texto simples.
package editor

import (
	"fmt"
	"strings"
)

// Editor é o originador: só ele conhece e altera o estado completo.
type Editor struct {
	linhas []string
	cursor int // índice da linha corrente
	titulo string
}

// Novo cria um editor com um título e uma linha vazia.
func Novo(titulo string) *Editor {
	return &Editor{linhas: []string{""}, titulo: titulo}
}

// Escrever acrescenta texto ao fim da linha corrente.
func (e *Editor) Escrever(texto string) {
	e.linhas[e.cursor] += texto
}

// NovaLinha abre uma linha depois da corrente e move o cursor para ela.
func (e *Editor) NovaLinha() {
	e.linhas = append(e.linhas, "")
	e.cursor = len(e.linhas) - 1
}

// Renomear troca o título.
func (e *Editor) Renomear(titulo string) { e.titulo = titulo }

// Texto devolve o conteúdo atual, uma linha por linha.
func (e *Editor) Texto() string { return strings.Join(e.linhas, "\n") }

func (e *Editor) String() string {
	return fmt.Sprintf("%s [%d linhas, cursor %d]", e.titulo, len(e.linhas), e.cursor)
}

// Memento é o instantâneo. Os campos são não exportados: quem está fora
// deste pacote pode guardar, passar e devolver um Memento, mas não lê-lo.
type Memento struct {
	linhas []string
	cursor int
	titulo string
}

// Descricao é o único metadado visível para o guardião.
func (m Memento) Descricao() string {
	return fmt.Sprintf("%q com %d linhas", m.titulo, len(m.linhas))
}

// Salvar cria um Memento com uma cópia do estado atual.
func (e *Editor) Salvar() Memento {
	return Memento{
		linhas: append([]string(nil), e.linhas...),
		cursor: e.cursor,
		titulo: e.titulo,
	}
}

// Restaurar volta ao estado guardado no Memento.
func (e *Editor) Restaurar(m Memento) {
	e.linhas = append([]string(nil), m.linhas...)
	e.cursor = m.cursor
	e.titulo = m.titulo
}
