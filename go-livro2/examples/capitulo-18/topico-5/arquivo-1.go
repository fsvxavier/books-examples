// Command e05 cria valores com reflexão e usa DeepEqual, TypeAssert e Seq.
package main

import (
	"fmt"
	"reflect"
	"slices"
)

type Ponto struct{ X, Y int }

func main() {
	// Criar valores sem conhecer o tipo em compilação.
	t := reflect.TypeFor[Ponto]()
	p := reflect.New(t).Elem() // um Ponto zerado, endereçável
	p.Field(0).SetInt(3)
	p.FieldByName("Y").SetInt(4)
	fmt.Printf("reflect.New: %+v\n", p.Interface())

	fatia := reflect.MakeSlice(reflect.TypeFor[[]string](), 0, 4)
	fatia = reflect.Append(fatia, reflect.ValueOf("a"), reflect.ValueOf("b"))
	fmt.Println("MakeSlice+Append:", fatia.Interface(), "| len:", fatia.Len())

	mapa := reflect.MakeMap(reflect.TypeFor[map[string]int]())
	mapa.SetMapIndex(reflect.ValueOf("um"), reflect.ValueOf(1))
	fmt.Println("MakeMap:", mapa.Interface())

	// MakeFunc: uma função cuja implementação é uma função de reflect.Value.
	var somar func(int, int) int
	reflect.ValueOf(&somar).Elem().Set(reflect.MakeFunc(reflect.TypeOf(somar), func(args []reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(int(args[0].Int() + args[1].Int()))}
	}))
	fmt.Println("MakeFunc: somar(20, 22) =", somar(20, 22))

	// TypeAssert (Go 1.25): como uma asserção de tipo, sobre um Value, sem alocar.
	v := reflect.ValueOf(Ponto{1, 2})
	pt, ok := reflect.TypeAssert[Ponto](v)
	_, okErrado := reflect.TypeAssert[string](v)
	fmt.Println("TypeAssert[Ponto]:", pt, ok, "| TypeAssert[string]:", okErrado)

	// Seq e Seq2 (Go 1.23): iterar sobre um Value como um range. Seq produz o que um
	// "for x := range" de UMA variável produziria: os ÍNDICES de uma fatia, não os valores.
	fmt.Println("CanSeq de []int:", reflect.TypeFor[[]int]().CanSeq(), "| de int (range sobre inteiro):", reflect.TypeFor[int]().CanSeq(), "| de string:", reflect.TypeFor[string]().CanSeq2())
	var indices []int64
	for i := range reflect.ValueOf([]string{"a", "b", "c"}).Seq() {
		indices = append(indices, i.Int())
	}
	fmt.Println("Seq sobre []string (índices):", indices)
	var soma int64
	for _, x := range reflect.ValueOf([]int{10, 20, 30}).Seq2() {
		soma += x.Int()
	}
	fmt.Println("Seq2 sobre []int (índice e valor), soma dos valores:", soma)
	var chaves []string
	for k := range reflect.ValueOf(map[string]int{"b": 2, "a": 1}).Seq() {
		chaves = append(chaves, k.String())
	}
	slices.Sort(chaves)
	fmt.Println("Seq sobre mapa (chaves):", chaves)

	// DeepEqual e as suas pegadinhas.
	fmt.Println("\nDeepEqual:")
	fmt.Println("  []int{1} e []int{1}:               ", reflect.DeepEqual([]int{1}, []int{1}))
	fmt.Println("  []int(nil) e []int{}:               ", reflect.DeepEqual([]int(nil), []int{}), "(nil e vazio são diferentes)")
	fmt.Println("  map nil e map vazio:                ", reflect.DeepEqual(map[string]int(nil), map[string]int{}))
	fmt.Println("  int(1) e int64(1):                  ", reflect.DeepEqual(1, int64(1)), "(tipos diferentes)")
	fmt.Println("  &Ponto{1,2} e &Ponto{1,2}:          ", reflect.DeepEqual(&Ponto{1, 2}, &Ponto{1, 2}), "(compara o conteúdo apontado)")
	fmt.Println("  slices.Equal([]int{1}, []int{1}):   ", slices.Equal([]int{1}, []int{1}))
	f1, f2 := func() {}, func() {}
	fmt.Println("  funções (nenhuma é nil):            ", reflect.DeepEqual(f1, f2), "(funções só são iguais se as duas forem nil)")
}
