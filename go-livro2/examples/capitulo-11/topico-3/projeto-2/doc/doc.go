// Package doc mostra duas formas de encapsular o Memento além do tipo
// com campos não exportados: a interface selada e o memento estrito.
package doc

import (
	"errors"
	"fmt"
)

type Doc struct {
	nome  string
	texto string
}

func Novo(nome string) *Doc { return &Doc{nome: nome} }

func (d *Doc) Escrever(s string) { d.texto += s }
func (d *Doc) Texto() string     { return d.texto }
func (d *Doc) String() string    { return fmt.Sprintf("%s=%q", d.nome, d.texto) }

// --- Forma 2: interface selada -----------------------------------------
//
// O guardião só enxerga Snapshot. Como restaurar não é exportado, ninguém
// fora do pacote consegue fabricar um Snapshot nem inspecionar o conteúdo.

type Snapshot interface {
	Descricao() string
	restaurar(*Doc)
}

type instantaneo struct {
	nome  string
	texto string
}

func (i instantaneo) Descricao() string { return fmt.Sprintf("%s (%d bytes)", i.nome, len(i.texto)) }
func (i instantaneo) restaurar(d *Doc)  { d.nome, d.texto = i.nome, i.texto }

func (d *Doc) Foto() Snapshot { return instantaneo{d.nome, d.texto} }

func (d *Doc) Voltar(s Snapshot) { s.restaurar(d) }

// --- Forma 3: memento estrito ------------------------------------------
//
// O memento conhece o originador que o criou e se restaura sozinho. O
// guardião só chama Restaurar e nunca precisa passar o originador.

type Estrito struct {
	origem *Doc
	nome   string
	texto  string
}

var ErrOutroDocumento = errors.New("doc: foto de outro documento")

func (d *Doc) Estrito() *Estrito { return &Estrito{d, d.nome, d.texto} }

func (e *Estrito) Restaurar() {
	e.origem.nome, e.origem.texto = e.nome, e.texto
}

// Em contraste, a forma 1 (Memento com campos não exportados, como em e01)
// permite aplicar a foto de um documento em OUTRO. Aqui a checagem é
// explícita:
func (d *Doc) VoltarChecado(s Snapshot) error {
	i, ok := s.(instantaneo)
	if !ok || i.nome != d.nome {
		return ErrOutroDocumento
	}
	s.restaurar(d)
	return nil
}
