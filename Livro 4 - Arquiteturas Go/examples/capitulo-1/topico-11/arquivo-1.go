// Comando costuras: mostra, para um pacote, quais arquivos dependem de quais.
package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "uso: costuras <diretório>")
		os.Exit(2)
	}
	if err := executar(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "costuras:", err)
		os.Exit(1)
	}
}

func executar(dir string) error {
	nomes, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	var arquivos []*ast.File
	for _, n := range nomes {
		if strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, n, nil, 0)
		if err != nil {
			return err
		}
		arquivos = append(arquivos, f)
	}

	info := &types.Info{Uses: make(map[*ast.Ident]types.Object)}
	conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
	pkg, err := conf.Check("p", fset, arquivos, info)
	if err != nil {
		return err
	}

	// arestas[de][para] = quantas referências o arquivo "de" faz a nomes
	// declarados no arquivo "para".
	arestas := make(map[string]map[string]int)
	for id, obj := range info.Uses {
		if obj.Pkg() != pkg {
			continue
		}
		de := filepath.Base(fset.Position(id.Pos()).Filename)
		para := filepath.Base(fset.Position(obj.Pos()).Filename)
		if de == para {
			continue
		}
		if arestas[de] == nil {
			arestas[de] = make(map[string]int)
		}
		arestas[de][para]++
	}

	origens := make([]string, 0, len(arestas))
	for de := range arestas {
		origens = append(origens, de)
	}
	sort.Strings(origens)
	for _, de := range origens {
		destinos := make([]string, 0)
		for para := range arestas[de] {
			destinos = append(destinos, para)
		}
		slices.Sort(destinos)
		for _, para := range destinos {
			fmt.Printf("%-14s -> %-14s %3d referências\n", de, para, arestas[de][para])
		}
	}
	return nil
}
