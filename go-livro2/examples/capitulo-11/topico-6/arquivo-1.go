package main

import (
	"fmt"
	"strings"
)

// Estado é pequeno e só tem tipos de valor imutáveis (string e int):
// copiar a struct já é tirar um instantâneo completo.
type Estado struct {
	Texto  string
	Cursor int
}

type Editor struct {
	estado Estado
}

func (e *Editor) Inserir(s string) {
	t := e.estado.Texto
	e.estado.Texto = t[:e.estado.Cursor] + s + t[e.estado.Cursor:]
	e.estado.Cursor += len(s)
}

func (e *Editor) Apagar(n int) {
	c := e.estado.Cursor
	n = min(n, c)
	e.estado.Texto = e.estado.Texto[:c-n] + e.estado.Texto[c:]
	e.estado.Cursor = c - n
}

func (e *Editor) Salvar() Estado     { return e.estado }
func (e *Editor) Restaurar(s Estado) { e.estado = s }

func (e *Editor) String() string {
	return fmt.Sprintf("%q (cursor %d)", e.estado.Texto, e.estado.Cursor)
}

// Historico guarda duas pilhas: desfazer e refazer, com limite.
type Historico struct {
	desfazer []Estado
	refazer  []Estado
	limite   int
}

func NovoHistorico(limite int) *Historico { return &Historico{limite: limite} }

// Registrar guarda o estado ANTES de uma edição. Uma edição nova
// invalida o futuro: a pilha de refazer é descartada.
func (h *Historico) Registrar(antes Estado) {
	h.desfazer = append(h.desfazer, antes)
	if len(h.desfazer) > h.limite {
		h.desfazer = h.desfazer[1:]
	}
	h.refazer = nil
}

func (h *Historico) Desfazer(atual Estado) (Estado, bool) {
	if len(h.desfazer) == 0 {
		return Estado{}, false
	}
	n := len(h.desfazer) - 1
	anterior := h.desfazer[n]
	h.desfazer = h.desfazer[:n]
	h.refazer = append(h.refazer, atual)
	return anterior, true
}

func (h *Historico) Refazer(atual Estado) (Estado, bool) {
	if len(h.refazer) == 0 {
		return Estado{}, false
	}
	n := len(h.refazer) - 1
	proximo := h.refazer[n]
	h.refazer = h.refazer[:n]
	h.desfazer = append(h.desfazer, atual)
	return proximo, true
}

func main() {
	ed := &Editor{}
	h := NovoHistorico(3)

	editar := func(f func()) {
		h.Registrar(ed.Salvar())
		f()
		fmt.Println("editou:   ", ed)
	}

	editar(func() { ed.Inserir("Olá") })
	editar(func() { ed.Inserir(", mundo") })
	editar(func() { ed.Inserir("!") })
	editar(func() { ed.Apagar(1) })

	fmt.Println(strings.Repeat("-", 30))
	for range 4 {
		if s, ok := h.Desfazer(ed.Salvar()); ok {
			ed.Restaurar(s)
			fmt.Println("desfez:   ", ed)
		} else {
			fmt.Println("desfez:    nada (limite de 3 já consumido)")
		}
	}

	fmt.Println(strings.Repeat("-", 30))
	if s, ok := h.Refazer(ed.Salvar()); ok {
		ed.Restaurar(s)
		fmt.Println("refez:    ", ed)
	}

	editar(func() { ed.Inserir("?") })
	if _, ok := h.Refazer(ed.Salvar()); !ok {
		fmt.Println("refazer:   vazio, a edição nova cortou o futuro")
	}
}
