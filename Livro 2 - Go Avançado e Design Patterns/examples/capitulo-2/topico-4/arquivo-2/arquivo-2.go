package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// ---- 1. Fabrica que devolve o recurso E a funcao que o libera (como context.WithCancel) ----

func NovoArquivoTemporario(conteudo string) (caminho string, limpar func(), err error) {
	dir, err := os.MkdirTemp("", "fab")
	if err != nil {
		return "", nil, err
	}
	caminho = filepath.Join(dir, "dado.txt")
	if err := os.WriteFile(caminho, []byte(conteudo), 0o600); err != nil {
		os.RemoveAll(dir)
		return "", nil, err
	}
	return caminho, func() { os.RemoveAll(dir) }, nil
}

// ---- 2. sync.Pool: o campo New e uma fabrica chamada quando o pool esta vazio ----

var criados atomic.Int64

var pool = sync.Pool{
	New: func() any {
		criados.Add(1)
		return new(bytes.Buffer)
	},
}

func formatar(n int) string {
	b := pool.Get().(*bytes.Buffer)
	defer func() { b.Reset(); pool.Put(b) }()
	fmt.Fprintf(b, "item-%04d", n)
	return b.String()
}

func main() {
	caminho, limpar, err := NovoArquivoTemporario("ola")
	if err != nil {
		panic(err)
	}
	dados, _ := os.ReadFile(caminho)
	_, errAntes := os.Stat(caminho)
	limpar()
	_, errDepois := os.Stat(caminho)
	fmt.Printf("lido=%q existia=%v existe_apos_limpar=%v\n", dados, errAntes == nil, errDepois == nil)

	for i := range 1000 {
		formatar(i)
	}
	fmt.Println("1000 formatacoes sequenciais criaram", criados.Load(), "buffer(s)")
}
