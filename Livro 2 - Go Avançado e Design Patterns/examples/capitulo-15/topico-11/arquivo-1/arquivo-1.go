package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"slices"
)

// ---- 1. nil tipado dentro de interface ----

type ErroValidacao struct{ Campo string }

func (e *ErroValidacao) Error() string { return "campo inválido: " + e.Campo }

func validarRuim(nome string) error {
	var err *ErroValidacao // ponteiro nil do tipo concreto
	if nome == "" {
		err = &ErroValidacao{Campo: "nome"}
	}
	return err // RUIM: vira uma interface error NÃO nula (tipo *ErroValidacao, valor nil)
}

func validar(nome string) error {
	if nome == "" {
		return &ErroValidacao{Campo: "nome"}
	}
	return nil // nil da interface, sem tipo
}

// ---- 2. fatias que compartilham o mesmo array ----

func primeirosDois(s []int) []int { return s[:2] }

// ---- 3. comparar erro embrulhado com == ----

func abrir() error {
	_, err := os.Open("/nao/existe")
	return fmt.Errorf("carregando configuração: %w", err)
}

// ---- 4. sombra de err ----

func gravar() error { return errors.New("disco cheio") }

func processarRuim() error {
	var err error
	if true {
		fmt.Println("gravando...")
		err := gravar() // := cria um err NOVO, visível só dentro deste bloco
		_ = err
	}
	return err // devolve o err de fora, que continua nil
}

func processar() error {
	var err error
	if true {
		fmt.Println("gravando...")
		err = gravar()
	}
	return err
}

// ---- 5. any demais ----

func total(dados map[string]any) int {
	return dados["quantidade"].(int) * 2 // pânico se vier float64 (ex.: JSON) ou ausente
}

func main() {
	fmt.Println("== nil tipado")
	fmt.Println("validarRuim(\"ana\") == nil?", validarRuim("ana") == nil)
	fmt.Printf("  conteúdo: %T %v\n", validarRuim("ana"), validarRuim("ana"))
	fmt.Println("validar(\"ana\") == nil?   ", validar("ana") == nil)

	fmt.Println("== fatias")
	original := []int{1, 2, 3, 4, 5}
	dois := primeirosDois(original)
	dois = append(dois, 99) // cabe na capacidade: escreve dentro do array original
	fmt.Println("subfatia comum:   original =", original, "| subfatia =", dois)

	original = []int{1, 2, 3, 4, 5}
	seguro := slices.Clone(original[:2])
	seguro = append(seguro, 42)
	fmt.Println("com slices.Clone: original =", original, "| cópia =", seguro)

	original = []int{1, 2, 3, 4, 5}
	limitada := original[:2:2] // expressão de 3 índices limita a capacidade
	limitada = append(limitada, 7)
	fmt.Println("com s[:2:2]:      original =", original, "| subfatia =", limitada)

	fmt.Println("== erro embrulhado")
	err := abrir()
	fmt.Println("err == fs.ErrNotExist?        ", err == fs.ErrNotExist)
	fmt.Println("errors.Is(err, fs.ErrNotExist)?", errors.Is(err, fs.ErrNotExist))
	var pe *fs.PathError
	fmt.Println("errors.As(err, &pe)?          ", errors.As(err, &pe), "caminho:", pe.Path)

	fmt.Println("== sombra de err")
	fmt.Println("processarRuim:", processarRuim())
	fmt.Println("processar:    ", processar())

	fmt.Println("== ordem do map")
	m := map[string]int{}
	for i := range 20 {
		m[fmt.Sprintf("k%02d", i)] = i
	}
	for i := range 3 {
		fmt.Printf("  volta %d:", i)
		n := 0
		for k := range m {
			fmt.Print(" ", k)
			if n++; n == 6 {
				break
			}
		}
		fmt.Println(" ...")
	}
	fmt.Println("  ordenado:", slices.Sorted(maps.Keys(m))[:6], "...")

	fmt.Println("== any demais")
	var dados map[string]any
	json.Unmarshal([]byte(`{"quantidade": 4}`), &dados)
	fmt.Printf("  o JSON {\"quantidade\": 4} virou %T no mapa\n", dados["quantidade"])
	func() {
		defer func() { fmt.Println("  pânico:", recover()) }()
		fmt.Println("  total com int:", total(map[string]any{"quantidade": 4}))
		fmt.Println("  total com o valor vindo do JSON:", total(dados))
	}()
}
