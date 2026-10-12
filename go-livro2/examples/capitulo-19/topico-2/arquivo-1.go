// Command e01 mostra o tamanho, o alinhamento e o preenchimento (padding) dos tipos do Go.
package main

import (
	"fmt"
	"unsafe"
)

// Desordenada alterna campos pequenos e grandes: o compilador insere preenchimento.
type Desordenada struct {
	Ativo  bool  // 1 byte, depois 7 de preenchimento
	Total  int64 // 8
	Nivel  int8  // 1, depois 7 de preenchimento
	Criado int64 // 8
	Versao int16 // 2, depois 6 de preenchimento no fim
}

// Ordenada tem os mesmos campos, do maior para o menor.
type Ordenada struct {
	Total  int64
	Criado int64
	Versao int16
	Nivel  int8
	Ativo  bool
}

// Vazia não ocupa espaço.
type Vazia struct{}

// ComVazia termina com um campo de tamanho zero: o compilador acrescenta preenchimento.
type ComVazia struct {
	N int32
	_ struct{}
}

func main() {
	fmt.Println("Tipos básicos (tamanho / alinhamento):")
	fmt.Printf("  bool %d/%d | int8 %d/%d | int16 %d/%d | int32 %d/%d | int64 %d/%d | float64 %d/%d | int %d/%d\n",
		unsafe.Sizeof(true), unsafe.Alignof(true), unsafe.Sizeof(int8(0)), unsafe.Alignof(int8(0)),
		unsafe.Sizeof(int16(0)), unsafe.Alignof(int16(0)), unsafe.Sizeof(int32(0)), unsafe.Alignof(int32(0)),
		unsafe.Sizeof(int64(0)), unsafe.Alignof(int64(0)), unsafe.Sizeof(float64(0)), unsafe.Alignof(float64(0)),
		unsafe.Sizeof(0), unsafe.Alignof(0))

	var (
		s  string
		sl []int
		m  map[string]int
		i  any
		e  error
		f  func()
		c  chan int
		p  *int
		a  [3]int32
	)
	fmt.Println("Tipos com cabeçalho (o tamanho NÃO inclui o que eles apontam):")
	fmt.Printf("  string %d | slice %d | map %d | interface any %d | error %d | func %d | chan %d | *int %d | [3]int32 %d\n",
		unsafe.Sizeof(s), unsafe.Sizeof(sl), unsafe.Sizeof(m), unsafe.Sizeof(i), unsafe.Sizeof(e),
		unsafe.Sizeof(f), unsafe.Sizeof(c), unsafe.Sizeof(p), unsafe.Sizeof(a))

	var d Desordenada
	var o Ordenada
	fmt.Println("\nOs mesmos cinco campos, em duas ordens:")
	fmt.Printf("  Desordenada: tamanho %d, alinhamento %d | offsets: Ativo %d, Total %d, Nivel %d, Criado %d, Versao %d\n",
		unsafe.Sizeof(d), unsafe.Alignof(d), unsafe.Offsetof(d.Ativo), unsafe.Offsetof(d.Total),
		unsafe.Offsetof(d.Nivel), unsafe.Offsetof(d.Criado), unsafe.Offsetof(d.Versao))
	fmt.Printf("  Ordenada:    tamanho %d, alinhamento %d | offsets: Total %d, Criado %d, Versao %d, Nivel %d, Ativo %d\n",
		unsafe.Sizeof(o), unsafe.Alignof(o), unsafe.Offsetof(o.Total), unsafe.Offsetof(o.Criado),
		unsafe.Offsetof(o.Versao), unsafe.Offsetof(o.Nivel), unsafe.Offsetof(o.Ativo))
	soma := unsafe.Sizeof(d.Ativo) + unsafe.Sizeof(d.Total) + unsafe.Sizeof(d.Nivel) + unsafe.Sizeof(d.Criado) + unsafe.Sizeof(d.Versao)
	fmt.Printf("  soma dos campos: %d bytes; preenchimento na Desordenada: %d bytes; na Ordenada: %d bytes\n",
		soma, unsafe.Sizeof(d)-soma, unsafe.Sizeof(o)-soma)

	fmt.Println("\nCampo de tamanho zero:")
	fmt.Printf("  Vazia %d | struct{N int32} %d | ComVazia (N int32 e _ struct{}) %d\n",
		unsafe.Sizeof(Vazia{}), unsafe.Sizeof(struct{ N int32 }{}), unsafe.Sizeof(ComVazia{}))

	// Em um slice, o preenchimento se repete a cada elemento.
	const n = 1_000_000
	fmt.Printf("\nUm slice de %d elementos: Desordenada %d MB, Ordenada %d MB\n",
		n, n*unsafe.Sizeof(d)/1_000_000, n*unsafe.Sizeof(o)/1_000_000)
}
