// Command e02 percorre os campos de uma struct, com tags, campos embutidos e não exportados.
package main

import (
	"fmt"
	"reflect"
)

// Auditoria é embutida em Pedido: os campos dela são promovidos.
type Auditoria struct {
	CriadoPor string `json:"criado_por"`
}

// Pedido tem de tudo: exportado, não exportado, tags, struct embutida e um campo aninhado.
type Pedido struct {
	Auditoria
	ID       int      `json:"id" validate:"obrigatorio"`
	Cliente  string   `json:"cliente,omitempty" validate:"obrigatorio,min=3"`
	Itens    []string `json:"itens"`
	segredo  string
	Opcional *string `json:"opcional"`
}

func main() {
	t := reflect.TypeFor[Pedido]()
	fmt.Println("Campos de nível superior (Type.Fields):")
	for f := range t.Fields() {
		fmt.Printf("  #%d %-10s tipo=%-10s anonimo=%-5v exportado=%-5v json=%q validate=%q\n",
			f.Index[0], f.Name, f.Type, f.Anonymous, f.IsExported(), f.Tag.Get("json"), f.Tag.Get("validate"))
	}

	fmt.Println("\nTodos os campos visíveis, com os promovidos (reflect.VisibleFields):")
	for _, f := range reflect.VisibleFields(t) {
		fmt.Printf("  %-10s índice=%v\n", f.Name, f.Index)
	}

	// Lookup distingue "a tag existe e está vazia" de "a tag não existe".
	f, _ := t.FieldByName("Cliente")
	v, ok := f.Tag.Lookup("validate")
	fmt.Printf("\nLookup(validate) em Cliente: %q, %v\n", v, ok)
	v, ok = f.Tag.Lookup("xml")
	fmt.Printf("Lookup(xml) em Cliente:      %q, %v\n", v, ok)

	// Um campo promovido se acha pelo nome, e o índice é um caminho.
	cp, _ := t.FieldByName("CriadoPor")
	fmt.Println("CriadoPor: índice", cp.Index, "| json:", cp.Tag.Get("json"))

	// Value.Fields devolve o campo e o valor juntos; o campo não exportado aparece, mas não se lê com Interface.
	p := Pedido{Auditoria: Auditoria{"ana"}, ID: 7, Cliente: "Bruno", segredo: "xyz"}
	fmt.Println("\nValue.Fields:")
	for f, val := range reflect.ValueOf(p).Fields() {
		if f.IsExported() {
			fmt.Printf("  %-10s = %v\n", f.Name, val.Interface())
		} else {
			fmt.Printf("  %-10s = (não exportado: CanInterface=%v, mas String() mostra %q)\n", f.Name, val.CanInterface(), val.String())
		}
	}
}
