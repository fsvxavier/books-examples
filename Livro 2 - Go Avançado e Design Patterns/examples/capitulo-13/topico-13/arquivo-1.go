package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
	"testing/fstest"
)

const fonte = `package demo

import "fmt"

func Soma(a, b int) int { return a + b }

func Mostra(xs []int) {
	for _, x := range xs {
		fmt.Println(Soma(x, 1))
	}
}

type T struct{}

func (T) Metodo() { Mostra(nil) }
`

// Contador implementa ast.Visitor. O pacote go/ast é o "molde" de
// travessia (Walk); Visit é o passo variável.
type Contador struct {
	Funcoes  int
	Chamadas map[string]int
	profund  int
	MaxProf  int
}

// Visit é chamado para cada nó. Devolve o visitante a usar nos FILHOS
// (ou nil para não descer).
func (c *Contador) Visit(n ast.Node) ast.Visitor {
	if n == nil { // chamado ao SAIR de um nó
		c.profund--
		return nil
	}
	c.profund++
	c.MaxProf = max(c.MaxProf, c.profund)
	switch n := n.(type) {
	case *ast.FuncDecl:
		c.Funcoes++
	case *ast.CallExpr:
		switch f := n.Fun.(type) {
		case *ast.Ident:
			c.Chamadas[f.Name]++
		case *ast.SelectorExpr:
			if x, ok := f.X.(*ast.Ident); ok {
				c.Chamadas[x.Name+"."+f.Sel.Name]++
			}
		}
	}
	return c
}

func main() {
	fset := token.NewFileSet()
	arq, err := parser.ParseFile(fset, "demo.go", fonte, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("== ast.Walk com um visitante com estado ==")
	c := &Contador{Chamadas: map[string]int{}}
	ast.Walk(c, arq)
	fmt.Printf("funções: %d | profundidade máxima: %d | chamadas: %v\n", c.Funcoes, c.MaxProf, c.Chamadas)

	fmt.Println("== ast.Inspect: a mesma travessia com UMA função ==")
	var nomes []string
	ast.Inspect(arq, func(n ast.Node) bool {
		if f, ok := n.(*ast.FuncDecl); ok {
			nomes = append(nomes, f.Name.Name)
			return false // não desce no corpo da função
		}
		return true
	})
	fmt.Println("funções declaradas:", strings.Join(nomes, ", "))

	fmt.Println("== fs.WalkDir: o visitante controla a travessia ==")
	sistema := fstest.MapFS{
		"README.md":         {Data: []byte("leia-me")},
		"src/main.go":       {Data: []byte("package main")},
		"src/util.go":       {Data: []byte("package main // util")},
		"node_modules/x.js": {Data: []byte("lixo lixo lixo lixo")},
		"node_modules/y.js": {Data: []byte("mais lixo")},
		"docs/guia.md":      {Data: []byte("guia")},
	}
	total, visitados := 0, 0
	_ = fs.WalkDir(sistema, ".", func(caminho string, d fs.DirEntry, err error) error {
		visitados++
		if d.IsDir() && d.Name() == "node_modules" {
			return fs.SkipDir // poda a subárvore inteira
		}
		if !d.IsDir() {
			info, _ := d.Info()
			total += int(info.Size())
		}
		return nil
	})
	fmt.Printf("bytes em arquivos fora de node_modules: %d (entradas visitadas: %d)\n", total, visitados)
}
