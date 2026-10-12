package main

import (
	"fmt"
	"maps"
	"slices"
	"sort"
)

// Clonavel e a interface do Prototype: um tipo que sabe se copiar (profundamente).
type Clonavel[T any] interface{ Clone() T }

// Prototipos guarda modelos prontos e entrega CLONES, nunca o proprio modelo.
type Prototipos[T Clonavel[T]] struct{ modelos map[string]T }

func NovoPrototipos[T Clonavel[T]]() *Prototipos[T] { return &Prototipos[T]{modelos: map[string]T{}} }

func (p *Prototipos[T]) Guardar(nome string, modelo T) { p.modelos[nome] = modelo.Clone() } // guarda uma copia

func (p *Prototipos[T]) Novo(nome string) (T, error) {
	m, ok := p.modelos[nome]
	if !ok {
		var zero T
		return zero, fmt.Errorf("modelo %q nao existe", nome)
	}
	return m.Clone(), nil
}

func (p *Prototipos[T]) Nomes() []string {
	n := slices.Collect(maps.Keys(p.modelos))
	sort.Strings(n)
	return n
}

// ---- um tipo concreto: configuracao de um job pre-pronta ----

type Job struct {
	Nome       string
	Comando    []string
	Ambiente   map[string]string
	Tentativas int
}

func (j Job) Clone() Job {
	j.Comando = slices.Clone(j.Comando)
	j.Ambiente = maps.Clone(j.Ambiente)
	return j
}

func main() {
	p := NovoPrototipos[Job]()
	p.Guardar("backup", Job{
		Nome: "backup", Comando: []string{"pg_dump", "livro"},
		Ambiente: map[string]string{"PGHOST": "db"}, Tentativas: 3,
	})
	p.Guardar("relatorio", Job{Nome: "relatorio", Comando: []string{"gerar", "--mensal"}, Tentativas: 1})
	fmt.Println("modelos:", p.Nomes())

	a, _ := p.Novo("backup")
	b, _ := p.Novo("backup")
	a.Nome = "backup-noturno"
	a.Comando = append(a.Comando, "--compress")
	a.Ambiente["PGHOST"] = "replica"
	b.Tentativas = 5

	fmt.Printf("a: %+v\n", a)
	fmt.Printf("b: %+v\n", b)
	c, _ := p.Novo("backup")
	fmt.Printf("modelo intacto: %+v\n", c)

	_, err := p.Novo("inexistente")
	fmt.Println("erro:", err)
}
