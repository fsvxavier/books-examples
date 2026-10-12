package main

import (
	"errors"
	"fmt"
)

// ---------- Abordagem 1: TRANSPARENTE (Adicionar na interface do componente) ----------

var ErrFolha = errors.New("folha nao tem filhos")

type ItemT interface {
	Tamanho() int
	Adicionar(ItemT) error // existe em toda folha, mas la so devolve erro
}

type ArquivoT struct{ Bytes int }

func (a ArquivoT) Tamanho() int          { return a.Bytes }
func (a ArquivoT) Adicionar(ItemT) error { return ErrFolha }

type PastaT struct{ filhos []ItemT }

func (p *PastaT) Tamanho() int {
	t := 0
	for _, f := range p.filhos {
		t += f.Tamanho()
	}
	return t
}
func (p *PastaT) Adicionar(i ItemT) error { p.filhos = append(p.filhos, i); return nil }

// ---------- Abordagem 2: SEGURA (so o container tem Adicionar) ----------

type Item interface{ Tamanho() int }

type Arquivo struct{ Bytes int }

func (a Arquivo) Tamanho() int { return a.Bytes }

type Pasta struct{ filhos []Item }

func (p *Pasta) Tamanho() int {
	t := 0
	for _, f := range p.filhos {
		t += f.Tamanho()
	}
	return t
}
func (p *Pasta) Adicionar(i Item) { p.filhos = append(p.filhos, i) }

// Container e a interface OPCIONAL: quem quiser montar a arvore pede por ela.
type Container interface {
	Item
	Adicionar(Item)
}

func montar(raiz Item, filhos ...Item) error {
	c, ok := raiz.(Container)
	if !ok {
		return fmt.Errorf("%T nao e um container", raiz)
	}
	for _, f := range filhos {
		c.Adicionar(f)
	}
	return nil
}

func main() {
	fmt.Println("Abordagem transparente:")
	var raizT ItemT = &PastaT{}
	fmt.Println("  adicionar a pasta:", raizT.Adicionar(ArquivoT{100}))
	var folhaT ItemT = ArquivoT{50}
	fmt.Println("  adicionar a arquivo (compila, falha em execucao):", folhaT.Adicionar(ArquivoT{1}))

	fmt.Println("Abordagem segura:")
	pasta := &Pasta{}
	fmt.Println("  montar na pasta:", montar(pasta, Arquivo{100}, Arquivo{200}), "| tamanho:", pasta.Tamanho())
	fmt.Println("  montar no arquivo:", montar(Arquivo{50}, Arquivo{1}))
	// Arquivo{50}.Adicionar(Arquivo{1}) nem compila: Arquivo nao tem o metodo Adicionar.
}
