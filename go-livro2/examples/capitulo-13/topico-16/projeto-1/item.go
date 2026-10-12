package main

import "fmt"

// Item é uma interface selada: só este pacote cria itens.
type Item interface{ isItem() }

type (
	Livro struct {
		Titulo string
		Preco  int // centavos
	}
	Eletronico struct {
		Nome          string
		Preco         int
		GarantiaMeses int
	}
	Alimento struct {
		Nome      string
		Preco     int
		Perecivel bool
	}
	Servico struct {
		Descricao string
		Horas     int
		ValorHora int
	}
)

func (Livro) isItem()      {}
func (Eletronico) isItem() {}
func (Alimento) isItem()   {}
func (Servico) isItem()    {}

// Visitante devolve um R por item.
type Visitante[R any] interface {
	VisitarLivro(Livro) R
	VisitarEletronico(Eletronico) R
	VisitarAlimento(Alimento) R
	VisitarServico(Servico) R
}

func Aceitar[R any](i Item, v Visitante[R]) R {
	switch i := i.(type) {
	case Livro:
		return v.VisitarLivro(i)
	case Eletronico:
		return v.VisitarEletronico(i)
	case Alimento:
		return v.VisitarAlimento(i)
	case Servico:
		return v.VisitarServico(i)
	}
	panic(fmt.Sprintf("Aceitar: item %T não tratado", i))
}

// Valor: operação 1 (R = int, em centavos).
type Valor struct{}

func (Valor) VisitarLivro(l Livro) int           { return l.Preco }
func (Valor) VisitarEletronico(e Eletronico) int { return e.Preco }
func (Valor) VisitarAlimento(a Alimento) int     { return a.Preco }
func (Valor) VisitarServico(s Servico) int       { return s.Horas * s.ValorHora }

// Imposto: operação 2 (R = int). As regras fiscais ficam TODAS aqui,
// fora dos tipos de item, e mudam sem tocar neles.
type Imposto struct{}

func pct(valor, centesimos int) int { return valor * centesimos / 10000 }

func (Imposto) VisitarLivro(Livro) int { return 0 } // imune
func (Imposto) VisitarEletronico(e Eletronico) int {
	return pct(e.Preco, 1800) // 18%
}
func (Imposto) VisitarAlimento(a Alimento) int {
	if a.Perecivel {
		return pct(a.Preco, 400) // 4%
	}
	return pct(a.Preco, 700) // 7%
}
func (Imposto) VisitarServico(s Servico) int {
	return pct(s.Horas*s.ValorHora, 500) // 5% de ISS
}

// Descricao: operação 3 (R = string).
type Descricao struct{}

func (Descricao) VisitarLivro(l Livro) string { return "Livro: " + l.Titulo }
func (Descricao) VisitarEletronico(e Eletronico) string {
	return fmt.Sprintf("%s (garantia %d meses)", e.Nome, e.GarantiaMeses)
}
func (Descricao) VisitarAlimento(a Alimento) string {
	if a.Perecivel {
		return a.Nome + " (perecível)"
	}
	return a.Nome
}
func (Descricao) VisitarServico(s Servico) string {
	return fmt.Sprintf("%s (%d h)", s.Descricao, s.Horas)
}

// Amostras: um item de cada tipo, para o teste de exaustividade.
func Amostras() []Item {
	return []Item{Livro{"l", 1}, Eletronico{"e", 1, 1}, Alimento{"a", 1, true}, Servico{"s", 1, 1}}
}
