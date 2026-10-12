// Command e03 mostra o padrão do ponteiro com método (PT interface{ *T; M() }).
package main

import (
	"fmt"
	"strings"
)

// Conexao tem um método de ponteiro: Iniciar altera o valor.
type Conexao struct {
	Endereco string
	Pronta   bool
}

// Iniciar liga a conexão. O receptor é ponteiro porque o método altera o valor.
func (c *Conexao) Iniciar(endereco string) {
	c.Endereco = endereco
	c.Pronta = true
}

// Fila é outro tipo que também precisa de Iniciar.
type Fila struct {
	Nome  string
	Itens []string
}

// Iniciar prepara a fila.
func (f *Fila) Iniciar(nome string) {
	f.Nome = nome
	f.Itens = make([]string, 0, 4)
}

// Iniciavel exige o método, mas com T como VALOR: nenhum tipo com receptor de ponteiro satisfaz.
type Iniciavel interface{ Iniciar(string) }

// Novo cria um T e o inicia. O padrão: T é o tipo do valor, PT é *T e tem o método.
// A constraint de PT diz duas coisas: "PT é um ponteiro para T" e "PT tem Iniciar".
func Novo[T any, PT interface {
	*T
	Iniciar(string)
}](arg string) *T {
	var v T
	PT(&v).Iniciar(arg) // &v é *T; a conversão para PT dá acesso ao método
	return &v
}

// NovosVarios cria n valores iguais: PT é INFERIDO a partir de T.
func NovosVarios[T any, PT interface {
	*T
	Iniciar(string)
}](n int, arg string) []*T {
	saida := make([]*T, n)
	for i := range saida {
		saida[i] = Novo[T, PT](fmt.Sprintf("%s-%d", arg, i))
	}
	return saida
}

func main() {
	c := Novo[Conexao]("db:5432") // só T é escrito: o compilador infere PT = *Conexao
	fmt.Printf("%+v\n", *c)
	f := Novo[Fila]("pedidos")
	fmt.Printf("%+v\n", *f)

	filas := NovosVarios[Fila](3, "q")
	nomes := make([]string, len(filas))
	for i, q := range filas {
		nomes[i] = q.Nome
	}
	fmt.Println(strings.Join(nomes, ","))
}
