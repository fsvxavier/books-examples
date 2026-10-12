package main

import (
	"errors"
	"fmt"
)

type Pedido struct {
	ID    int
	Total int // centavos
}

// ---- Interface gorda: 7 métodos. Todo dublê precisa saber de todos. ----

type GerenciadorPedidos interface {
	Criar(Pedido) error
	Buscar(id int) (Pedido, error)
	Atualizar(Pedido) error
	Remover(id int) error
	Listar() ([]Pedido, error)
	ExportarPDF() ([]byte, error)
	EnviarEmail(Pedido) error
}

// Função que só precisa ler um pedido, mas recebe a interface inteira.
func TotalComImpostoGordo(g GerenciadorPedidos, id int) (int, error) {
	p, err := g.Buscar(id)
	if err != nil {
		return 0, err
	}
	return p.Total * 110 / 100, nil
}

// Dublê "preguiçoso": embute a interface (nil) para só implementar o que o teste usa.
type fakeGordo struct {
	GerenciadorPedidos
}

func (fakeGordo) Buscar(id int) (Pedido, error) { return Pedido{ID: id, Total: 1000}, nil }

// ---- Interface pequena, declarada por quem consome. ----

type Buscador interface {
	Buscar(id int) (Pedido, error)
}

func TotalComImposto(b Buscador, id int) (int, error) {
	p, err := b.Buscar(id)
	if err != nil {
		return 0, err
	}
	return p.Total * 110 / 100, nil
}

// Dublê para a versão pequena: uma função basta.
type buscadorFunc func(int) (Pedido, error)

func (f buscadorFunc) Buscar(id int) (Pedido, error) { return f(id) }

// ---- O dublê preguiçoso explode quando o código evolui. ----

func RelatorioDoPedido(g GerenciadorPedidos, id int) ([]byte, error) {
	p, err := g.Buscar(id)
	if err != nil {
		return nil, err
	}
	if err := g.EnviarEmail(p); err != nil { // código novo: chama outro método da interface gorda
		return nil, err
	}
	return g.ExportarPDF()
}

func main() {
	v, err := TotalComImpostoGordo(fakeGordo{}, 1)
	fmt.Println("gordo + dublê preguiçoso:", v, err)

	v, err = TotalComImposto(buscadorFunc(func(id int) (Pedido, error) {
		return Pedido{ID: id, Total: 1000}, nil
	}), 1)
	fmt.Println("interface pequena + função:", v, err)

	v, err = TotalComImposto(buscadorFunc(func(int) (Pedido, error) { return Pedido{}, errors.New("não achei") }), 1)
	fmt.Println("caminho de erro, mesma função:", v, err)

	defer func() {
		fmt.Println("depois da evolução do código:", recover())
	}()
	RelatorioDoPedido(fakeGordo{}, 1)
}
