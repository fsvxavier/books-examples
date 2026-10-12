// Command e05 usa unsafe para ler e escrever um campo não exportado, que a reflexão recusa.
package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Conta é de um "outro pacote" (aqui, só para o exemplo): o campo saldo não é exportado.
type Conta struct {
	Titular string
	saldo   int64
}

func (c Conta) Saldo() int64 { return c.saldo }

func main() {
	c := &Conta{Titular: "Ana", saldo: 100}

	// 1. A reflexão recusa escrever.
	campo := reflect.ValueOf(c).Elem().FieldByName("saldo")
	fmt.Println("reflexão: CanSet =", campo.CanSet())

	// 2. Com unsafe: o endereço do campo vem de UnsafeAddr (padrão 5) e vira um ponteiro tipado.
	acessivel := reflect.NewAt(campo.Type(), unsafe.Pointer(campo.UnsafeAddr())).Elem()
	fmt.Println("via NewAt: CanSet =", acessivel.CanSet(), "| valor lido =", acessivel.Int())
	acessivel.SetInt(999_999)
	fmt.Println("depois de SetInt: c.Saldo() =", c.Saldo())

	// 3. Sem reflexão: o endereço do campo pela posição (Offsetof) e unsafe.Add.
	p := (*int64)(unsafe.Add(unsafe.Pointer(c), unsafe.Offsetof(c.saldo)))
	*p = 1
	fmt.Println("via Offsetof + Add: c.Saldo() =", c.Saldo())
	fmt.Println("(Offsetof de saldo =", unsafe.Offsetof(c.saldo), "bytes: depende do layout; se o autor da struct trocar a ordem dos campos, isto quebra em silêncio)")
}
