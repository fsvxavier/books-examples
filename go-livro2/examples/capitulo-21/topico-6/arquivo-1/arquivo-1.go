package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
)

// cache nunca remove nada: o defeito que o perfil de heap vai apontar.
var cache = map[string][]byte{}

// lixo recebe o buffer temporario, para o compilador nao o manter na pilha.
var lixo []byte

func atender(id int) {
	lixo = make([]byte, 4096) // temporario: aparece em alloc_space, nao em inuse_space
	cache["sessao-"+strconv.Itoa(id)] = make([]byte, 2048)
}

func fotografar(arquivo string) {
	runtime.GC() // a documentacao recomenda: o perfil reflete o ultimo GC concluido
	f, err := os.Create(arquivo)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := pprof.Lookup("heap").WriteTo(f, 0); err != nil {
		panic(err)
	}
}

func main() {
	for i := 0; i < 20000; i++ {
		atender(i)
	}
	fotografar("heap1.prof")
	fmt.Println("foto 1: 20.000 sessoes no cache")

	for i := 20000; i < 60000; i++ {
		atender(i)
	}
	fotografar("heap2.prof")
	fmt.Println("foto 2: 60.000 sessoes no cache")
}
